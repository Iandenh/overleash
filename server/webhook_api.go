package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

func (c *Server) registerWebhookApi(s *http.ServeMux) {
	s.HandleFunc("/api/webhook", func(w http.ResponseWriter, request *http.Request) {
		log.Debug("webhook api refreshed for feature files")

		go func() {
			time.Sleep(time.Second)
			c.Overleash.RefreshFeatureFiles()
		}()

		w.Header().Set("Content-Type", "application/json")
		status := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(status)
	})

}
