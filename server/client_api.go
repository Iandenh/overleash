package server

import (
	"encoding/json"
	"net/http"
)

func (c *Config) registerClientApi(s *http.ServeMux) {
	s.HandleFunc("GET /api/client/features", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(c.Overleash.CachedJson())
	})

	s.HandleFunc("GET /api/client/features/{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flag, _ := c.Overleash.FeatureFile().Features.Get(key)

		writer := json.NewEncoder(w)

		writer.Encode(flag)
	})

	s.HandleFunc("POST /api/client/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.HandleFunc("POST /api/client/register", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
