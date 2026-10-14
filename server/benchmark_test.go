package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/overleash"
)

func benchmarkContext(size int) *overleash.OverleashContext {
	o := overleash.NewOverleash(&config.Config{Token: "*:dev.benchmark", Storage: "null", Reload: "0"})
	file := overleash.FeatureFile{Version: 2, Features: make(overleash.FeatureFlags, size)}
	for i := range file.Features {
		file.Features[i] = overleash.Feature{
			Name: fmt.Sprintf("feature-%04d", i), Enabled: true, Project: "default",
			Strategies: []overleash.Strategy{{
				Name:        "flexibleRollout",
				Parameters:  overleash.ParameterMap{"rollout": "50", "stickiness": "userId"},
				Constraints: []overleash.Constraint{{ContextName: "region", Operator: overleash.OperatorIn, Values: []string{"eu", "us"}}},
			}},
		}
	}
	o.LoadFeatureFile(file)
	return o
}

func BenchmarkClientFeatures(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			o := benchmarkContext(size)
			mux := http.NewServeMux()
			New(o, b.Context()).registerClientApi(mux)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/client/features", nil))
			}
		})
	}
}

func BenchmarkClientFeaturesParallel(b *testing.B) {
	o := benchmarkContext(1000)
	mux := http.NewServeMux()
	New(o, b.Context()).registerClientApi(mux)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/client/features", nil))
		}
	})
}

func BenchmarkDashboardSearch(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			o := benchmarkContext(size)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				search(httptest.NewRequest(http.MethodGet, "/?sort=name-asc", nil), o)
			}
		})
	}
}

func BenchmarkDashboardRender(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			o := benchmarkContext(size)
			file := o.ActiveFeatureEnvironment().FeatureFile()
			stale := false
			for i := range file.Features {
				file.Features[i].Stale = &stale
			}
			o.LoadFeatureFile(file)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				list := search(httptest.NewRequest(http.MethodGet, "/?sort=name-asc", nil), o)
				if err := featureTemplate(list, o).Render(b.Context(), io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
