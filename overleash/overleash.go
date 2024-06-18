package overleash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Unleash/unleash-client-go/v4/api"
	"overleash/cache"
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
	lockMutex         sync.Mutex
	lastSync          time.Time
	paused            bool
	ticker            ticker
}

type Override struct {
	FeatureFlag string
	Enabled     bool
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
	}

	overrides, err := ReadOverrides()
	if err == nil {
		o.overrides = overrides
	}

	o.compileDatafile()

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
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	e := error(nil)

	for idx, token := range o.tokens {
		featureFile, err := getFeatures(o.url, token)

		if err != nil {
			fmt.Println("Error loading features")
			e = errors.Join(e, err)
		}

		o.featureFiles[idx] = *featureFile
	}

	o.compileDatafile()
	o.lastSync = time.Now()

	return e
}

func (o *OverleashContext) RefreshDatafiles() error {
	err := o.loadRemotes()
	o.ticker.resetTicker()

	return err
}

func (o *OverleashContext) DataFileIdx() int {
	return o.featureIdx
}

func (o *OverleashContext) SetDataFileIdx(idx int) error {
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	if idx < 0 && idx >= len(o.featureFiles) {
		return fmt.Errorf("invalid data file index: %d", idx)
	}

	o.featureIdx = idx

	o.compileDatafile()

	return nil
}

func (o *OverleashContext) AddOverride(featureFlag string, enabled bool) {
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	o.overrides[featureFlag] = &Override{
		FeatureFlag: featureFlag,
		Enabled:     enabled,
	}

	o.compileDatafile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) DeleteOverride(featureFlag string) {
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	delete(o.overrides, featureFlag)

	o.compileDatafile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) DeleteAllOverride() {
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	o.overrides = make(map[string]*Override)

	o.compileDatafile()
	WriteOverrides(o.overrides)
}

func (o *OverleashContext) SetPaused(paused bool) {
	o.lockMutex.Lock()
	defer o.lockMutex.Unlock()

	o.paused = paused

	o.compileDatafile()
}

func (o *OverleashContext) Paused() bool {
	return o.paused
}
func (o *OverleashContext) FeatureFile() FeatureFile {
	return o.cachedFeatureFile
}

func (o *OverleashContext) RemoteDatafile() FeatureFile {
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

func (o *OverleashContext) compileDatafile() {
	df := o.featureFileWithOverwrites()

	o.cachedFeatureFile = df

	buf := new(bytes.Buffer)
	writer := json.NewEncoder(buf)

	err := writer.Encode(df)

	if err != nil {
		panic(err)
	}

	o.cachedJson = buf.Bytes()
}

func (o *OverleashContext) featureFileWithOverwrites() FeatureFile {
	featureFile := o.RemoteDatafile()

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
					featureFile.Features[idx].Strategies = []api.Strategy{forceEnable}
				} else {
					featureFile.Features[idx].Enabled = false
				}

				break
			}
		}
	}

	return featureFile
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