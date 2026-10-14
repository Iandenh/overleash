package server

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CAFxX/httpcompression"
)

func TestCompressionAdapterIsConstructedOnce(t *testing.T) {
	constructed := 0
	compressed := 0
	nextCalls := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextCalls++ })
	handler := compress(next, func(next http.Handler) http.Handler {
		constructed++
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			compressed++
			next.ServeHTTP(w, r)
		})
	})
	for _, path := range []string{"/api/client/features", "/api/client/features", "/api/client/streaming"} {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	if constructed != 1 || compressed != 2 || nextCalls != 3 {
		t.Fatalf("adapter constructions=%d, compressed calls=%d, next calls=%d; want 1, 2, 3", constructed, compressed, nextCalls)
	}
}

func TestClientCompressionRoundTrip(t *testing.T) {
	o := benchmarkContext(100)
	mux := http.NewServeMux()
	New(o, t.Context()).registerClientApi(mux)
	adapter, err := httpcompression.DefaultAdapter()
	if err != nil {
		t.Fatal(err)
	}
	handler := compress(mux, adapter)
	var expected bytes.Buffer
	snapshot := o.ActiveFeatureEnvironment().Snapshot()
	snapshot.WriteTo(&expected)
	for i := 0; i < 3; i++ {
		request := httptest.NewRequest(http.MethodGet, "/api/client/features", nil)
		request.Header.Set("Accept-Encoding", "gzip")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Header().Get("Content-Encoding") != "gzip" || response.Header().Get("ETag") != snapshot.HTTPETag() {
			t.Fatal("incorrect compressed response headers")
		}
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		reader.Close()
		if err != nil || !bytes.Equal(body, expected.Bytes()) {
			t.Fatal("compressed response does not match published JSON")
		}
	}
}

func TestClientServingAllocationBudget(t *testing.T) {
	o := testStateContext()
	mux := http.NewServeMux()
	New(o, t.Context()).registerClientApi(mux)
	request := httptest.NewRequest(http.MethodGet, "/api/client/features", nil)
	writer := &benchmarkResponseWriter{header: make(http.Header)}
	allocations := testing.AllocsPerRun(100, func() {
		writer.reset()
		mux.ServeHTTP(writer, request)
	})
	if allocations > 2 {
		t.Fatalf("cached client response allocated %g times, want at most two header values", allocations)
	}
}
