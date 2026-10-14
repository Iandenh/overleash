package overleash

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/unleashengine"
)

type upstreamTransport func(*http.Request) (*http.Response, error)

func (f upstreamTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

var ordinaryUpstreamOperations = []struct {
	name   string
	status int
	body   string
	call   func(*overleashClient) error
}{
	{
		name: "features", status: http.StatusOK,
		body: `{"version":2,"features":[],"segments":[]}`,
		call: func(c *overleashClient) error {
			_, err := c.getFeatures("test-token")
			return err
		},
	},
	{
		name: "validate", status: http.StatusOK,
		body: `{"tokens":[{"token":"test-token"}]}`,
		call: func(c *overleashClient) error {
			_, err := c.validateToken("test-token")
			return err
		},
	},
	{
		name: "register", status: http.StatusOK, body: "acknowledged",
		call: func(c *overleashClient) error {
			return c.registerClient(&EdgeToken{Token: "test-token"})
		},
	},
	{
		name: "metrics", status: http.StatusAccepted, body: "acknowledged",
		call: func(c *overleashClient) error {
			return c.bulkMetrics(c.ctx, "test-token", nil, nil)
		},
	},
}

func TestUpstreamRequestsUseLifecycleContext(t *testing.T) {
	for _, operation := range ordinaryUpstreamOperations {
		t.Run(operation.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			c := newClient("http://upstream.test", time.Minute, ctx)
			c.httpClient.Transport = upstreamTransport(func(r *http.Request) (*http.Response, error) {
				if err := r.Context().Err(); err != nil {
					return nil, err
				}
				return &http.Response{StatusCode: operation.status, Body: io.NopCloser(strings.NewReader(operation.body)), Header: make(http.Header)}, nil
			})
			if err := operation.call(c); !errors.Is(err, context.Canceled) {
				t.Fatalf("request ignored canceled lifecycle context: %v", err)
			}
		})
	}
}

func TestUpstreamRequestsHaveTimeout(t *testing.T) {
	c := newClient("http://upstream.test", time.Minute, t.Context())
	if c.httpClient.Timeout <= 0 || c.httpClient.Timeout > 30*time.Second {
		t.Fatalf("upstream requests are not bounded: timeout=%s", c.httpClient.Timeout)
	}
}

type upstreamBody struct {
	*strings.Reader
	closed bool
}

func (b *upstreamBody) Close() error { b.closed = true; return nil }

func TestUpstreamResponseBodiesAreClosed(t *testing.T) {
	for _, operation := range ordinaryUpstreamOperations {
		for _, failed := range []bool{false, true} {
			name := operation.name + "/success"
			if failed {
				name = operation.name + "/error"
			}
			t.Run(name, func(t *testing.T) {
				body := &upstreamBody{Reader: strings.NewReader(operation.body)}
				status := operation.status
				if failed {
					status = http.StatusServiceUnavailable
				}
				c := newClient("http://upstream.test", time.Minute, t.Context())
				c.httpClient.Transport = upstreamTransport(func(*http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
				})
				err := operation.call(c)
				if failed && err == nil {
					t.Error("non-success HTTP status was accepted")
				}
				if !failed && err != nil {
					t.Error(err)
				}
				if !body.closed {
					t.Error("response body was not closed")
				}
				if !failed && body.Len() != 0 {
					t.Error("successful response body was not drained for connection reuse")
				}
			})
		}
	}
}

type blockedUpstreamClient struct {
	fakeClient
	entered chan struct{}
	release chan struct{}
}

func (c *blockedUpstreamClient) getFeatures(string) (*FeatureFile, error) {
	close(c.entered)
	<-c.release
	return &c.featureFile, c.err
}

func TestUpstreamRefreshDoesNotBlockStateUpdates(t *testing.T) {
	o := NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0"})
	o.ActiveFeatureEnvironment().engine = &concurrencyEngine{}
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "flag"}}})
	c := &blockedUpstreamClient{entered: make(chan struct{}), release: make(chan struct{}), fakeClient: fakeClient{featureFile: FeatureFile{Version: 2, Features: FeatureFlags{{Name: "flag"}}}}}
	o.client = c
	refreshed := make(chan error, 1)
	go func() { refreshed <- o.RefreshFeatureFiles() }()
	<-c.entered
	updated := make(chan struct{})
	go func() {
		_, err := o.ActiveFeatureEnvironment().Engine().Resolve(&unleashengine.Context{}, "flag")
		if err != nil {
			t.Error(err)
		}
		o.AddOverride("flag", true)
		close(updated)
	}()
	blocked := false
	select {
	case <-updated:
	case <-time.After(200 * time.Millisecond):
		blocked = true
	}
	close(c.release)
	if err := <-refreshed; err != nil {
		t.Fatal(err)
	}
	<-updated
	if blocked {
		t.Fatal("upstream network wait held the state lock and blocked evaluation or overrides")
	}
	if !o.ActiveFeatureEnvironment().Snapshot().Get("flag").Enabled() {
		t.Fatal("refresh lost the override applied during fetching")
	}
}

func TestUpstreamCancellationWhileReadingBody(t *testing.T) {
	for _, operation := range ordinaryUpstreamOperations {
		t.Run(operation.name, func(t *testing.T) {
			bodyStarted := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(operation.status)
				w.(http.Flusher).Flush()
				close(bodyStarted)
				<-r.Context().Done()
			}))
			defer server.Close()
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			c := newClient(server.URL, time.Minute, ctx)
			defer c.httpClient.CloseIdleConnections()
			completed := make(chan error, 1)
			go func() { completed <- operation.call(c) }()
			<-bodyStarted
			cancel()
			select {
			case err := <-completed:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("body read did not honor cancellation: %v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("body read did not stop after lifecycle cancellation")
			}
		})
	}
}

func TestUpstreamTimeoutIncludesResponseBody(t *testing.T) {
	for _, operation := range ordinaryUpstreamOperations {
		t.Run(operation.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(operation.status)
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			defer server.Close()
			c := newClient(server.URL, time.Minute, t.Context())
			defer c.httpClient.CloseIdleConnections()
			c.httpClient.Timeout = 30 * time.Millisecond
			if err := operation.call(c); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("stalled response did not time out: %v", err)
			}
		})
	}
}

func TestUpstreamRefreshPreservesNewerRemoteState(t *testing.T) {
	for _, streaming := range []bool{false, true} {
		name := "import"
		if streaming {
			name = "streaming"
		}
		t.Run(name, func(t *testing.T) {
			o := NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0"})
			o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "old"}}})
			c := &blockedUpstreamClient{entered: make(chan struct{}), release: make(chan struct{}), fakeClient: fakeClient{featureFile: FeatureFile{Features: FeatureFlags{{Name: "stale"}}}}}
			o.client = c
			completed := make(chan error, 1)
			go func() { completed <- o.RefreshFeatureFiles() }()
			<-c.entered
			if streaming {
				o.ActiveFeatureEnvironment().processEvents(Events{Events: []Event{&HydrationEvent{Features: FeatureFlags{{Name: "new"}}}}}, o, true)
			} else {
				o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "new"}}})
			}
			close(c.release)
			if err := <-completed; err != nil {
				t.Fatal(err)
			}
			snapshot := o.ActiveFeatureEnvironment().Snapshot()
			if !snapshot.Get("new").Exists() || snapshot.Get("stale").Exists() {
				t.Fatal("late HTTP response overwrote a newer remote update")
			}
		})
	}
}

type blockedBackupStore struct {
	fakeStore
	entered chan struct{}
	release chan struct{}
}

func (s *blockedBackupStore) Write(string, []byte) error {
	close(s.entered)
	<-s.release
	return nil
}

func TestUpstreamBackupDoesNotHoldStateLock(t *testing.T) {
	o := NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0", Backup: true})
	o.client = &fakeClient{featureFile: FeatureFile{Features: FeatureFlags{{Name: "flag"}}}}
	store := &blockedBackupStore{entered: make(chan struct{}), release: make(chan struct{})}
	o.store = store
	completed := make(chan error, 1)
	go func() { completed <- o.RefreshFeatureFiles() }()
	<-store.entered
	read := make(chan bool, 1)
	go func() { read <- o.IsPaused() }()
	blocked := false
	select {
	case <-read:
	case <-time.After(200 * time.Millisecond):
		blocked = true
	}
	close(store.release)
	if err := <-completed; err != nil {
		t.Fatal(err)
	}
	if blocked {
		<-read
		t.Fatal("backup I/O held the state lock")
	}
}

type shutdownMetricsClient struct {
	fakeClient
	flushed chan error
}

func (c *shutdownMetricsClient) bulkMetrics(ctx context.Context, _ string, _ []*ClientData, _ []*MetricsData) error {
	if _, bounded := ctx.Deadline(); !bounded {
		c.flushed <- errors.New("shutdown flush has no deadline")
	} else {
		c.flushed <- ctx.Err()
	}
	return nil
}

func TestUpstreamShutdownMetricsFlushHasIndependentDeadline(t *testing.T) {
	o := NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0"})
	c := &shutdownMetricsClient{flushed: make(chan error, 1)}
	o.client = c
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	o.startMetrics(ctx)
	o.AddMetric(&MetricsData{})
	cancel()
	select {
	case err := <-c.flushed:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown metrics were not flushed")
	}
}

type serialRefreshClient struct {
	fakeClient
	mu     sync.Mutex
	active int
	max    int
}

func (c *serialRefreshClient) getFeatures(string) (*FeatureFile, error) {
	c.mu.Lock()
	c.active++
	if c.active > c.max {
		c.max = c.active
	}
	c.mu.Unlock()
	time.Sleep(time.Millisecond)
	c.mu.Lock()
	c.active--
	c.mu.Unlock()
	return &FeatureFile{}, nil
}

func TestUpstreamRefreshesStaySerialized(t *testing.T) {
	o := NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0"})
	c := &serialRefreshClient{}
	o.client = c
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			if err := o.RefreshFeatureFiles(); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if c.max != 1 {
		t.Fatalf("%d concurrent upstream refreshes, want one", c.max)
	}
}

func TestUpstreamRequestsReuseConnections(t *testing.T) {
	for _, operation := range ordinaryUpstreamOperations {
		t.Run(operation.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.Copy(io.Discard, r.Body)
				w.WriteHeader(operation.status)
				io.WriteString(w, operation.body)
			}))
			defer server.Close()
			var reused []bool
			ctx := httptrace.WithClientTrace(t.Context(), &httptrace.ClientTrace{
				GotConn: func(info httptrace.GotConnInfo) { reused = append(reused, info.Reused) },
			})
			c := newClient(server.URL, time.Minute, ctx)
			defer c.httpClient.CloseIdleConnections()
			for i := 0; i < 2; i++ {
				if err := operation.call(c); err != nil {
					t.Fatal(err)
				}
			}
			if len(reused) != 2 || reused[0] || !reused[1] {
				t.Fatalf("second request did not reuse the first connection: %v", reused)
			}
		})
	}
}

type partialUpstreamClient struct{ fakeClient }

func (*partialUpstreamClient) getFeatures(token string) (*FeatureFile, error) {
	if token == "*:prod.test" {
		return nil, errors.New("upstream unavailable")
	}
	return &FeatureFile{Features: FeatureFlags{{Name: "fresh"}}}, nil
}

func TestUpstreamPartialRefreshPreservesFailedEnvironment(t *testing.T) {
	o := NewOverleash(&config.Config{Token: "*:dev.test,*:prod.test", Storage: "null", Reload: "0"})
	for _, environment := range o.featureEnvironments {
		environment.setRemoteFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "previous"}}})
	}
	o.compileFeatureFiles()
	o.client = &partialUpstreamClient{}
	if err := o.RefreshFeatureFiles(); err == nil {
		t.Fatal("partial failure was not reported")
	}
	if !o.featureEnvironments[0].Snapshot().Get("fresh").Exists() {
		t.Error("successful environment was not refreshed")
	}
	if !o.featureEnvironments[1].Snapshot().Get("previous").Exists() {
		t.Error("failed environment lost its last good state")
	}
}
