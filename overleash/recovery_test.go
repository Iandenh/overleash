package overleash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Iandenh/overleash/config"
	"github.com/charmbracelet/log"
)

type recoveryStore struct {
	fakeStore
	readBackup  func(string) ([]byte, error)
	writeBackup func(string, []byte) error
}

func (s *recoveryStore) Read(key string) ([]byte, error) {
	if strings.HasSuffix(key, "-backup.json") && s.readBackup != nil {
		return s.readBackup(key)
	}
	return s.fakeStore.Read(key)
}

func (s *recoveryStore) Write(key string, data []byte) error {
	if strings.HasSuffix(key, "-backup.json") && s.writeBackup != nil {
		return s.writeBackup(key, data)
	}
	return s.fakeStore.Write(key, data)
}

func recoveryContext(tokens string, backup bool) (*OverleashContext, *recoveryStore) {
	o := NewOverleash(&config.Config{Token: tokens, Storage: "null", Reload: "0", Backup: backup})
	store := &recoveryStore{}
	o.store = store
	return o, store
}

func startForRecovery(t *testing.T, o *OverleashContext) {
	t.Helper()
	defer func() {
		if failure := recover(); failure != nil {
			t.Fatalf("startup failed despite usable state: %v", failure)
		}
	}()
	o.Start(t.Context())
}

func TestStartupRecoversOnlyFailedEnvironment(t *testing.T) {
	for _, freshBackup := range []string{"", `{"features":[{"name":"stale"}]}`, "not json"} {
		t.Run(fmt.Sprintf("fresh_backup_%q", freshBackup), func(t *testing.T) {
			o, store := recoveryContext("*:dev.test,*:prod.test", true)
			o.client = &partialUpstreamClient{}
			store.data = map[string][]byte{"*:prod-backup.json": []byte(`{"features":[{"name":"recovered"}]}`)}
			if freshBackup != "" {
				store.data["*:dev-backup.json"] = []byte(freshBackup)
			}
			store.writeBackup = func(string, []byte) error { return errors.New("backup write failed") }
			store.readBackup = func(key string) ([]byte, error) {
				if key != "*:prod-backup.json" {
					t.Errorf("read backup for successful environment: %s", key)
				}
				return store.fakeStore.Read(key)
			}

			startForRecovery(t, o)

			environments := o.FeatureEnvironments()
			if !environments[0].Snapshot().Get("fresh").Exists() || environments[0].Snapshot().Get("stale").Exists() {
				t.Error("successful environment lost its fresh upstream state")
			}
			if !environments[1].Snapshot().Get("recovered").Exists() {
				t.Error("failed environment was not recovered from its own backup")
			}
		})
	}
}

func TestStartupRetainsUsableState(t *testing.T) {
	for _, backup := range []bool{false, true} {
		for _, empty := range []bool{false, true} {
			t.Run(fmt.Sprintf("backup_%t/empty_%t", backup, empty), func(t *testing.T) {
				o, store := recoveryContext("*:dev.test", backup)
				state := FeatureFile{Version: 2}
				if !empty {
					state.Features = FeatureFlags{{Name: "retained"}}
				}
				o.LoadFeatureFile(state)
				o.client = &fakeClient{err: errors.New("upstream unavailable")}
				store.readBackup = func(key string) ([]byte, error) {
					t.Errorf("read backup despite usable in-memory state: %s", key)
					return []byte(`{"features":[{"name":"stale"}]}`), nil
				}

				startForRecovery(t, o)

				if got := o.ActiveFeatureEnvironment().RemoteFeatureFile(); !reflect.DeepEqual(got, state) {
					t.Fatalf("retained state changed: got %+v, want %+v", got, state)
				}
			})
		}
	}
}

func TestStartupFailsWithoutUsableState(t *testing.T) {
	for _, test := range []struct {
		name   string
		backup bool
		data   string
	}{
		{name: "backup disabled", data: `{"features":[]}`},
		{name: "missing backup", backup: true},
		{name: "corrupt backup", backup: true, data: "not json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			o, store := recoveryContext("*:dev.test,*:prod.test", test.backup)
			o.client = &partialUpstreamClient{}
			if test.data != "" {
				store.data = map[string][]byte{"*:prod-backup.json": []byte(test.data)}
			}
			var failure any
			func() {
				defer func() { failure = recover() }()
				o.Start(t.Context())
			}()
			if failure == nil {
				t.Fatal("startup succeeded with an uninitialized environment")
			}
			if !strings.Contains(fmt.Sprint(failure), "*:prod") {
				t.Fatalf("startup failure does not identify the unavailable environment: %v", failure)
			}
		})
	}
}

func TestStartupAcceptsEmptyFeatureState(t *testing.T) {
	for _, fromBackup := range []bool{false, true} {
		t.Run(fmt.Sprintf("from_backup_%t", fromBackup), func(t *testing.T) {
			o, store := recoveryContext("*:dev.test", true)
			o.client = &fakeClient{featureFile: FeatureFile{Version: 2}}
			if fromBackup {
				o.client = &fakeClient{err: errors.New("upstream unavailable")}
				store.data = map[string][]byte{"*:dev-backup.json": []byte(`{"version":2,"features":[],"segments":[]}`)}
			} else {
				store.readBackup = func(key string) ([]byte, error) {
					t.Errorf("read backup despite successful empty upstream state: %s", key)
					return nil, errors.New("unexpected backup read")
				}
			}

			startForRecovery(t, o)

			got := o.ActiveFeatureEnvironment().RemoteFeatureFile()
			if got.Version != 2 || len(got.Features) != 0 {
				t.Fatalf("empty state was not loaded: %+v", got)
			}
			if len(o.ActiveFeatureEnvironment().CachedJson()) == 0 {
				t.Fatal("empty state was not published")
			}
		})
	}
}

func TestRefreshRetainsStateInsteadOfReadingBackup(t *testing.T) {
	o, store := recoveryContext("*:dev.test", true)
	o.LoadFeatureFile(FeatureFile{Version: 2})
	o.client = &fakeClient{err: errors.New("upstream unavailable")}
	store.readBackup = func(key string) ([]byte, error) {
		t.Errorf("read backup despite usable in-memory state: %s", key)
		return []byte(`{"features":[{"name":"stale"}]}`), nil
	}
	if err := o.RefreshFeatureFiles(); err == nil {
		t.Fatal("refresh did not report the upstream failure")
	}
	if got := o.ActiveFeatureEnvironment().RemoteFeatureFile(); got.Version != 2 || len(got.Features) != 0 {
		t.Fatalf("refresh replaced retained empty state: %+v", got)
	}
}

func TestRefreshRecoversUninitializedEnvironment(t *testing.T) {
	o, store := recoveryContext("*:dev.test", true)
	upstreamErr := errors.New("upstream unavailable")
	o.client = &fakeClient{err: upstreamErr}
	store.data = map[string][]byte{"*:dev-backup.json": []byte(`{"features":[{"name":"recovered"}]}`)}
	store.writeBackup = func(key string, _ []byte) error {
		t.Errorf("recovered state was unnecessarily rewritten as a fresh backup: %s", key)
		return nil
	}
	if err := o.RefreshFeatureFiles(); !errors.Is(err, upstreamErr) {
		t.Fatalf("refresh did not preserve the upstream error: %v", err)
	}
	if !o.ActiveFeatureEnvironment().Snapshot().Get("recovered").Exists() {
		t.Fatal("refresh did not recover the uninitialized environment")
	}
}

func TestBackupWriteFailureDoesNotRejectFreshState(t *testing.T) {
	for _, startup := range []bool{false, true} {
		t.Run(fmt.Sprintf("startup_%t", startup), func(t *testing.T) {
			var output bytes.Buffer
			previous := log.Default()
			log.SetDefault(log.New(&output))
			t.Cleanup(func() { log.SetDefault(previous) })
			o, store := recoveryContext("*:dev.test", true)
			o.client = &fakeClient{featureFile: FeatureFile{Features: FeatureFlags{{Name: "fresh"}}}}
			store.writeBackup = func(string, []byte) error { return errors.New("disk is full") }

			if startup {
				startForRecovery(t, o)
			} else if err := o.RefreshFeatureFiles(); err != nil {
				t.Fatalf("backup write failure rejected successful upstream retrieval: %v", err)
			}

			if !o.ActiveFeatureEnvironment().Snapshot().Get("fresh").Exists() {
				t.Fatal("fresh state was not published")
			}
			if text := output.String(); !strings.Contains(text, "disk is full") || !strings.Contains(text, "*:dev") {
				t.Fatalf("backup failure was not reported with environment context: %s", text)
			}
		})
	}
}

func TestRecoveryBackupsExcludeOverrides(t *testing.T) {
	o, store := recoveryContext("*:dev.test", true)
	o.AddOverride("flag", true)
	o.client = &fakeClient{featureFile: FeatureFile{Features: FeatureFlags{{Name: "flag", Enabled: false}}}}
	startForRecovery(t, o)

	if feature := o.ActiveFeatureEnvironment().FeatureFile().Get("flag"); feature == nil || !feature.Enabled {
		t.Fatal("local override was not applied to served state")
	}
	data, err := store.Read("*:dev-backup.json")
	if err != nil {
		t.Fatal(err)
	}
	var backup FeatureFile
	if err := json.Unmarshal(data, &backup); err != nil {
		t.Fatal(err)
	}
	if feature := backup.Get("flag"); feature == nil || feature.Enabled {
		t.Fatal("backup contains local overrides instead of upstream state")
	}
}

func TestBackupRecoveryDoesNotBlockStateAndPreservesNewerImports(t *testing.T) {
	for _, importState := range []bool{false, true} {
		t.Run(fmt.Sprintf("import_%t", importState), func(t *testing.T) {
			o, store := recoveryContext("*:dev.test", true)
			o.client = &fakeClient{err: errors.New("upstream unavailable")}
			entered := make(chan struct{})
			release := make(chan struct{})
			store.readBackup = func(string) ([]byte, error) {
				close(entered)
				<-release
				return []byte(`{"features":[{"name":"backup"}]}`), nil
			}
			completed := make(chan any, 1)
			go func() {
				defer func() { completed <- recover() }()
				o.Start(t.Context())
			}()
			<-entered
			updated := make(chan struct{})
			go func() {
				o.IsPaused()
				if importState {
					o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "newer"}}})
				}
				close(updated)
			}()
			blocked := false
			select {
			case <-updated:
			case <-time.After(200 * time.Millisecond):
				blocked = true
			}
			close(release)
			if failure := <-completed; failure != nil {
				t.Fatalf("startup failed: %v", failure)
			}
			<-updated
			if blocked {
				t.Fatal("backup recovery held the state lock")
			}
			want := "backup"
			if importState {
				want = "newer"
			}
			if !o.ActiveFeatureEnvironment().Snapshot().Get(want).Exists() {
				t.Fatalf("expected state %q after recovery", want)
			}
			if importState && o.ActiveFeatureEnvironment().Snapshot().Get("backup").Exists() {
				t.Fatal("late backup read overwrote the newer import")
			}
		})
	}
}

func TestDiscardedUpstreamResponseDoesNotOverwriteBackup(t *testing.T) {
	o, store := recoveryContext("*:dev.test", true)
	client := &blockedUpstreamClient{
		entered: make(chan struct{}), release: make(chan struct{}),
		fakeClient: fakeClient{featureFile: FeatureFile{Features: FeatureFlags{{Name: "stale"}}}},
	}
	o.client = client
	store.writeBackup = func(key string, _ []byte) error {
		t.Errorf("discarded upstream state was written to backup: %s", key)
		return nil
	}
	completed := make(chan error, 1)
	go func() { completed <- o.RefreshFeatureFiles() }()
	<-client.entered
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "newer"}}})
	close(client.release)
	if err := <-completed; err != nil {
		t.Fatal(err)
	}
	if !o.ActiveFeatureEnvironment().Snapshot().Get("newer").Exists() {
		t.Fatal("late upstream response overwrote the newer import")
	}
}

type recoveryClient struct {
	fakeClient
	fetch func(string) (*FeatureFile, error)
}

func (c *recoveryClient) getFeatures(token string) (*FeatureFile, error) {
	return c.fetch(token)
}

func TestBackupWritesStaySerialized(t *testing.T) {
	o, store := recoveryContext("*:dev.test", true)
	var generation atomic.Int32
	secondFetched := make(chan struct{})
	o.client = &recoveryClient{fetch: func(string) (*FeatureFile, error) {
		version := int(generation.Add(1))
		if version == 2 {
			close(secondFetched)
		}
		return &FeatureFile{Version: version}, nil
	}}
	entered := make(chan struct{})
	release := make(chan struct{})
	store.writeBackup = func(key string, data []byte) error {
		var file FeatureFile
		if err := json.Unmarshal(data, &file); err != nil {
			return err
		}
		if file.Version == 1 {
			close(entered)
			<-release
		}
		return store.fakeStore.Write(key, data)
	}
	completed := make(chan error, 2)
	go func() { completed <- o.RefreshFeatureFiles() }()
	<-entered
	go func() { completed <- o.RefreshFeatureFiles() }()
	reordered := false
	select {
	case <-secondFetched:
		reordered = true
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	for range 2 {
		if err := <-completed; err != nil {
			t.Fatal(err)
		}
	}
	if reordered {
		t.Fatal("a second refresh fetched state before the first backup finished")
	}
	data, err := store.Read("*:dev-backup.json")
	if err != nil {
		t.Fatal(err)
	}
	var backup FeatureFile
	if err := json.Unmarshal(data, &backup); err != nil {
		t.Fatal(err)
	}
	if backup.Version != 2 {
		t.Fatalf("an older refresh overwrote the newer backup: %+v", backup)
	}
}
