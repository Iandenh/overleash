package unleashengine

import (
	"encoding/json"
	"errors"
	"fmt"
)

// statusCodeOk is the only status that means the state was applied cleanly.
const statusCodeOk = "Ok"

// takeStateResponse mirrors the JSON document take_state hands back.
//
// status_code is a string — "Ok", "NotFound" or "Error" — and not a number. The
// Rust enum declares -2/-1/1 discriminants, but serde ignores those for unit
// variants, so decoding this into an int silently fails.
type takeStateResponse struct {
	StatusCode   string `json:"status_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// parseTakeStateResponse reports whether the engine accepted a state update.
//
// Shared by both build variants, and deliberately free of cgo so it can be
// tested without linking the engine.
func parseTakeStateResponse(raw string) error {
	if raw == "" {
		return errors.New("engine returned an empty response")
	}

	var response takeStateResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return fmt.Errorf("could not parse engine response %q: %w", raw, err)
	}

	// The engine reports a partial update — state applied, but with warnings —
	// as an error carrying the detail, so prefer the message when there is one.
	if response.ErrorMessage != "" {
		return errors.New(response.ErrorMessage)
	}

	if response.StatusCode != statusCodeOk {
		return fmt.Errorf("engine reported status %q", response.StatusCode)
	}

	return nil
}
