package overleash

import (
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/go-cmp/cmp"
)

type StreamSubscriber interface {
	Notify(e SseEvent)
}

type SseEvent struct {
	Id    string `json:"id"`
	Event string `json:"event"`
	Data  string `json:"data"`
}

type HydrationEvent struct {
	Type     string    `json:"type"`
	EventId  int       `json:"eventId"`
	Features []Feature `json:"features"`
	Segments []Segment `json:"segments"`
}

func (e *HydrationEvent) GetType() string { return e.Type }
func (e *HydrationEvent) GetEventId() int { return e.EventId }

type SegmentUpdatedEvent struct {
	Type    string  `json:"type"`
	EventId int     `json:"eventId"`
	Segment Segment `json:"segment"`
}

func (e *SegmentUpdatedEvent) GetType() string { return e.Type }
func (e *SegmentUpdatedEvent) GetEventId() int { return e.EventId }

type SegmentRemovedEvent struct {
	Type      string `json:"type"`
	EventId   int    `json:"eventId"`
	SegmentId int    `json:"segmentId"`
}

func (e *SegmentRemovedEvent) GetType() string { return e.Type }
func (e *SegmentRemovedEvent) GetEventId() int { return e.EventId }

type ClientFeaturesDelta struct {
	Events []Event `json:"events"`
}

type Streamer struct {
	subscribers []StreamSubscriber
	mutex       sync.RWMutex
	i           atomic.Int32
}

func (s *Streamer) createNewUpdateDelta(id int, events []Event) SseEvent {
	j, _ := json.Marshal(ClientFeaturesDelta{events})

	return SseEvent{
		Id:    strconv.Itoa(id),
		Event: "unleash-updated",
		Data:  string(j),
	}
}

func (s *Streamer) createNewConnectDelta(id int, events []Event) SseEvent {
	j, _ := json.Marshal(ClientFeaturesDelta{events})

	return SseEvent{
		Id:    strconv.Itoa(id),
		Event: "unleash-connected",
		Data:  string(j),
	}
}

func NewStreamer() *Streamer {
	return &Streamer{
		subscribers: make([]StreamSubscriber, 0),
		mutex:       sync.RWMutex{},
		i:           atomic.Int32{},
	}
}

func (s *Streamer) AddSubscriber(client StreamSubscriber, file FeatureFile) {
	s.subscribers = append(s.subscribers, client)
	h := HydrationEvent{
		Type:     "hydration",
		EventId:  1,
		Features: file.Features,
		Segments: file.Segments,
	}

	client.Notify(s.createNewConnectDelta(1, []Event{&h}))
}

func (s *Streamer) RemoveSubscriber(client StreamSubscriber) {
	newSubs := make([]StreamSubscriber, 0, len(s.subscribers))
	for _, sub := range s.subscribers {
		if sub != client {
			newSubs = append(newSubs, sub)
		}
	}
	s.subscribers = newSubs
}

func (s *Streamer) process(old, new FeatureFile) {
	log.Debug("processing feature file")
	s.mutex.Lock()
	defer s.mutex.Unlock()

	oldFlagsMap := keyFeatureFlags(old)
	newFlagsMap := keyFeatureFlags(new)

	events := make([]Event, 0)

	id := int(s.i.Add(1))
	for flagName, feature := range newFlagsMap {
		oldFeature, ok := oldFlagsMap[flagName]

		if !ok || !cmp.Equal(oldFeature, feature) {
			events = append(events, &FeatureUpdatedEvent{
				Type:    "feature-updated",
				EventId: int(s.i.Add(1)),
				Feature: feature,
			})

			continue
		}
	}

	for _, m := range missingFeatures(oldFlagsMap, newFlagsMap) {
		events = append(events, &FeatureRemovedEvent{
			Type:        "feature-removed",
			EventId:     int(s.i.Add(1)),
			FeatureName: m.Name,
			Project:     m.Project,
		})
	}

	oldSegments := keySegments(old)
	newSegments := keySegments(new)

	for id, segment := range newSegments {
		oldSegment, ok := oldSegments[id]

		if !ok || !cmp.Equal(oldSegment, segment) {
			events = append(events, &SegmentUpdatedEvent{
				Type:    "segment-updated",
				EventId: int(s.i.Add(1)),
				Segment: segment,
			})

			continue
		}
	}

	for _, m := range missingSegments(oldSegments, newSegments) {
		events = append(events, &SegmentRemovedEvent{
			Type:      "segment-removed",
			EventId:   int(s.i.Add(1)),
			SegmentId: m.Id,
		})
	}

	if len(events) == 0 {
		return
	}

	for _, subscriber := range s.subscribers {
		subscriber.Notify(s.createNewUpdateDelta(id, events))
	}
}

func keyFeatureFlags(file FeatureFile) map[string]Feature {
	m := make(map[string]Feature, len(file.Features))

	for _, f := range file.Features {
		m[f.Name] = f
	}

	return m
}
func keySegments(file FeatureFile) map[int]Segment {
	m := make(map[int]Segment, len(file.Segments))

	for _, s := range file.Segments {
		m[s.Id] = s
	}

	return m
}

func missingFeatures(oldMap, newMap map[string]Feature) []Feature {
	var missing []Feature
	for k, f := range oldMap {
		if _, ok := newMap[k]; !ok {
			missing = append(missing, f)
		}
	}
	return missing
}

func missingSegments(oldMap, newMap map[int]Segment) []Segment {
	var missing []Segment
	for k, f := range oldMap {
		if _, ok := newMap[k]; !ok {
			missing = append(missing, f)
		}
	}
	return missing
}

type Event interface {
	GetType() string
	GetEventId() int
}

type FeatureUpdatedEvent struct {
	Type    string  `json:"type"`
	EventId int     `json:"eventId"`
	Feature Feature `json:"feature"`
}

func (e *FeatureUpdatedEvent) GetType() string { return e.Type }
func (e *FeatureUpdatedEvent) GetEventId() int { return e.EventId }

type FeatureRemovedEvent struct {
	Type        string `json:"type"`
	EventId     int    `json:"eventId"`
	FeatureName string `json:"featureName"`
	Project     string `json:"project"`
}

func (e *FeatureRemovedEvent) GetType() string { return e.Type }
func (e *FeatureRemovedEvent) GetEventId() int { return e.EventId }

func (f Feature) Equal(other Feature) bool {
	// Compare primitive fields directly
	if f.Name != other.Name ||
		f.Type != other.Type ||
		f.Enabled != other.Enabled ||
		f.Project != other.Project ||
		f.Strategy != other.Strategy ||
		f.Description != other.Description ||
		f.ImpressionData != other.ImpressionData {
		return false
	}

	// Compare optional bool
	if (f.Stale == nil) != (other.Stale == nil) || (f.Stale != nil && *f.Stale != *other.Stale) {
		return false
	}

	// Compare time pointers
	if !timePtrEqual(f.CreatedAt, other.CreatedAt) {
		return false
	}
	if !timePtrEqual(f.LastSeenAt, other.LastSeenAt) {
		return false
	}

	// Compare slices with cmp (safe if Strategy, Variant, Dependency don't recurse back into Feature)
	opts := []cmp.Option{
		cmp.Comparer(func(t1, t2 time.Time) bool { return t1.Equal(t2) }),
		cmp.Comparer(func(t1, t2 *time.Time) bool {
			if t1 == nil && t2 == nil {
				return true
			}
			if t1 == nil || t2 == nil {
				return false
			}
			return t1.Equal(*t2)
		}),
	}

	if !cmp.Equal(f.Strategies, other.Strategies, opts...) {
		return false
	}
	if !cmp.Equal(f.Variants, other.Variants, opts...) {
		return false
	}
	if !cmp.Equal(f.Dependencies, other.Dependencies, opts...) {
		return false
	}

	// SearchTerm is ignored intentionally
	return true
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

func (s Segment) Equal(other Segment) bool {
	if s.Id != other.Id || s.Name != other.Name {
		return false
	}

	// Compare Constraints slice
	if len(s.Constraints) != len(other.Constraints) {
		return false
	}
	for i := range s.Constraints {
		if !s.Constraints[i].Equal(other.Constraints[i]) {
			return false
		}
	}

	return true
}

func (c Constraint) Equal(other Constraint) bool {
	if c.ContextName != other.ContextName ||
		c.Operator != other.Operator ||
		c.Value != other.Value ||
		c.CaseInsensitive != other.CaseInsensitive ||
		c.Inverted != other.Inverted {
		return false
	}

	// Compare Values slice
	if len(c.Values) != len(other.Values) {
		return false
	}
	for i := range c.Values {
		if c.Values[i] != other.Values[i] {
			return false
		}
	}

	return true
}
