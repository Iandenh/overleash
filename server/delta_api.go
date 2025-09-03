package server

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Iandenh/overleash/overleash"
)

type httpSubscriber struct {
	flusher     http.Flusher
	writer      http.ResponseWriter
	lock        sync.Mutex
	isOverleash bool
}

func (h *httpSubscriber) Notify(e overleash.SseEvent) {
	if e.OverleashEvent == true && h.isOverleash == false {
		return
	}

	h.lock.Lock()
	defer h.lock.Unlock()

	if e.Id != "" {
		fmt.Fprintf(h.writer, "id: %s\n", e.Id)
	}
	if e.Event != "" {
		fmt.Fprintf(h.writer, "event: %s\n", e.Event)
	}

	fmt.Fprintf(h.writer, "data: %s\n\n", e.Data)

	h.flusher.Flush()
}

func (c *Config) registerDeltaApi(s *http.ServeMux) {
	s.HandleFunc("/api/client/streaming", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		isOverleash := r.Header.Get("X-Overleash") == "yes"
		subscriber := &httpSubscriber{flusher: flusher, writer: w, lock: sync.Mutex{}, isOverleash: isOverleash}

		env := c.featureEnvironmentFromRequest(r)
		env.AddStreamerSubscriber(subscriber, c.Overleash, isOverleash)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				println("Client disconnected")
				env.RemoveStreamerSubscriber(subscriber)
				return
			case <-ticker.C:
				fmt.Fprintf(w, ": keep-alive\n\n") // comment line = SSE heartbeat
				flusher.Flush()
			}
		}
	})
}
