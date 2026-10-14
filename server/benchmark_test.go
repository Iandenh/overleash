package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// These benchmarks deliberately include request creation and response buffering.
// BenchmarkClientServing measures handler overhead separately.
func BenchmarkClientFeatures(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			handler := newClientHandler(b, newFeatureContext(size), false)
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/client/features", nil))
			}
		})
	}
}

func BenchmarkClientFeaturesParallel(b *testing.B) {
	handler := newClientHandler(b, newFeatureContext(1000), false)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/client/features", nil))
		}
	})
}

func BenchmarkDashboardSearch(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(fmt.Sprintf("flags=%d", size), func(b *testing.B) {
			o := newFeatureContext(size)
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
			o := newFeatureContext(size)
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
