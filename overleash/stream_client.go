package overleash

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/charmbracelet/log"
	"github.com/launchdarkly/eventsource"
)

func (fe *FeatureEnvironment) processSseEvent(event eventsource.Event, o *OverleashContext, main bool) {
	eventType := event.Event()

	if eventType != "unleash-connected" && eventType != "unleash-updated" {
		log.Warn("Unknown event type: ", eventType)

		return
	}

	eventData := []byte(event.Data())

	var events Events

	fmt.Printf("%+v\n", string(eventData))

	if err := json.Unmarshal(eventData, &events); err != nil {
		log.Errorf("Unable to unmarshal event data: %v", err)
		return
	}

	fe.processEvents(events, o, main)
}

func (fe *FeatureEnvironment) processEvents(events Events, o *OverleashContext, main bool) {
	o.LockMutex.Lock()
	defer o.LockMutex.Unlock()

	currentFeatures := make(map[string]Feature)

	for _, f := range fe.featureFile.Features {
		currentFeatures[f.Name] = f
	}

	currentSegments := make(map[int]Segment)
	for _, s := range fe.featureFile.Segments {
		currentSegments[s.Id] = s
	}

	for _, event := range events.Events {
		switch e := event.(type) {
		case *HydrationEvent:
			currentFeatures = make(map[string]Feature)
			if e.OriginalFeatures != nil {
				for _, feature := range e.OriginalFeatures {
					currentFeatures[feature.Name] = feature
				}
			} else {
				for _, feature := range e.Features {
					currentFeatures[feature.Name] = feature
				}
			}

			// Replace segments
			currentSegments = make(map[int]Segment)
			for _, segment := range e.Segments {
				currentSegments[segment.Id] = segment
			}

		case *SegmentUpdatedEvent:
			currentSegments[e.Segment.Id] = e.Segment

		case *SegmentRemovedEvent:
			delete(currentSegments, e.SegmentId)

		case *FeatureUpdatedEvent:
			if e.OriginalFeature != nil {
				currentFeatures[e.Feature.Name] = *e.OriginalFeature
			} else {
				currentFeatures[e.Feature.Name] = e.Feature
			}

		case *FeatureRemovedEvent:
			delete(currentFeatures, e.FeatureName)

		case *HydrationOverleashEvent:
			if !main {
				return
			}

			o.paused = e.Paused
			o.overrides = e.Overrides

		default:
			return
		}
	}

	featureSlice := make(FeatureFlags, 0, len(currentFeatures))

	for _, f := range currentFeatures {
		featureSlice = append(featureSlice, f)
	}

	segmentSlice := make([]Segment, 0, len(currentSegments))

	for _, segment := range currentSegments {
		segmentSlice = append(segmentSlice, segment)
	}

	fe.featureFile.Features = featureSlice
	fe.featureFile.Segments = segmentSlice

	fe.compile(o)
	o.lastSync = time.Now()
}
