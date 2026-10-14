package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkClientServing(b *testing.B) {
	cases := []struct {
		name            string
		withCompression bool
		conditional     bool
		acceptEncoding  string
	}{
		{name: "plain"},
		{name: "not-modified", conditional: true},
		{name: "compression-wrapper", withCompression: true},
		{name: "gzip", withCompression: true, acceptEncoding: "gzip"},
	}
	for _, testCase := range cases {
		b.Run(testCase.name, func(b *testing.B) {
			o := newFeatureContext(1000)
			handler := newClientHandler(b, o, testCase.withCompression)
			request := httptest.NewRequest(http.MethodGet, "/api/client/features", nil)
			if testCase.conditional {
				request.Header.Set("If-None-Match", o.ActiveFeatureEnvironment().Snapshot().HTTPETag())
			}
			if testCase.acceptEncoding != "" {
				request.Header.Set("Accept-Encoding", testCase.acceptEncoding)
			}
			writer := &discardResponseWriter{header: make(http.Header)}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				writer.reset()
				handler.ServeHTTP(writer, request)
			}
		})
	}
}
