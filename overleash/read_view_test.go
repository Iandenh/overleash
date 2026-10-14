package overleash

import (
	"bytes"
	"io"
	"testing"
)

func TestPublishedSnapshotsSurviveUpdates(t *testing.T) {
	o := concurrencyContext()
	constraints := make([]Constraint, 1, 8)
	constraints[0] = Constraint{ContextName: "region", Values: []string{"eu"}}
	o.LoadFeatureFile(FeatureFile{
		Features: FeatureFlags{{Name: "flag", Enabled: true, Strategies: []Strategy{{Constraints: constraints}}}},
		Segments: []Segment{{Id: 1, Constraints: constraints}},
	})
	fe := o.ActiveFeatureEnvironment()
	o.AddOverrideConstraint("flag", false, Constraint{ContextName: "userId", Values: []string{"old"}})
	old := fe.Snapshot()
	remote := fe.RemoteSnapshot()
	var before bytes.Buffer
	old.WriteTo(&before)
	o.DeleteOverride("flag")
	o.AddOverrideConstraint("flag", false, Constraint{ContextName: "userId", Values: []string{"new"}})
	if got := old.Get("flag").Strategies()[0].Constraints; len(got) != 2 || got[1].Values[0] != "old" {
		t.Fatal("constraint override mutated a published snapshot")
	}
	if got := remote.Get("flag").Strategies()[0].Constraints; len(got) != 1 || got[0].Values[0] != "eu" {
		t.Fatal("constraint override mutated the remote snapshot")
	}
	rows := old.Features()
	rows[0].SearchTerm = "changed"
	rows[0] = FeatureView{}
	feature, err := old.Feature("flag")
	if err != nil {
		t.Fatal(err)
	}
	feature.Strategies[0].Constraints[0].Values[0] = "changed"
	strategies := old.Get("flag").Strategies()
	strategies[0].Constraints[0].Values[0] = "changed"
	segments := old.SegmentsMap()
	segments[1][0].Values[0] = "changed"
	if old.Get("flag").Strategies()[0].Constraints[0].Values[0] != "eu" || old.SegmentsMap()[1][0].Values[0] != "eu" {
		t.Fatal("mutable getters alias published state")
	}
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "replacement"}}})
	var after bytes.Buffer
	old.WriteTo(&after)
	if before.String() != after.String() || !old.Get("flag").Exists() || fe.Snapshot().Get("flag").Exists() {
		t.Fatal("a refresh changed a retained snapshot")
	}
}

func TestSnapshotReadAllocationBudget(t *testing.T) {
	o := concurrencyContext()
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "flag", Strategies: []Strategy{{Parameters: ParameterMap{"rollout": "50"}}}}}})
	fe := o.ActiveFeatureEnvironment()
	if allocations := testing.AllocsPerRun(100, func() {
		snapshot := fe.Snapshot()
		snapshot.Get("flag").Name()
		snapshot.WriteTo(io.Discard)
	}); allocations != 0 {
		t.Fatalf("snapshot reads allocate: %g allocations, want zero", allocations)
	}
	var rows FeatureViews
	if allocations := testing.AllocsPerRun(100, func() { rows = fe.Snapshot().Features() }); allocations > 1 {
		t.Fatalf("listing views copies nested feature data: %g allocations, want at most one", allocations)
	}
	if len(rows) != 1 {
		t.Fatal("missing feature views")
	}
}

func TestSnapshotNameOrder(t *testing.T) {
	o := concurrencyContext()
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "b"}, {Name: "C"}, {Name: "a"}}})
	fe := o.ActiveFeatureEnvironment()
	snapshot := fe.Snapshot()
	ascending := snapshot.FeaturesByName(false)
	descending := snapshot.FeaturesByName(true)
	for i, name := range []string{"a", "b", "C"} {
		if ascending[i].Name() != name || descending[2-i].Name() != name {
			t.Fatal("snapshot order is not case-insensitive")
		}
	}
	ascending[0] = FeatureView{}
	if snapshot.FeaturesByName(false)[0].Name() != "a" || snapshot.Features()[0].Name() != "b" {
		t.Fatal("sorting aliases snapshot state or changes upstream order")
	}
	o.AddOverride("b", true)
	if !fe.Snapshot().FeaturesByName(false)[1].Enabled() || snapshot.Get("b").Enabled() {
		t.Fatal("name order refers to the wrong compiled version")
	}
	o.LoadFeatureFile(FeatureFile{Features: FeatureFlags{{Name: "z"}, {Name: "y"}, {Name: "x"}}})
	if fe.Snapshot().FeaturesByName(false)[0].Name() != "x" || snapshot.FeaturesByName(false)[0].Name() != "a" {
		t.Fatal("name index was not refreshed or a retained snapshot changed")
	}
}
