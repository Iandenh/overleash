package unleashengine

import (
	"encoding/json"
	"errors"
	"fmt"
)

// statusCodeOk is the status the engine reports when the state was applied.
const statusCodeOk = "Ok"

// takeStateResponse mirrors the JSON document take_state hands back.
//
// status_code is a string — "Ok", "NotFound" or "Error" — and not a number. The
// Rust enum declares -2/-1/1 discriminants, but serde ignores those for unit
// variants, so decoding this into an int silently fails.
//
// On success, Value lists the warnings raised while compiling the state.
type takeStateResponse struct {
	StatusCode   string   `json:"status_code,omitempty"`
	Value        []string `json:"value,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
}

// parseTakeStateResponse reports whether the engine applied a state update, and
// any warnings it raised while doing so.
//
// Warnings are not failures: the state was applied, and only the toggles named
// in them are affected. A real Unleash feature file routinely contains a toggle
// the engine cannot compile, so warnings must not be surfaced as an error.
//
// Shared by both build variants, and deliberately free of cgo so it can be
// tested without linking the engine.
func parseTakeStateResponse(raw string) ([]string, error) {
	if raw == "" {
		return nil, errors.New("engine returned an empty response")
	}

	var response takeStateResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return nil, fmt.Errorf("could not parse engine response %q: %w", raw, err)
	}

	if response.ErrorMessage != "" {
		return nil, errors.New(response.ErrorMessage)
	}

	if response.StatusCode != statusCodeOk {
		return nil, fmt.Errorf("engine reported status %q", response.StatusCode)
	}

	return response.Value, nil
}
