package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Iandenh/overleash/overleash"
)

func (c *Server) registerClientApi(s *http.ServeMux) {
	s.Handle("GET /api/client/features", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := c.featureEnvironmentFromRequest(r)
		snapshot := env.Snapshot()
		etag := snapshot.ETag()

		ifNoneMatch := strings.Trim(strings.TrimPrefix(r.Header.Get("If-None-Match"), "W/"), "\"")

		if ifNoneMatch != "" && ifNoneMatch == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		h := w.Header()
		h.Set("ETag", snapshot.HTTPETag())
		h.Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		snapshot.WriteTo(w)
	}))

	s.Handle("GET /api/client/features/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flag, _ := c.featureEnvironmentFromRequest(r).Snapshot().Feature(key)

		writer := json.NewEncoder(w)

		writer.Encode(flag)
	}))

	s.Handle("POST /api/client/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var metric overleash.MetricsData
		err := decoder.Decode(&metric)

		if err != nil {
			http.Error(w, "Error parsing json", http.StatusBadRequest)
			return
		}

		metric.Environment = c.featureEnvironmentFromRequest(r).Environment()

		c.Overleash.AddMetric(&metric)

		w.WriteHeader(http.StatusOK)
	}))

	s.Handle("POST /api/client/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		var metric overleash.ClientData
		err := decoder.Decode(&metric)

		if err != nil {
			http.Error(w, "Error parsing json", http.StatusBadRequest)
			return
		}

		metric.Environment = c.featureEnvironmentFromRequest(r).Environment()

		c.Overleash.AddRegistration(&metric)

		w.WriteHeader(http.StatusOK)
	}))
}
