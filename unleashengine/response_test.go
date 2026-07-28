package unleashengine

import (
	"strings"
	"testing"
)

func TestParseTakeStateResponseCleanUpdate(t *testing.T) {
	warnings, err := parseTakeStateResponse(`{"status_code":"Ok","value":[],"error_message":null}`)

	if err != nil {
		t.Fatalf("a clean update is not an error: %v", err)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

// Warnings mean the state was applied but some toggles could not be compiled.
// Treating that as an error made every refresh of a real feature file look like
// a failure, so it must come back as a success carrying detail.
func TestParseTakeStateResponseWarningsAreNotErrors(t *testing.T) {
	raw := `{"status_code":"Ok","value":["some.flag: could not compile"],"error_message":null}`

	warnings, err := parseTakeStateResponse(raw)

	if err != nil {
		t.Fatalf("warnings must not be reported as an error: %v", err)
	}
	if len(warnings) != 1 || warnings[0] != "some.flag: could not compile" {
		t.Errorf("expected the warning to be returned, got %v", warnings)
	}
}

func TestParseTakeStateResponseRejectedUpdate(t *testing.T) {
	raw := `{"status_code":"Error","value":null,"error_message":"Failed to parse JSON: boom"}`

	warnings, err := parseTakeStateResponse(raw)

	if err == nil {
		t.Fatal("a rejected update must be reported as an error")
	}
	if err.Error() != "Failed to parse JSON: boom" {
		t.Errorf("expected the engine's message, got %q", err.Error())
	}
	if warnings != nil {
		t.Errorf("a rejected update has no warnings, got %v", warnings)
	}
}

func TestParseTakeStateResponseUnexpectedStatus(t *testing.T) {
	if _, err := parseTakeStateResponse(`{"status_code":"NotFound"}`); err == nil {
		t.Error("an unexpected status must be reported")
	}
}

func TestParseTakeStateResponseEmpty(t *testing.T) {
	if _, err := parseTakeStateResponse(""); err == nil {
		t.Error("an empty response must be reported")
	}
}

func TestParseTakeStateResponseUnparseable(t *testing.T) {
	if _, err := parseTakeStateResponse("not json"); err == nil {
		t.Error("an unparseable response must be reported")
	}
}

// A response with no status at all is a different problem from a response with
// an unexpected status, and the message should say which.
func TestParseTakeStateResponseMissingStatus(t *testing.T) {
	_, err := parseTakeStateResponse(`{"value":[]}`)

	if err == nil {
		t.Fatal("a response with no status must be reported")
	}
	if !strings.Contains(err.Error(), "no status") {
		t.Errorf("message should say the status was missing, got: %q", err.Error())
	}
}
