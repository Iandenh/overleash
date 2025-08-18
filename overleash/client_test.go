package overleash

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGetFeatures tests the getFeatures method.
func TestGetFeatures(t *testing.T) {
	// Create an expected FeatureFile using your defined types.
	expectedFeatures := FeatureFile{
		Version: 1,
		Features: FeatureFlags{
			{
				Name:           "test-feature",
				Type:           "toggle",
				Enabled:        true,
				Project:        "default",
				Strategies:     []Strategy{},
				Description:    "Test feature",
				ImpressionData: false,
			},
		},
		Segments: []Segment{
			{
				Id:          1,
				Name:        "Test Segment",
				Constraints: []Constraint{},
			},
		},
	}

	// Set up a test server that simulates the /api/client/features endpoint.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/features" {
			t.Errorf("Expected path /api/client/features, got %s", r.URL.Path)
		}
		// Optionally, you can verify required headers.
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(expectedFeatures); err != nil {
			t.Fatalf("Error encoding JSON: %v", err)
		}
	}))
	defer ts.Close()

	// Create a new overleashClient with a dummy interval.
	c := newClient(ts.URL, 1)
	features, err := c.getFeatures("dummy-token")
	if err != nil {
		t.Fatalf("getFeatures returned error: %v", err)
	}

	// Verify the version.
	if features.Version != expectedFeatures.Version {
		t.Errorf("Expected version %d, got %d", expectedFeatures.Version, features.Version)
	}
	// Verify the features slice.
	if len(features.Features) != len(expectedFeatures.Features) {
		t.Fatalf("Expected %d features, got %d", len(expectedFeatures.Features), len(features.Features))
	}
	for i, feature := range features.Features {
		if feature.Name != expectedFeatures.Features[i].Name {
			t.Errorf("Expected feature %s, got %s", expectedFeatures.Features[i].Name, feature.Name)
		}
	}
	// Optionally, you can verify segments.
	if len(features.Segments) != len(expectedFeatures.Segments) {
		t.Errorf("Expected %d segments, got %d", len(expectedFeatures.Segments), len(features.Segments))
	}
}

// TestValidateToken tests the validateToken method.
func TestValidateToken(t *testing.T) {
	expectedEdgeToken := EdgeToken{
		Token:       "valid-token",
		Environment: "test",
	}

	// Set up a test server for the /edge/validate endpoint.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/edge/validate" {
			t.Errorf("Expected path /edge/validate, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		// Read the request body.
		var reqData validationRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		if err := json.Unmarshal(body, &reqData); err != nil {
			t.Fatalf("Error unmarshalling request body: %v", err)
		}
		if len(reqData.Tokens) != 1 || reqData.Tokens[0] != "dummy-token" {
			t.Errorf("Unexpected tokens in request: %+v", reqData.Tokens)
		}

		// Respond with a valid validationResponse.
		resData := validationResponse{
			Tokens: []*EdgeToken{&expectedEdgeToken},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resData); err != nil {
			t.Fatalf("Error encoding response: %v", err)
		}
	}))
	defer ts.Close()

	c := newClient(ts.URL, 1)
	token, err := c.validateToken("dummy-token")
	if err != nil {
		t.Fatalf("validateToken returned error: %v", err)
	}
	if token.Token != expectedEdgeToken.Token {
		t.Errorf("Expected token %s, got %s", expectedEdgeToken.Token, token.Token)
	}
	if token.Environment != expectedEdgeToken.Environment {
		t.Errorf("Expected environment %s, got %s", expectedEdgeToken.Environment, token.Environment)
	}
}

// TestRegisterClient tests the registerClient method.
func TestRegisterClient(t *testing.T) {
	// Set up a test server for the /api/client/register endpoint.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/client/register" {
			t.Errorf("Expected path /api/client/register, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		// Verify the Authorization header.
		auth := r.Header.Get("Authorization")
		if auth != "valid-token" {
			t.Errorf("Expected Authorization header to be 'valid-token', got %s", auth)
		}
		// Read and verify the request body.
		var reqData registerRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Error reading request body: %v", err)
		}
		if err := json.Unmarshal(body, &reqData); err != nil {
			t.Fatalf("Error unmarshalling request body: %v", err)
		}
		if reqData.AppName != "Overleash" {
			t.Errorf("Expected AppName 'Overleash', got %s", reqData.AppName)
		}
		if !strings.HasPrefix(reqData.SdkVersion, "overleash@") {
			t.Errorf("Expected SdkVersion to start with 'overleash@', got %s", reqData.SdkVersion)
		}
		// Check that the interval is as expected.
		expectedInterval := newClient("", 1).interval
		if reqData.Interval != expectedInterval {
			t.Errorf("Expected Interval %d, got %d", expectedInterval, reqData.Interval)
		}
		// Respond with HTTP 200.
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := newClient(ts.URL, 1)
	dummyToken := &EdgeToken{
		Token:       "valid-token",
		Environment: "test",
	}

	// Test a successful registration.
	if err := c.registerClient(dummyToken); err != nil {
		t.Fatalf("registerClient returned error: %v", err)
	}

	// Test failure when the status code is not OK.
	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer tsFail.Close()

	cFail := newClient(tsFail.URL, 1)
	err := cFail.registerClient(dummyToken)
	if err == nil {
		t.Fatalf("Expected error due to non-OK status code, got nil")
	}
}
