package overleash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Iandenh/overleash/cache"
	"github.com/Iandenh/overleash/unleashengine"
	unleash "github.com/Unleash/unleash-client-go/v4/api"
	"github.com/charmbracelet/log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var forceEnable = unleash.Strategy{
	Id:          123,
	Name:        "default",
	Constraints: make([]unleash.Constraint, 0),
	Parameters:  make(map[string]interface{}),
	Segments:    make([]int, 0),
	Variants:    make([]unleash.VariantInternal, 0),
}

type OverleashContext struct {
	url               string
	dynamicMode       bool
	dynamicToken      *EdgeToken
	tokens            []string
	featureFiles      []FeatureFile
	featureIdx        int
	cachedFeatureFile FeatureFile
	cachedJson        []byte
	cachedBuf         bytes.Buffer
	overrides         map[string]*Override
	LockMutex         sync.RWMutex
	lastSync          time.Time
	paused            bool
	ticker            ticker
	engine            *unleashengine.UnleashEngine
}

func (o *OverleashContext) Engine() *unleashengine.UnleashEngine {
	return o.engine
}

type OverrideConstraint struct {
	Enabled    bool
	Constraint unleash.Constraint
}

type Override struct {
	FeatureFlag string
	Enabled     bool
	IsGlobal    bool
	Constraints []OverrideConstraint
}

func NewOverleash(url string, tokens []string, dynamicMode bool) *OverleashContext {
	o := &OverleashContext{
		url:          url,
		dynamicMode:  dynamicMode,
		dynamicToken: nil,
		tokens:       tokens,
		featureFiles: make([]FeatureFile, len(tokens)),
		featureIdx:   0,
		overrides:    make(map[string]*Override),
		lastSync:     time.Now(),
		paused:       false,
		engine:       unleashengine.NewUnleashEngine(),
	}

	overrides, err := ReadOverrides()
	if err == nil {
		o.overrides = overrides
	}

	o.compileFeatureFile()

	return o
}

func (o *OverleashContext) Start(reload int) {
	err := o.loadRemotesWithLock()

	if err != nil {
		panic(err)
	}

	if reload == 0 {
		log.Info("Start without reloading")
		return
	}

	o.ticker = createTicker(time.Duration(reload) * time.Minute)

	log.Infof("Start with reloading with %d", reload)

	go func() {
		defer o.ticker.ticker.Stop()
		log.Info("Reloading remotes")

		for range o.ticker.ticker.C {
			o.loadRemotesWithLock()
		}
	}()
}

func (o *OverleashContext) loadRemotesWithLock() error {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	return o.loadRemotes()
}

func (o *OverleashContext) loadRemotes() error {
	e := error(nil)

	for idx, token := range o.tokens {
		featureFile, err := getFeatures(o.url, token)

		if err != nil {
			log.Errorf("Error loading features: %s", err.Error())
			e = errors.Join(e, err)
			continue
		}

		o.featureFiles[idx] = *featureFile
	}

	o.compileFeatureFile()
	o.lastSync = time.Now()

	return e
}

func (o *OverleashContext) RefreshFeatureFiles() error {
	err := o.loadRemotesWithLock()
	o.ticker.resetTicker()

	return err
}

func (o *OverleashContext) FeatureFileIdx() int {
	return o.featureIdx
}

func (o *OverleashContext) SetFeatureFileIdx(idx int) error {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	if idx < 0 && idx >= len(o.featureFiles) {
		return fmt.Errorf("invalid data file index: %d", idx)
	}

	o.featureIdx = idx

	o.compileFeatureFile()

	return nil
}

func (o *OverleashContext) AddOverride(featureFlag string, enabled bool) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.overrides[featureFlag] = &Override{
		FeatureFlag: featureFlag,
		Enabled:     enabled,
		IsGlobal:    true,
	}

	o.compileFeatureFile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) AddOverrideConstraint(featureFlag string, enabled bool, constraint unleash.Constraint) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	if o.overrides[featureFlag] == nil || o.overrides[featureFlag].IsGlobal == true {
		o.overrides[featureFlag] = &Override{
			FeatureFlag: featureFlag,
			Enabled:     true,
			IsGlobal:    false,
			Constraints: make([]OverrideConstraint, 0),
		}
	}

	o.overrides[featureFlag].Constraints = append(o.overrides[featureFlag].Constraints, OverrideConstraint{
		Enabled:    enabled,
		Constraint: constraint,
	})

	o.compileFeatureFile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) DeleteOverride(featureFlag string) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	delete(o.overrides, featureFlag)

	o.compileFeatureFile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) DeleteAllOverride() {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.overrides = make(map[string]*Override)

	o.compileFeatureFile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) SetPaused(paused bool) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.paused = paused

	o.compileFeatureFile()
}

func (o *OverleashContext) IsPaused() bool {
	return o.paused
}
func (o *OverleashContext) FeatureFile() FeatureFile {
	return o.cachedFeatureFile
}

func (o *OverleashContext) RemoteFeatureFile() FeatureFile {
	return o.featureFiles[o.featureIdx]
}

func (o *OverleashContext) ActiveToken() string {
	return o.tokens[o.featureIdx]
}

func (o *OverleashContext) GetRemotes() []string {
	remotes := make([]string, len(o.tokens))

	for idx, token := range o.tokens {
		parts := strings.Split(token, ".")
		remotes[idx] = parts[0]
	}

	return remotes
}

func (o *OverleashContext) IsDynamicMode() bool {
	return o.dynamicMode
}

func (o *OverleashContext) Url() string {
	return o.url
}

func (o *OverleashContext) ShouldDoDynamicCheck() bool {
	if o.dynamicMode == false {
		return false
	}

	return o.dynamicToken == nil
}

func (o *OverleashContext) AddDynamicToken(token string) bool {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	edgeToken, err := validateToken(o.url, token)

	if err != nil {
		return false
	}

	if edgeToken.TokenType != Client {
		return false
	}

	o.dynamicToken = edgeToken
	o.tokens = []string{edgeToken.Token}

	o.loadRemotes()

	return true
}

func (o *OverleashContext) CachedJson() []byte {
	return o.cachedJson
}

func (o *OverleashContext) Overrides() map[string]*Override {
	return o.overrides
}

func (o *OverleashContext) LastSync() time.Time {
	return o.lastSync
}

func (o *OverleashContext) compileFeatureFile() {
	log.Debug("Compiling feature file")
	df := o.featureFileWithOverwrites()

	o.cachedFeatureFile = df

	buf := new(bytes.Buffer)
	writer := json.NewEncoder(buf)

	err := writer.Encode(df)

	if err != nil {
		log.Error(err)
		panic(err)
	}

	o.cachedJson = buf.Bytes()

	o.engine.TakeState(string(o.cachedJson))
}

func (o *OverleashContext) featureFileWithOverwrites() FeatureFile {
	featureFile := o.RemoteFeatureFile()

	f := make(FeatureFlags, len(featureFile.Features))
	copy(f, featureFile.Features)
	featureFile.Features = f

	if o.paused {
		return featureFile
	}

	for _, override := range o.overrides {
		for idx, flag := range featureFile.Features {
			if flag.Name == override.FeatureFlag {
				if override.Enabled {
					featureFile.Features[idx].Enabled = true
					featureFile.Features[idx].Strategies = mapOverrideToStrategies(override, featureFile.Features[idx].Strategies)
				} else {
					featureFile.Features[idx].Enabled = false
				}

				break
			}
		}
	}

	return featureFile
}

func mapOverrideToStrategies(override *Override, currentStrategies []unleash.Strategy) []unleash.Strategy {
	if override.IsGlobal {
		return []unleash.Strategy{forceEnable}
	}

	strategies := make([]unleash.Strategy, len(currentStrategies))
	copy(strategies, currentStrategies)

	var enabledConstraints []unleash.Constraint
	var disabledConstraints []unleash.Constraint

	for _, constraint := range override.Constraints {
		if constraint.Enabled {
			enabledConstraints = append(enabledConstraints, constraint.Constraint)
		} else {
			if constraint.Constraint.Operator == unleash.OperatorIn {
				constraint.Constraint.Operator = unleash.OperatorNotIn
			} else {
				constraint.Constraint.Inverted = !constraint.Constraint.Inverted
			}

			disabledConstraints = append(disabledConstraints, constraint.Constraint)
		}
	}

	if len(disabledConstraints) > 0 {
		for idx, strategy := range strategies {
			strategies[idx].Constraints = append(strategy.Constraints, disabledConstraints...)
		}
	}

	if len(enabledConstraints) > 0 {
		for _, constraint := range enabledConstraints {
			strategies = append(strategies, unleash.Strategy{
				Id:   0,
				Name: "flexibleRollout",
				Parameters: map[string]interface{}{
					"groupId":    override.FeatureFlag,
					"rollout":    "100",
					"stickiness": "default",
				},
				Constraints: []unleash.Constraint{constraint},
				Segments:    nil,
				Variants:    make([]unleash.VariantInternal, 0),
			})
		}
	}

	return strategies
}

func (o *OverleashContext) HasOverride(key string) (bool, bool) {
	override, ok := o.overrides[key]

	if !ok {
		return false, false
	}

	return true, override.Enabled
}

func (o *OverleashContext) GetOverride(key string) *Override {
	override, ok := o.overrides[key]

	if !ok {
		return nil
	}

	return override
}

func WriteOverrides(overrides map[string]*Override) error {
	data, err := json.Marshal(overrides)

	if err != nil {
		return err
	}

	return cache.WriteFile(filepath.Join(cache.DataDir(), "overrides.json"), data)
}

func ReadOverrides() (map[string]*Override, error) {
	overrides := &map[string]*Override{}

	data, err := cache.ReadFile(filepath.Join(cache.DataDir(), "overrides.json"))

	if err != nil {
		return *overrides, err
	}

	err = json.Unmarshal(data, overrides)

	if err != nil {
		return *overrides, err
	}

	return *overrides, nil
}
