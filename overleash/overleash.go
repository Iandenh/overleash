package overleash

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Iandenh/overleash/internal/storage"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/charmbracelet/log"
	"strings"
	"sync"
	"time"
)

var forceEnable = Strategy{
	Name:        "default",
	Constraints: make([]Constraint, 0),
	Parameters:  make(map[string]interface{}),
	Segments:    make([]int, 0),
	Variants:    make([]StrategyVariant, 0),
}

type OverleashContext struct {
	upstream          string
	tokens            []string
	featureFiles      []FeatureFile
	featureIdx        int
	cachedFeatureFile FeatureFile
	cachedJson        []byte
	etagOfCachedJson  string
	overrides         map[string]*Override
	LockMutex         sync.RWMutex
	lastSync          time.Time
	paused            bool
	ticker            ticker
	engine            unleashengine.Engine
	store             storage.Store
	client            client
	reload            int
}

func (o *OverleashContext) EtagOfCachedJson() string {
	return o.etagOfCachedJson
}

func (o *OverleashContext) Engine() unleashengine.Engine {
	return o.engine
}

type OverrideConstraint struct {
	Enabled    bool
	Constraint Constraint
}

type Override struct {
	FeatureFlag string
	Enabled     bool
	IsGlobal    bool
	Constraints []OverrideConstraint
}

func NewOverleash(upstream string, tokens []string, reload int) *OverleashContext {
	o := &OverleashContext{
		upstream:     upstream,
		tokens:       tokens,
		featureFiles: make([]FeatureFile, len(tokens)),
		featureIdx:   0,
		overrides:    make(map[string]*Override),
		lastSync:     time.Now(),
		paused:       false,
		engine:       unleashengine.NewUnleashEngine(),
		store:        storage.NewFileStore(),
		client:       newClient(upstream, reload),
		reload:       reload,
	}

	return o
}

func (o *OverleashContext) Start(ctx context.Context) {
	if overrides, err := o.readOverrides(); err == nil {
		o.overrides = overrides
	}

	err := o.loadRemotesWithLock()

	if err != nil {
		panic(err)
	}

	if o.reload == 0 {
		log.Info("Start without reloading")
		return
	}

	o.ticker = createTicker(time.Duration(o.reload) * time.Minute)

	log.Infof("Start with reloading with %d", o.reload)

	go func() {
		defer o.ticker.ticker.Stop()
		log.Info("Reloading remotes")

		for {
			select {
			case <-ctx.Done():
				return
			case <-o.ticker.ticker.C:
				o.loadRemotesWithLock()
			}
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
		featureFile, err := o.client.getFeatures(token)

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

	if idx < 0 || idx >= len(o.featureFiles) {
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
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) AddOverrideConstraint(featureFlag string, enabled bool, constraint Constraint) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	if o.overrides[featureFlag] == nil || o.overrides[featureFlag].IsGlobal {
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
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) DeleteOverride(featureFlag string) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	delete(o.overrides, featureFlag)

	o.compileFeatureFile()
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) DeleteAllOverride() {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.overrides = make(map[string]*Override)

	o.compileFeatureFile()
	o.writeOverrides(o.overrides)
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

func (o *OverleashContext) Upstream() string {
	return o.upstream
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

	o.etagOfCachedJson = calculateETag(o.cachedJson)

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
					featureFile.Features[idx].Strategies = mapOverrideToStrategies(override, featureFile.Features[idx])
					featureFile.Features[idx].Enabled = true
				} else {
					featureFile.Features[idx].Enabled = false
				}

				break
			}
		}
	}

	return featureFile
}

func mapOverrideToStrategies(override *Override, feature Feature) []Strategy {
	if override.IsGlobal {
		return []Strategy{forceEnable}
	}

	var strategies []Strategy

	if feature.Enabled {
		strategies = make([]Strategy, len(feature.Strategies))
		copy(strategies, feature.Strategies)
	} else {
		strategies = []Strategy{}
	}

	var enabledConstraints []Constraint
	var disabledConstraints []Constraint

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
			strategies = append(strategies, Strategy{
				Name: "flexibleRollout",
				Parameters: map[string]interface{}{
					"groupId":    override.FeatureFlag,
					"rollout":    "100",
					"stickiness": "default",
				},
				Constraints: []Constraint{constraint},
				Segments:    nil,
				Variants:    make([]StrategyVariant, 0),
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

func (o *OverleashContext) writeOverrides(overrides map[string]*Override) error {
	data, err := json.Marshal(overrides)

	if err != nil {
		return err
	}

	err = o.store.Write("overrides.json", data)

	if err != nil {
		log.Debug(err.Error())
	}

	return err
}

func (o *OverleashContext) readOverrides() (map[string]*Override, error) {
	overrides := &map[string]*Override{}

	data, err := o.store.Read("overrides.json")

	if err != nil {
		return *overrides, err
	}

	err = json.Unmarshal(data, overrides)

	if err != nil {
		return *overrides, err
	}

	return *overrides, nil
}

func calculateETag(bytes []byte) string {
	hasher := sha256.New()
	hasher.Write(bytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
