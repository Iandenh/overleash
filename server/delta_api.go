package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Iandenh/overleash/overleash"
)

type httpSubscriber struct {
	flusher http.Flusher
	writer  http.ResponseWriter
}

func (h *httpSubscriber) Notify(e overleash.SseEvent) {
	if e.Id != "" {
		fmt.Fprintf(h.writer, "id: %s\n", e.Id)
	}
	if e.Event != "" {
		fmt.Fprintf(h.writer, "event: %s\n", e.Event)
	}

	fmt.Fprintf(h.writer, "data: %s\n\n", e.Data)

	h.flusher.Flush()
	return
}

func (c *Config) registerDelta(s *http.ServeMux) {
	s.HandleFunc("/api/client/streaming", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		subscriber := &httpSubscriber{flusher: flusher, writer: w}

		c.Overleash.ActiveFeatureEnvironment().Streamer.AddSubscriber(subscriber, c.Overleash.ActiveFeatureEnvironment().FeatureFile())

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				c.Overleash.ActiveFeatureEnvironment().Streamer.RemoveSubscriber(subscriber)
				return
			case <-ticker.C:
				fmt.Fprintf(w, ": keep-alive\n\n") // comment line = SSE heartbeat
				flusher.Flush()
			}
		}
	})
}
