package overleash

import (
	"errors"
	"io"
	"slices"
	"strings"
)

// FeatureSnapshot is an immutable view of one published feature file. It remains
// valid across updates. Mutable data is copied only when explicitly requested.
type FeatureSnapshot struct {
	file  FeatureFile
	index *featureIndex
	json  []byte
	etag  string
}

// Name lookups and sort order depend only on upstream names, not overrides.
// Build them once per upstream file and share them across compiled versions.
type featureIndex struct {
	byName    map[string]int
	byNameAsc []int
}

func newFeatureIndex(features FeatureFlags) *featureIndex {
	index := &featureIndex{byName: make(map[string]int, len(features)), byNameAsc: make([]int, len(features))}
	names := make([]string, len(features))
	for i, feature := range features {
		if _, exists := index.byName[feature.Name]; !exists {
			index.byName[feature.Name] = i
		}
		index.byNameAsc[i] = i
		names[i] = strings.ToLower(feature.Name)
	}
	slices.SortFunc(index.byNameAsc, func(a, b int) int { return strings.Compare(names[a], names[b]) })
	return index
}

type environmentSnapshot struct {
	cached FeatureSnapshot
	remote FeatureSnapshot
}

// Snapshot acquires the complete published state without copying its contents.
func (fe *FeatureEnvironment) Snapshot() FeatureSnapshot {
	if state := fe.snapshot.Load(); state != nil {
		return state.cached
	}
	return FeatureSnapshot{}
}

func (fe *FeatureEnvironment) RemoteSnapshot() FeatureSnapshot {
	if state := fe.snapshot.Load(); state != nil {
		return state.remote
	}
	return FeatureSnapshot{}
}

func (s FeatureSnapshot) ETag() string { return s.etag }

// WriteTo writes immutable cached bytes. As required by io.Writer, the writer
// must neither modify nor retain the supplied byte slice.
func (s FeatureSnapshot) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(s.json)
	if n != len(s.json) && err == nil {
		err = io.ErrShortWrite
	}
	return int64(n), err
}

// Features returns a detached list of immutable views. Sorting the list or
// changing a view's SearchTerm does not change the published feature state.
func (s FeatureSnapshot) Features() FeatureViews {
	views := make(FeatureViews, len(s.file.Features))
	for i := range views {
		views[i].feature = &s.file.Features[i]
	}
	return views
}

// FeaturesByName copies lightweight views in a precomputed name order. Each
// request owns its list, but does not repeat sorting or copy nested flag data.
func (s FeatureSnapshot) FeaturesByName(descending bool) FeatureViews {
	views := make(FeatureViews, len(s.file.Features))
	if s.index == nil {
		return views
	}
	for i, position := range s.index.byNameAsc {
		destination := i
		if descending {
			destination = len(views) - 1 - i
		}
		views[destination].feature = &s.file.Features[position]
	}
	return views
}

func (s FeatureSnapshot) Get(name string) FeatureView {
	if s.index != nil {
		if i, ok := s.index.byName[name]; ok {
			return FeatureView{feature: &s.file.Features[i]}
		}
	}
	return FeatureView{}
}

// Feature returns a detached, mutable copy of just the requested feature.
func (s FeatureSnapshot) Feature(name string) (Feature, error) {
	view := s.Get(name)
	if !view.Exists() {
		return Feature{}, errors.New("not found")
	}
	return cloneFeatureFile(FeatureFile{Features: FeatureFlags{*view.feature}}).Features[0], nil
}

func (s FeatureSnapshot) SegmentsMap() map[int][]Constraint {
	segments := make(map[int][]Constraint, len(s.file.Segments))
	for _, segment := range s.file.Segments {
		segments[segment.Id] = cloneConstraints(segment.Constraints)
	}
	return segments
}

type FeatureViews []FeatureView

func (views FeatureViews) Len() int            { return len(views) }
func (views FeatureViews) String(i int) string { return views[i].Name() }

// FeatureView exposes scalars by value and nested data only through copies.
type FeatureView struct {
	feature    *Feature
	SearchTerm string
}

func (v FeatureView) Exists() bool { return v.feature != nil }

func (v FeatureView) Name() string {
	if !v.Exists() {
		return ""
	}
	return v.feature.Name
}

func (v FeatureView) Description() string {
	if !v.Exists() {
		return ""
	}
	return v.feature.Description
}

func (v FeatureView) Enabled() bool { return v.Exists() && v.feature.Enabled }
func (v FeatureView) Stale() bool {
	return v.Exists() && v.feature.Stale != nil && *v.feature.Stale
}

func (v FeatureView) Status() []FeatureFlagStatus {
	if !v.Exists() {
		return nil
	}
	return (FeatureFile{Features: FeatureFlags{*v.feature}}).FeatureFlagStatus(v.Name())
}

func (v FeatureView) Strategies() []Strategy {
	if !v.Exists() {
		return nil
	}
	return cloneFeatureFile(FeatureFile{Features: FeatureFlags{*v.feature}}).Features[0].Strategies
}
