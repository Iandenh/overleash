package server

import (
	"encoding/json"
	"net/http"
)

func (c *Config) registerClientApi(s *http.ServeMux) {
	middleware := createNewDynamicModeMiddleware(c.Overleash)

	s.Handle("GET /api/client/features", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(c.Overleash.CachedJson())
	})))

	s.Handle("GET /api/client/features/{key}", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flag, _ := c.Overleash.FeatureFile().Features.Get(key)

		writer := json.NewEncoder(w)

		writer.Encode(flag)
	})))

	s.Handle("POST /api/client/metrics", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	s.Handle("POST /api/client/register", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}
