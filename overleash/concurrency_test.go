package overleash

import (
	"sync"
	"testing"

	"github.com/Iandenh/overleash/config"
	"github.com/Iandenh/overleash/unleashengine"
)

func concurrencyContext() *OverleashContext {
	return NewOverleash(&config.Config{Token: "*:dev.test,*:prod.test", Storage: "null", Reload: "0", Streamer: true})
}

func TestReadSnapshotsAreDetached(t *testing.T) {
	o := concurrencyContext()
	value := "original"
	state := FeatureFile{Features: FeatureFlags{{Name: "flag", Enabled: true, Strategies: []Strategy{{Parameters: ParameterMap{"nested": map[string]any{"value": "original"}}, Constraints: []Constraint{{Values: []string{"original"}, Value: &value}}}}}}}
	o.LoadFeatureFile(state)
	constraint := Constraint{Values: []string{"original"}, Value: &value}
	o.AddOverrideConstraint("flag", true, constraint)

	got := o.ActiveFeatureEnvironment().FeatureFile()
	got.Features[0].Name = "changed"
	got.Features[0].Strategies[0].Parameters["nested"].(map[string]any)["value"] = "changed"
	remote := o.ActiveFeatureEnvironment().RemoteFeatureFile()
	remote.Features[0].Strategies[0].Constraints[0].Values[0] = "changed"
	*remote.Features[0].Strategies[0].Constraints[0].Value = "changed"
	data := o.ActiveFeatureEnvironment().CachedJson()
	data[0] = '!'
	o.Overrides()["flag"].Constraints[0].Constraint.Values[0] = "changed"
	*o.GetOverride("flag").Constraints[0].Constraint.Value = "changed"
	o.GetOverride("flag").FeatureFlag = "changed"
	state.Features[0].Name = "input-changed"
	constraint.Values[0] = "input-changed"
	value = "input-changed"
	environments := o.FeatureEnvironments()
	environments[0] = nil

	if got := o.ActiveFeatureEnvironment().RemoteFeatureFile(); got.Features[0].Name != "flag" || got.Features[0].Strategies[0].Constraints[0].Values[0] != "original" || *got.Features[0].Strategies[0].Constraints[0].Value != "original" || got.Features[0].Strategies[0].Parameters["nested"].(map[string]any)["value"] != "original" {
		t.Fatal("feature snapshot or input aliases live state")
	}
	if o.ActiveFeatureEnvironment().CachedJson()[0] == '!' {
		t.Fatal("cached JSON aliases live state")
	}
	if got := o.GetOverride("flag"); got.FeatureFlag != "flag" || got.Constraints[0].Constraint.Values[0] != "original" || *got.Constraints[0].Constraint.Value != "original" {
		t.Fatal("override snapshot aliases live state")
	}
}

type concurrencySubscriber struct{ id int }

func (*concurrencySubscriber) Notify(SseEvent)            {}
func (*concurrencySubscriber) UseActiveEnvironment() bool { return true }
func (*concurrencySubscriber) IsOverleashClient() bool    { return true }

func TestConcurrentStateReadsAndUpdates(t *testing.T) {
	o := concurrencyContext()
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "flag"}}})
	var wg sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		wg.Go(func() {
			for i := 0; i < 100; i++ {
				o.AddOverride("flag", i%2 == 0)
				o.SetPaused(i%2 == 0)
				o.DeleteOverride("flag")
				o.SetFeatureFileIdx(i % 2)
				o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "flag"}}})
			}
		})
		wg.Go(func() {
			for i := 0; i < 100; i++ {
				fe := o.ActiveFeatureEnvironment()
				snapshot := fe.Snapshot()
				snapshot.Get("flag").Name()
				snapshot.Get("flag").Strategies()
				snapshot.Features()
				fe.RemoteSnapshot().Get("flag").Status()
				fe.FeatureFile()
				fe.RemoteFeatureFile()
				fe.CachedJson()
				fe.EtagOfCachedJson()
				o.Overrides()
				o.GetOverride("flag")
				o.HasOverride("flag")
				o.IsPaused()
				o.LastSync()
				o.FeatureFileIdx()
				sub := &concurrencySubscriber{id: i}
				fe.AddStreamerSubscriber(sub)
				fe.RemoveStreamerSubscriber(sub)
			}
		})
	}
	wg.Wait()
}

func TestSubscriberRemovalAfterEnvironmentChange(t *testing.T) {
	o := concurrencyContext()
	original := o.ActiveFeatureEnvironment()
	sub := &concurrencySubscriber{}
	original.AddStreamerSubscriber(sub)
	if err := o.SetFeatureFileIdx(1); err != nil {
		t.Fatal(err)
	}
	if len(original.Streamer.subscriberSnapshot()) != 0 || len(o.ActiveFeatureEnvironment().Streamer.subscriberSnapshot()) != 1 {
		t.Fatal("subscriber was not moved to the active environment")
	}
	original.RemoveStreamerSubscriber(sub)
	if len(o.ActiveFeatureEnvironment().Streamer.subscriberSnapshot()) != 0 {
		t.Fatal("moved subscriber was not removed")
	}
}

// Intentionally unsynchronized: the environment must serialize engine updates
// against evaluation, rather than requiring every engine implementation to do so.
type concurrencyEngine struct{ state string }

func (e *concurrencyEngine) TakeState(state string) error {
	e.state = state
	return nil
}

func (e *concurrencyEngine) Resolve(_ *unleashengine.Context, _ string) (*unleashengine.EvaluatedToggle, error) {
	return &unleashengine.EvaluatedToggle{Name: e.state}, nil
}

func (e *concurrencyEngine) ResolveAll(_ *unleashengine.Context, _ bool) (*unleashengine.EvaluatedToggleList, error) {
	return &unleashengine.EvaluatedToggleList{Toggles: []*unleashengine.EvaluatedToggle{{Name: e.state}}}, nil
}

func TestConcurrentEngineUpdatesAndEvaluation(t *testing.T) {
	o := concurrencyContext()
	fe := o.ActiveFeatureEnvironment()
	fe.engine = &concurrencyEngine{}
	engine := fe.Engine()
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := 0; i < 200; i++ {
			o.LoadFeatureFile(FeatureFile{Version: i})
		}
	})
	wg.Go(func() {
		for i := 0; i < 200; i++ {
			engine.Resolve(&unleashengine.Context{}, "flag")
			engine.ResolveAll(&unleashengine.Context{}, false)
		}
	})
	wg.Wait()
}
