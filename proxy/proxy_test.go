package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestProxyRequest sets up an upstream server and verifies that
func TestProxyRequest(t *testing.T) {
	// Create an upstream server that responds with a header and a body echoing the URL.
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// You can inspect r.URL.Path and r.URL.RawQuery here to validate path and query merging.
		w.Header().Set("X-Upstream", "true")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Upstream reached. Path: " + r.URL.Path + ", Query: " + r.URL.RawQuery))
	}))
	defer upstreamServer.Close()

	// Create a Proxy instance using the upstream server URL.
	proxyInstance := New(upstreamServer.URL)

	// Build a request that will be proxied.
	// For example, let’s use a nonempty path and a query parameter.
	req := httptest.NewRequest("GET", "/testpath?foo=bar", nil)
	// Use ResponseRecorder to capture the proxy’s response.
	recorder := httptest.NewRecorder()

	// Execute the proxy request.
	if err := proxyInstance.ProxyRequest(recorder, req); err != nil {
		t.Fatalf("ProxyRequest returned an error: %v", err)
	}

	// Verify that the response code from the upstream was propagated.
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Verify that the custom header from the upstream is present.
	if got := recorder.Header().Get("X-Upstream"); got != "true" {
		t.Errorf("Expected header X-Upstream to be 'true', got '%s'", got)
	}

	// Read and validate the response body.
	bodyBytes, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	bodyStr := string(bodyBytes)

	// Check that the upstream response body contains the expected path and query.
	if !strings.Contains(bodyStr, "/testpath") {
		t.Errorf("Expected response body to contain '/testpath', got %q", bodyStr)
	}
	if !strings.Contains(bodyStr, "foo=bar") {
		t.Errorf("Expected response body to contain 'foo=bar', got %q", bodyStr)
	}
}

// TestProxyRequest_PathConcatenation verifies that the proxy concatenates paths correctly.
func TestProxyRequest_PathConcatenation(t *testing.T) {
	// Create an upstream server whose URL has a non-empty path.
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with the path the upstream sees.
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Received path: " + r.URL.Path))
	}))
	defer upstreamServer.Close()

	// Append a base path to the upstream URL.
	baseURL := upstreamServer.URL + "/base"
	proxyInstance := New(baseURL)

	// Create a request with a path to be concatenated.
	req := httptest.NewRequest("GET", "/subpath", nil)
	recorder := httptest.NewRecorder()

	if err := proxyInstance.ProxyRequest(recorder, req); err != nil {
		t.Fatalf("ProxyRequest returned an error: %v", err)
	}

	// Read the response body.
	bodyBytes, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatalf("Error reading response body: %v", err)
	}
	bodyStr := string(bodyBytes)

	// The expected path is a naive concatenation: "/base/subpath".
	// (Depending on your needs, you may want to fix this logic to use path.Join.)
	expectedPath := "/base/subpath"
	if !strings.Contains(bodyStr, expectedPath) {
		t.Errorf("Expected response body to contain path %q, got %q", expectedPath, bodyStr)
	}
}
