package overleash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4/api"
	"overleash/cache"
	"overleash/unleashengine"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var forceEnable = api.Strategy{
	Id:          123,
	Name:        "default",
	Constraints: make([]api.Constraint, 0),
	Parameters:  make(map[string]interface{}),
	Segments:    make([]int, 0),
	Variants:    make([]api.VariantInternal, 0),
}

type OverleashContext struct {
	url               string
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
	Constraint api.Constraint
}

type Override struct {
	FeatureFlag string
	Enabled     bool
	IsGlobal    bool
	Constraints []OverrideConstraint
}

func NewOverleash(url string, tokens []string) *OverleashContext {
	o := &OverleashContext{
		url:          url,
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
	err := o.loadRemotes()
	o.ticker = createTicker(time.Duration(reload) * time.Minute)

	if err != nil {
		panic(err)
	}

	if reload == 0 {
		return
	}

	fmt.Printf("Start with reloading with %d\n", reload)

	go func() {
		defer o.ticker.ticker.Stop()

		for range o.ticker.ticker.C {
			o.loadRemotes()
		}
	}()
}

func (o *OverleashContext) loadRemotes() error {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	e := error(nil)

	for idx, token := range o.tokens {
		featureFile, err := getFeatures(o.url, token)

		if err != nil {
			fmt.Println("Error loading features")
			e = errors.Join(e, err)
		}

		o.featureFiles[idx] = *featureFile
	}

	o.compileFeatureFile()
	o.lastSync = time.Now()

	return e
}

func (o *OverleashContext) RefreshFeatureFiles() error {
	err := o.loadRemotes()
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

func (o *OverleashContext) AddOverrideConstraint(featureFlag string, enabled bool, constraint api.Constraint) {
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

func (o *OverleashContext) Paused() bool {
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
	df := o.featureFileWithOverwrites()

	o.cachedFeatureFile = df

	buf := new(bytes.Buffer)
	writer := json.NewEncoder(buf)

	err := writer.Encode(df)

	if err != nil {
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

func mapOverrideToStrategies(override *Override, currentStrategies []api.Strategy) []api.Strategy {
	if override.IsGlobal {
		return []api.Strategy{forceEnable}
	}

	strategies := make([]api.Strategy, len(currentStrategies))
	copy(strategies, currentStrategies)

	var enabledConstraints []api.Constraint
	var disabledConstraints []api.Constraint

	for _, constraint := range override.Constraints {
		if constraint.Enabled {
			enabledConstraints = append(enabledConstraints, constraint.Constraint)
		} else {
			constraint.Constraint.Inverted = !constraint.Constraint.Inverted
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
			strategies = append(strategies, api.Strategy{
				Id:   0,
				Name: "flexibleRollout",
				Parameters: map[string]interface{}{
					"groupId":    override.FeatureFlag,
					"rollout":    "100",
					"stickiness": "default",
				},
				Constraints: []api.Constraint{constraint},
				Segments:    nil,
				Variants:    make([]api.VariantInternal, 0),
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
