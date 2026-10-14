package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/CAFxX/httpcompression"
	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
)

// newFeatureContext supplies identical fixtures to tests and benchmarks. Null
// storage and disabled frontend evaluation keep measurements focused on Go.
func newFeatureContext(size int) *overleash.OverleashContext {
	o := overleash.NewOverleash(&config.Config{Token: "*:dev.benchmark", Storage: "null", Reload: "0"})
	file := overleash.FeatureFile{Version: 2, Features: make(overleash.FeatureFlags, size)}
	for i := range file.Features {
		file.Features[i] = overleash.Feature{
			Name:    fmt.Sprintf("feature-%04d", i),
			Enabled: true,
			Project: "default",
			Strategies: []overleash.Strategy{{
				Name:       "flexibleRollout",
				Parameters: overleash.ParameterMap{"rollout": "50", "stickiness": "userId"},
				Constraints: []overleash.Constraint{{
					ContextName: "region",
					Operator:    overleash.OperatorIn,
					Values:      []string{"eu", "us"},
				}},
			}},
		}
	}
	o.LoadFeatureFile(file)
	return o
}

func newClientHandler(tb testing.TB, o *overleash.OverleashContext, withCompression bool) http.Handler {
	tb.Helper()
	mux := http.NewServeMux()
	New(o, tb.Context()).registerClientApi(mux)
	if !withCompression {
		return mux
	}
	adapter, err := httpcompression.DefaultAdapter()
	if err != nil {
		tb.Fatal(err)
	}
	return compress(mux, adapter)
}

// discardResponseWriter owns response headers and counts bytes without storing
// the body. It measures handler overhead, not buffering or network transfer.
type discardResponseWriter struct {
	header http.Header
	bytes  int
}

func (w *discardResponseWriter) Header() http.Header { return w.header }
func (w *discardResponseWriter) WriteHeader(int)     {}
func (w *discardResponseWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

func (w *discardResponseWriter) reset() {
	clear(w.header)
	w.bytes = 0
}
