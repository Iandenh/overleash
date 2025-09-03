package overleash

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Iandenh/overleash/internal/storage"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/charmbracelet/log"
	"github.com/launchdarkly/eventsource"
)

var forceEnable = Strategy{
	Name:        "default",
	Constraints: make([]Constraint, 0),
	Parameters:  make(map[string]interface{}),
	Segments:    make([]int, 0),
	Variants:    make([]StrategyVariant, 0),
}

type OverleashContext struct {
	upstream            string
	featureEnvironments []*FeatureEnvironment
	activeFeatureIdx    int
	overrides           map[string]*Override
	LockMutex           sync.RWMutex
	lastSync            time.Time
	paused              bool
	ticker              ticker
	store               storage.Store
	client              client
	reload              time.Duration
	metrics             *metrics
	IsStreamer          bool
	FrontendApiEnabled  bool
}

type FeatureEnvironment struct {
	name              string
	environment       string
	token             string
	featureFile       FeatureFile
	cachedFeatureFile FeatureFile
	cachedJson        []byte
	etagOfCachedJson  string
	engine            unleashengine.Engine
	Streamer          *Streamer
}

func (o *OverleashContext) ActiveFeatureEnvironment() *FeatureEnvironment {
	return o.featureEnvironments[o.activeFeatureIdx]
}

func (o *OverleashContext) HasMultipleEnvironments() bool {
	return len(o.featureEnvironments) > 1
}

func (o *OverleashContext) FeatureEnvironments() []*FeatureEnvironment {
	return o.featureEnvironments
}

func (fe *FeatureEnvironment) EtagOfCachedJson() string {
	return fe.etagOfCachedJson
}

func (fe *FeatureEnvironment) Engine() unleashengine.Engine {
	return fe.engine
}

func (fe *FeatureEnvironment) Name() string {
	return fe.name
}

func (fe *FeatureEnvironment) Environment() string {
	return fe.environment
}

type OverrideConstraint struct {
	Enabled    bool       `json:"enabled"`
	Constraint Constraint `json:"constraint"`
}

type Override struct {
	FeatureFlag string               `json:"featureFlag"`
	Enabled     bool                 `json:"enabled"`
	IsGlobal    bool                 `json:"isGlobal"`
	Constraints []OverrideConstraint `json:"constraints"`
}

func NewOverleash(upstream string, tokens []string, reload time.Duration, streamer, frontendApiEnabled bool) *OverleashContext {
	o := &OverleashContext{
		upstream:            upstream,
		featureEnvironments: makeFeatureEnvironments(tokens, streamer, frontendApiEnabled),
		activeFeatureIdx:    0,
		overrides:           make(map[string]*Override),
		lastSync:            time.Now(),
		paused:              false,
		store:               storage.NewFileStore(),
		reload:              reload,
		IsStreamer:          streamer,
		FrontendApiEnabled:  frontendApiEnabled,
	}

	return o
}

func makeFeatureEnvironments(tokens []string, streamer, frontendApiEnabled bool) []*FeatureEnvironment {
	features := make([]*FeatureEnvironment, len(tokens))

	for i, token := range tokens {
		env, err := ExtractEnvironment(token)
		if err != nil {
			env = "default"
		}

		var s *Streamer
		var e unleashengine.Engine

		if streamer {
			s = NewStreamer()
		}

		if frontendApiEnabled {
			e = unleashengine.NewUnleashEngine()
		}

		features[i] = &FeatureEnvironment{
			name:        strings.SplitN(token, ".", 2)[0],
			token:       token,
			engine:      e,
			Streamer:    s,
			environment: env,
		}
	}

	return features
}

func (o *OverleashContext) Start(ctx context.Context, registerMetrics, register, useDeltaApi bool) {
	if o.client == nil {
		o.client = newClient(o.upstream, o.reload, ctx)
	}

	if overrides, err := o.readOverrides(); err == nil {
		o.overrides = overrides
	}

	if registerMetrics {
		o.startMetrics(ctx)
	}

	if register {
		o.registerRemotes()
	}

	err := o.loadRemotesWithLock()

	if err != nil {
		panic(err)
	}

	if o.reload == 0 {
		log.Info("Start without reloading")
		return
	}

	if useDeltaApi {
		o.startStreamListeners(ctx)
	} else {
		o.startFetcher(ctx)
	}
}

func (o *OverleashContext) startFetcher(ctx context.Context) {
	o.ticker = createTicker(o.reload)

	log.Infof("Start with reloading with %d seconds", int(o.reload.Seconds()))

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

func (o *OverleashContext) startStreamListeners(ctx context.Context) {
	o.ticker = createTicker(o.reload)

	log.Infof("Start with streaming")

	for idx, f := range o.FeatureEnvironments() {
		channel := make(chan eventsource.Event)

		o.client.streamFeatures(f.token, channel)

		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-channel:
					f.processSseEvent(event, o, idx == 0)
				}
			}
		}()
	}
}

func (o *OverleashContext) registerRemotes() {
	for _, featureEnvironment := range o.featureEnvironments {
		if token, ok := fromString(featureEnvironment.token); ok == true {
			o.client.registerClient(token)
		}
	}
}

func (o *OverleashContext) loadRemotesWithLock() error {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	return o.loadRemotes()
}

func (o *OverleashContext) loadRemotes() error {
	e := error(nil)

	for idx, featureEnvironment := range o.featureEnvironments {
		featureFile, err := o.client.getFeatures(featureEnvironment.token)

		if err != nil {
			log.Errorf("Error loading features: %s", err.Error())
			e = errors.Join(e, err)
			continue
		}

		o.featureEnvironments[idx].featureFile = *featureFile
	}

	o.compileFeatureFiles()
	o.lastSync = time.Now()

	return e
}

func (o *OverleashContext) RefreshFeatureFiles() error {
	err := o.loadRemotesWithLock()
	o.ticker.resetTicker()

	return err
}

func (o *OverleashContext) FeatureFileIdx() int {
	return o.activeFeatureIdx
}

func (o *OverleashContext) SetFeatureFileIdx(idx int) error {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	if idx < 0 || idx >= len(o.featureEnvironments) {
		return fmt.Errorf("invalid data file index: %d", idx)
	}

	o.activeFeatureIdx = idx

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

	o.compileFeatureFiles()
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

	o.compileFeatureFiles()
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) DeleteOverride(featureFlag string) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	delete(o.overrides, featureFlag)

	o.compileFeatureFiles()
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) DeleteAllOverride() {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.overrides = make(map[string]*Override)

	o.compileFeatureFiles()
	o.writeOverrides(o.overrides)
}

func (o *OverleashContext) SetPaused(paused bool) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	o.paused = paused

	go o.processOverleashStreaming()

	o.compileFeatureFiles()
}

func (o *OverleashContext) IsPaused() bool {
	return o.paused
}
func (fe *FeatureEnvironment) FeatureFile() FeatureFile {
	return fe.cachedFeatureFile
}

func (fe *FeatureEnvironment) RemoteFeatureFile() FeatureFile {
	return fe.featureFile
}

func (fe *FeatureEnvironment) Token() string {
	return fe.token
}

func (o *OverleashContext) GetRemotes() []string {
	remotes := make([]string, len(o.featureEnvironments))

	for idx, featureEnvironment := range o.featureEnvironments {
		remotes[idx] = featureEnvironment.name
	}

	return remotes
}

func (o *OverleashContext) Upstream() string {
	return o.upstream
}

func (fe *FeatureEnvironment) CachedJson() []byte {
	return fe.cachedJson
}

func (o *OverleashContext) Overrides() map[string]*Override {
	return o.overrides
}

func (o *OverleashContext) LastSync() time.Time {
	return o.lastSync
}

func (o *OverleashContext) compileFeatureFiles() {
	log.Debug("Compiling feature files")

	for _, featureEnvironment := range o.featureEnvironments {
		featureEnvironment.compile(o)
	}
}

func (fe *FeatureEnvironment) compile(o *OverleashContext) {
	df := fe.featureFileWithOverwrites(o)

	if fe.Streamer != nil {
		go fe.Streamer.processFeature(fe.cachedFeatureFile, df, fe.featureFile)
	}

	fe.cachedFeatureFile = df

	buf := new(bytes.Buffer)
	writer := json.NewEncoder(buf)

	err := writer.Encode(df)

	if err != nil {
		log.Error(err)
		panic(err)
	}

	fe.cachedJson = buf.Bytes()

	fe.etagOfCachedJson = calculateETag(fe.cachedJson)

	if fe.engine != nil {
		fe.engine.TakeState(string(fe.cachedJson))
	}
}

func (fe *FeatureEnvironment) featureFileWithOverwrites(o *OverleashContext) FeatureFile {
	featureFile := fe.RemoteFeatureFile()

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

func (o *OverleashContext) AddMetric(data *MetricsData) {
	if o.metrics == nil {
		return
	}

	o.metrics.metricChannel <- data
}

func (o *OverleashContext) AddRegistration(data *ClientData) {
	if o.metrics == nil {
		return
	}

	o.metrics.clientDataChannel <- data
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
	go o.processOverleashStreaming()

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
