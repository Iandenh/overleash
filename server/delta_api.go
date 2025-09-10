package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Iandenh/overleash/overleash"
	"github.com/charmbracelet/log"
)

type httpSubscriber struct {
	flusher           http.Flusher
	writer            http.ResponseWriter
	isOverleashClient bool
	send              chan overleash.SseEvent
}

func (h *httpSubscriber) Notify(e overleash.SseEvent) {
	select {
	case h.send <- e:
	default:
		log.Printf("dropping event for slow subscriber (overleash=%v)", h.isOverleashClient)
	}
}

func (h *httpSubscriber) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-h.send:
			if !ok {
				return
			}

			// only write Overleash Events when connected to overleash client
			if e.OverleashEvent && !h.isOverleashClient {
				continue
			}
			if err := h.writeEvent(e); err != nil {
				log.Printf("subscriber write error: %v", err)
				return
			}
		}
	}
}

func (h *httpSubscriber) writeEvent(e overleash.SseEvent) error {
	if e.Id != "" {
		if _, err := fmt.Fprintf(h.writer, "id: %s\n", e.Id); err != nil {
			return err
		}
	}

	if e.Event != "" {
		if _, err := fmt.Fprintf(h.writer, "event: %s\n", e.Event); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(h.writer, "data: %s\n\n", e.Data); err != nil {
		return err
	}

	h.flusher.Flush()

	return nil
}

func (c *Server) registerDeltaApi(s *http.ServeMux) {
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
		subscriber := &httpSubscriber{flusher: flusher, writer: w, isOverleashClient: isOverleash, send: make(chan overleash.SseEvent, 32)}

		env := c.featureEnvironmentFromRequest(r)
		env.AddStreamerSubscriber(subscriber, c.Overleash, isOverleash)
		defer env.RemoveStreamerSubscriber(subscriber)

		ctx := r.Context()
		go subscriber.run(ctx)

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				log.Printf("SSE client disconnected (overleash=%v)", isOverleash)
				return
			case <-ticker.C:
				if _, err := fmt.Fprintf(w, ": keep-alive\n\n"); err != nil {
					log.Printf("failed to write heartbeat: %v", err)
					return
				}
				flusher.Flush()
			}
		}
	})
}
