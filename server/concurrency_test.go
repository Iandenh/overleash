package server

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
)

func testStateContext() *overleash.OverleashContext {
	o := overleash.NewOverleash(&config.Config{Token: "*:dev.test", Storage: "null", Reload: "0"})
	o.LoadFeatureFile(overleash.FeatureFile{Features: overleash.FeatureFlags{{Name: "b"}, {Name: "a"}}})
	return o
}

func TestDashboardSearchDoesNotSortSharedState(t *testing.T) {
	o := testStateContext()
	list := search(httptest.NewRequest(http.MethodGet, "/?sort=name-asc", nil), o)
	if list.flags[0].Name != "a" {
		t.Fatal("search did not sort results")
	}
	if o.ActiveFeatureEnvironment().FeatureFile().Features[0].Name != "b" {
		t.Fatal("dashboard sorting mutated cached feature order")
	}
}

func TestConcurrentClientAndDashboardReads(t *testing.T) {
	o := testStateContext()
	s := New(o, t.Context())
	mux := http.NewServeMux()
	s.registerClientApi(mux)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; i < 200; i++ {
			o.AddOverride("b", i%2 == 0)
			o.DeleteOverride("b")
		}
	})
	for reader := 0; reader < 4; reader++ {
		wg.Go(func() {
			for i := 0; i < 200; i++ {
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/client/features", nil))
				want := fmt.Sprintf("W/\"%x\"", sha256.Sum256(w.Body.Bytes()))
				if w.Code != http.StatusOK || w.Header().Get("ETag") != want {
					t.Errorf("response body and ETag are not from the same snapshot")
					return
				}
				search(httptest.NewRequest(http.MethodGet, "/?sort=name-asc&filter=overridden", nil), o)
				stale := false
				if err := feature(overleash.Feature{Name: "b", Stale: &stale}, o, false).Render(t.Context(), io.Discard); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
}
