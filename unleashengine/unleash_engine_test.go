package unleashengine

import (
	"strings"
	"testing"
)

const validState = `{
  "version": 2,
  "features": [
    {"name": "flag.on",  "enabled": true,  "strategies": [{"name": "default"}]},
    {"name": "flag.off", "enabled": false, "strategies": [{"name": "default"}]}
  ]
}`

func TestTakeStateAcceptsValidState(t *testing.T) {
	engine := NewUnleashEngine()

	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("valid state should be accepted, got: %v", err)
	}
}

// A malformed payload used to abort the whole process: the engine unwrapped the
// parse result, and a panic crossing the FFI boundary cannot unwind. It now
// comes back as an error, which is only useful if we actually surface it.
func TestTakeStateReportsMalformedJson(t *testing.T) {
	engine := NewUnleashEngine()

	err := engine.TakeState("{ this is not json }")

	if err == nil {
		t.Fatal("malformed JSON should be reported as an error")
	}
	if !strings.Contains(err.Error(), "JSON") {
		t.Errorf("error should name the cause, got: %v", err)
	}
}

func TestTakeStateReportsJsonThatIsNotAStateDocument(t *testing.T) {
	engine := NewUnleashEngine()

	if err := engine.TakeState(`{"hello":"world"}`); err == nil {
		t.Fatal("a document that is not an UpdateMessage should be reported")
	}
}

// A rejected update must leave the previously loaded state intact and usable.
func TestRejectedUpdateLeavesEngineUsable(t *testing.T) {
	engine := NewUnleashEngine()

	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := engine.TakeState("not json at all"); err == nil {
		t.Fatal("expected the bad update to be rejected")
	}

	toggle, err := engine.Resolve(&Context{}, "flag.on")
	if err != nil {
		t.Fatalf("engine should still resolve after a rejected update: %v", err)
	}
	if !toggle.GetEnabled() {
		t.Error("flag.on should still be enabled")
	}
}

// An empty Context marshals to zero bytes, so cgo hands the engine a nil data
// pointer. That is normal operation, not abuse, and must be accepted.
func TestResolveWithEmptyContext(t *testing.T) {
	engine := NewUnleashEngine()
	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("setup: %v", err)
	}

	toggle, err := engine.Resolve(&Context{}, "flag.on")
	if err != nil {
		t.Fatalf("an empty context is valid input: %v", err)
	}
	if toggle.GetName() != "flag.on" || !toggle.GetEnabled() {
		t.Errorf("unexpected toggle: %+v", toggle)
	}
}

func TestResolveAllWithEmptyContext(t *testing.T) {
	engine := NewUnleashEngine()
	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("setup: %v", err)
	}

	list, err := engine.ResolveAll(&Context{}, true)
	if err != nil {
		t.Fatalf("an empty context is valid input: %v", err)
	}
	if len(list.GetToggles()) != 2 {
		t.Errorf("expected both toggles, got %d", len(list.GetToggles()))
	}
}

func TestResolveAllExcludesDisabledByDefault(t *testing.T) {
	engine := NewUnleashEngine()
	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("setup: %v", err)
	}

	list, err := engine.ResolveAll(&Context{}, false)
	if err != nil {
		t.Fatalf("resolve all: %v", err)
	}
	for _, toggle := range list.GetToggles() {
		if toggle.GetName() == "flag.off" {
			t.Error("flag.off should be excluded when includeAll is false")
		}
	}
}

func TestResolveUnknownToggle(t *testing.T) {
	engine := NewUnleashEngine()
	if err := engine.TakeState(validState); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if _, err := engine.Resolve(&Context{}, "does.not.exist"); err == nil {
		t.Error("an unknown toggle should be reported as an error")
	}
}

// stateWithWarnings applies cleanly but contains one toggle the engine cannot
// compile, which is routine in a real Unleash feature file.
const stateWithWarnings = `{
  "version": 2,
  "features": [
    {"name": "flag.on", "enabled": true, "strategies": [{"name": "default"}]},
    {
      "name": "flag.uncompilable",
      "enabled": true,
      "strategies": [{
        "name": "default",
        "constraints": [
          {"contextName": "userId", "operator": "NOT_A_REAL_OPERATOR", "values": ["1"]}
        ]
      }]
    }
  ]
}`

// The engine reports warnings for the toggles it could not compile. The state
// was still applied, so this must not be an error — otherwise every refresh of a
// real feature file looks like a failure.
func TestTakeStateAcceptsStateWithCompileWarnings(t *testing.T) {
	engine := NewUnleashEngine()

	if err := engine.TakeState(stateWithWarnings); err != nil {
		t.Fatalf("compile warnings are not a failure, got: %v", err)
	}

	// The toggles that did compile must still work.
	toggle, err := engine.Resolve(&Context{}, "flag.on")
	if err != nil {
		t.Fatalf("flag.on should resolve: %v", err)
	}
	if !toggle.GetEnabled() {
		t.Error("flag.on should be enabled")
	}
}
