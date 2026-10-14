package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CAFxX/httpcompression"
)

// This writer models header ownership and counts bytes without storing the
// response. It isolates handler work from httptest's response-body allocation.
type benchmarkResponseWriter struct {
	header http.Header
	bytes  int
}

func (w *benchmarkResponseWriter) Header() http.Header { return w.header }
func (w *benchmarkResponseWriter) WriteHeader(int)     {}
func (w *benchmarkResponseWriter) Write(p []byte) (int, error) {
	w.bytes += len(p)
	return len(p), nil
}

func (w *benchmarkResponseWriter) reset() {
	clear(w.header)
	w.bytes = 0
}

func BenchmarkClientServing(b *testing.B) {
	for _, mode := range []string{"plain", "not-modified", "compression-wrapper", "gzip"} {
		b.Run(mode, func(b *testing.B) {
			o := benchmarkContext(1000)
			mux := http.NewServeMux()
			New(o, b.Context()).registerClientApi(mux)
			var handler http.Handler = mux
			if mode == "compression-wrapper" || mode == "gzip" {
				adapter, err := httpcompression.DefaultAdapter()
				if err != nil {
					b.Fatal(err)
				}
				handler = compress(handler, adapter)
			}
			request := httptest.NewRequest(http.MethodGet, "/api/client/features", nil)
			if mode == "not-modified" {
				request.Header.Set("If-None-Match", "W/\""+o.ActiveFeatureEnvironment().EtagOfCachedJson()+"\"")
			}
			if mode == "gzip" {
				request.Header.Set("Accept-Encoding", "gzip")
			}
			writer := &benchmarkResponseWriter{header: make(http.Header)}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				writer.reset()
				handler.ServeHTTP(writer, request)
			}
		})
	}
}
