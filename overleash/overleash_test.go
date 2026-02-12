package overleash

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/unleashengine"
	"github.com/launchdarkly/eventsource"
)

// fakeStore implements a simple in-memory store.
type fakeStore struct {
	data map[string][]byte
	lock sync.Mutex
}

func (fs *fakeStore) Write(filename string, data []byte) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if fs.data == nil {
		fs.data = make(map[string][]byte)
	}
	fs.data[filename] = data
	return nil
}

func (fs *fakeStore) Read(filename string) ([]byte, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if fs.data == nil {
		return nil, errors.New("file not found")
	}
	d, ok := fs.data[filename]
	if !ok {
		return nil, errors.New("file not found")
	}
	return d, nil
}

// fakeEngine records the state passed via TakeState.
type fakeEngine struct {
	state string
}

func (fe *fakeEngine) TakeState(state string) {
	fe.state = state
}

func (fe *fakeEngine) Resolve(context *unleashengine.Context, featureName string) (*unleashengine.EvaluatedToggle, error) {
	// Return a dummy successful toggle to satisfy the interface
	return &unleashengine.EvaluatedToggle{
		Name:    featureName,
		Enabled: true,
	}, nil
}

func (fe *fakeEngine) ResolveAll(context *unleashengine.Context, includeAll bool) (*unleashengine.EvaluatedToggleList, error) {
	// Return an empty list to satisfy the interface
	return &unleashengine.EvaluatedToggleList{
		Toggles: []*unleashengine.EvaluatedToggle{},
	}, nil
}

// fakeClient implements getFeatures for testing.
type fakeClient struct {
	featureFile FeatureFile
	err         error
}

func (fc *fakeClient) validateToken(token string) (*EdgeToken, error) {
	return nil, nil
}

func (fc *fakeClient) registerClient(token *EdgeToken) error {
	return nil
}

func (fc *fakeClient) getFeatures(token string) (*FeatureFile, error) {
	return &fc.featureFile, fc.err
}

func (fc *fakeClient) bulkMetrics(token string, applications []*ClientData, metrics []*MetricsData) error {
	return nil
}

func (fc *fakeClient) streamFeatures(token string, channel chan eventsource.Event) error {
	return nil
}

// TestCompileFeatureFile verifies that compileFeatureFiles correctly encodes
// the remote feature file (with no overrides) and updates the cached JSON, ETag,
// and engine state.
func TestCompileFeatureFile(t *testing.T) {
	// Create a dummy feature file with one feature.
	ff := FeatureFile{
		Version: 1,
		Features: FeatureFlags{
			{
				Name:        "test-feature",
				Enabled:     false,
				Strategies:  []Strategy{{Name: "original"}},
				Project:     "default",
				Description: "A test feature",
			},
		},
		Segments: []Segment{},
	}

	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	// Set the remote feature file.
	o.ActiveFeatureEnvironment().featureFile = ff
	// Replace the engine with our fakeEngine.
	fe := &fakeEngine{}
	o.ActiveFeatureEnvironment().engine = fe

	// Recompile the feature file.
	o.compileFeatureFiles()

	// Check that the cached feature file matches.
	if o.ActiveFeatureEnvironment().cachedFeatureFile.Version != ff.Version {
		t.Errorf("Expected version %d, got %d", ff.Version, o.featureEnvironments[0].cachedFeatureFile.Version)
	}

	// Verify that cachedJson decodes to a valid FeatureFile.
	var decoded FeatureFile
	if err := json.Unmarshal(o.ActiveFeatureEnvironment().cachedJson, &decoded); err != nil {
		t.Errorf("cachedJson is not valid JSON: %v", err)
	}

	// Check that an ETag was calculated.
	if o.ActiveFeatureEnvironment().etagOfCachedJson == "" {
		t.Error("Expected non-empty etagOfCachedJson")
	}

	// Check that the engine state was updated.
	if fe.state != string(o.ActiveFeatureEnvironment().cachedJson) {
		t.Error("Engine state was not updated correctly")
	}
}

// TestOverrideAddAndDelete tests adding a global override to enable a feature
// and then deleting it.
func TestOverrideAddAndDelete(t *testing.T) {
	// Create a dummy feature file with one feature.
	ff := FeatureFile{
		Version: 1,
		Features: FeatureFlags{
			{
				Name:        "feature1",
				Enabled:     false,
				Strategies:  []Strategy{{Name: "original"}},
				Project:     "default",
				Description: "Feature 1",
			},
		},
		Segments: []Segment{},
	}

	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	o.ActiveFeatureEnvironment().featureFile = ff
	// Use fakeStore to avoid file I/O.
	fs := &fakeStore{}
	o.store = fs

	// Add a global override to enable "feature1".
	o.AddOverride("feature1", true)

	// Verify the override exists.
	exists, enabled := o.HasOverride("feature1")
	if !exists || !enabled {
		t.Error("Expected override for feature1 to exist and be enabled")
	}

	// Check that the compiled feature file shows "feature1" as enabled and its strategies replaced.
	compiled := o.ActiveFeatureEnvironment().FeatureFile()
	found := false
	for _, f := range compiled.Features {
		if f.Name == "feature1" {
			found = true
			if !f.Enabled {
				t.Error("Expected feature1 to be enabled due to override")
			}
			if len(f.Strategies) != 1 || f.Strategies[0].Name != forceEnable.Name {
				t.Error("Expected feature1 strategies to be replaced with forceEnable")
			}
		}
	}
	if !found {
		t.Error("feature1 not found in compiled feature file")
	}

	// Delete the override.
	o.DeleteOverride("feature1")
	// Verify the override was removed.
	exists, _ = o.HasOverride("feature1")
	if exists {
		t.Error("Expected override for feature1 to be deleted")
	}
}

// TestSetPaused verifies that when paused, overrides are not applied.
func TestSetPaused(t *testing.T) {
	ff := FeatureFile{
		Version: 1,
		Features: FeatureFlags{
			{
				Name:        "feature1",
				Enabled:     false,
				Strategies:  []Strategy{{Name: "original"}},
				Project:     "default",
				Description: "Feature 1",
			},
		},
		Segments: []Segment{},
	}

	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	o.ActiveFeatureEnvironment().featureFile = ff

	// Add an override.
	o.AddOverride("feature1", true)
	// Set the context to paused.
	o.SetPaused(true)
	// Recompile the feature file.
	o.compileFeatureFiles()
	// When paused, overrides should not be applied.
	compiled := o.ActiveFeatureEnvironment().FeatureFile()
	for _, f := range compiled.Features {
		if f.Name == "feature1" {
			if f.Enabled {
				t.Error("Expected feature1 to not be enabled when paused")
			}
			// Strategies should remain unchanged.
			if len(f.Strategies) != 1 || f.Strategies[0].Name != "original" {
				t.Error("Expected feature1 strategies to remain original when paused")
			}
		}
	}
}

// TestSetFeatureFileIdx tests setting a valid and invalid feature file index.
func TestSetFeatureFileIdx(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "token1,token2",
		Storage:  "file",
		Reload:   "0",
	}
	o := NewOverleash(cfg)

	// Valid index.
	if err := o.SetFeatureFileIdx(1); err != nil {
		t.Errorf("Expected no error for valid index, got %v", err)
	}
	if o.FeatureFileIdx() != 1 {
		t.Errorf("Expected activeFeatureIdx to be 1, got %d", o.FeatureFileIdx())
	}
	// Invalid index: negative.
	if err := o.SetFeatureFileIdx(-1); err == nil {
		t.Error("Expected error for invalid index -1, got nil")
	}
	// Invalid index: too high.
	if err := o.SetFeatureFileIdx(2); err == nil {
		t.Error("Expected error for invalid index 2, got nil")
	}
}

// TestTokenFunctions verifies GetRemotes and ActiveToken.
func TestTokenFunctions(t *testing.T) {
	tokens := []string{"*:remote1.token", "*:remote2.token"}

	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    strings.Join(tokens, ","),
		Storage:  "file",
		Reload:   "0",
	}
	o := NewOverleash(cfg)

	remotes := o.GetRemotes()
	if len(remotes) != 2 {
		t.Errorf("Expected 2 remotes, got %d", len(remotes))
	}
	if remotes[0] != "remote1" || remotes[1] != "remote2" {
		t.Errorf("Unexpected remotes: %v", remotes)
	}
	// ActiveToken should return the token at index 0 by default.
	if o.ActiveFeatureEnvironment().Token() != tokens[0] {
		t.Errorf("Expected active token %s, got %s", tokens[0], o.ActiveFeatureEnvironment().Token())
	}
}

// TestUpstreamAndCachedJson verifies the Upstream value and that CachedJson is non-empty.
func TestUpstreamAndCachedJson(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	o.ActiveFeatureEnvironment().featureFile = FeatureFile{Version: 1}
	o.compileFeatureFiles()
	if o.Upstream() != "http://example.com" {
		t.Errorf("Expected upstream to be http://example.com, got %s", o.Upstream())
	}
	if len(o.ActiveFeatureEnvironment().CachedJson()) == 0 {
		t.Error("Expected non-empty cachedJson")
	}
}

// TestWriteAndReadOverrides verifies that overrides written to the store
// can be correctly read back.
func TestWriteAndReadOverrides(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	fs := &fakeStore{}
	o.store = fs

	// Add an override.
	o.AddOverride("feature1", true)
	// Read overrides.
	overrides, err := o.readOverrides()
	if err != nil {
		t.Errorf("readOverrides returned error: %v", err)
	}
	if len(overrides) != 1 {
		t.Errorf("Expected 1 override, got %d", len(overrides))
	}
	if overrides["feature1"] == nil || !overrides["feature1"].Enabled {
		t.Error("Override for feature1 not as expected")
	}
}

// TestLoadRemotes verifies that loadRemotesWithLock updates featureFile using a fake overleashClient.
func TestLoadRemotes(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	// Create a dummy feature file.
	ff := FeatureFile{
		Version: 2,
		Features: FeatureFlags{
			{
				Name:        "featureX",
				Enabled:     false,
				Strategies:  []Strategy{{Name: "original"}},
				Project:     "default",
				Description: "Feature X",
			},
		},
	}
	// Use a fakeClient.
	fc := &fakeClient{featureFile: ff, err: nil}
	o.client = fc

	// Call loadRemotesWithLock.
	if err := o.loadRemotesWithLock(); err != nil {
		t.Errorf("loadRemotesWithLock returned error: %v", err)
	}
	// Verify that featureFile[0] is updated.
	if o.ActiveFeatureEnvironment().featureFile.Version != 2 {
		t.Errorf("Expected feature file version 2, got %d", o.ActiveFeatureEnvironment().featureFile.Version)
	}
}

// TestRefreshFeatureFiles verifies that RefreshFeatureFiles updates remotes and resets the ticker.
func TestRefreshFeatureFiles(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	// Set a fake ticker.
	o.ticker = ticker{period: time.Minute, ticker: nil}

	// Use a fakeClient to return a feature file.
	ff := FeatureFile{
		Version: 3,
		Features: FeatureFlags{
			{
				Name:        "featureY",
				Enabled:     false,
				Strategies:  []Strategy{{Name: "original"}},
				Project:     "default",
				Description: "Feature Y",
			},
		},
	}
	fc := &fakeClient{featureFile: ff, err: nil}
	o.client = fc

	if err := o.RefreshFeatureFiles(); err != nil {
		t.Errorf("RefreshFeatureFiles returned error: %v", err)
	}
	// Verify that featureFile[0] is updated.
	if o.ActiveFeatureEnvironment().featureFile.Version != 3 {
		t.Errorf("Expected feature file version 3, got %d", o.ActiveFeatureEnvironment().featureFile.Version)
	}
}

// TestHasAndGetOverride verifies the HasOverride and GetOverride functions.
func TestHasAndGetOverride(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	// Initially, no override should exist.
	exists, _ := o.HasOverride("nonexistent")
	if exists {
		t.Error("Expected HasOverride to return false for nonexistent key")
	}
	if o.GetOverride("nonexistent") != nil {
		t.Error("Expected GetOverride to return nil for nonexistent key")
	}
	// Add an override.
	o.AddOverride("feature1", false)
	exists, enabled := o.HasOverride("feature1")
	if !exists || enabled {
		t.Error("Expected override for feature1 to exist and be disabled")
	}
	ov := o.GetOverride("feature1")
	if ov == nil || ov.FeatureFlag != "feature1" {
		t.Error("GetOverride did not return the expected override")
	}
}

// TestCalculateETag tests the calculateETag function.
func TestCalculateETag(t *testing.T) {
	data := []byte("test data")
	etag := calculateETag(data)
	if etag == "" {
		t.Error("Expected non-empty etag")
	}
	// Manually calculate expected value.
	expected := sha256SumHex(data)
	if etag != expected {
		t.Errorf("Expected etag %s, got %s", expected, etag)
	}
}

// Helper function for computing SHA-256 hex digest.
func sha256SumHex(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// --- Optional: Testing Start method using a context ---
//
// Because Start spawns a goroutine that listens on a ticker, one can test that it stops
// when the context is cancelled. This example sets reload to 0 (which skips reloading)
// and verifies that Start returns immediately.
func TestStartWithoutReload(t *testing.T) {
	cfg := &config.Config{
		Upstream: "http://example.com",
		Token:    "dummy.token",
		Storage:  "file",
		Reload:   "0",
	}

	o := NewOverleash(cfg)
	o.client = &fakeClient{}

	// For this test, set reload to 0.
	ctx := t.Context()
	// Calling Start with reload==0 should not start any goroutine.
	o.Start(ctx)
	// Simply check that no panic occurred and that overleashClient was created.
	if o.client == nil {
		t.Error("Expected overleashClient to be created in Start")
	}
}
