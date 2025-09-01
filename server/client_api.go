package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Iandenh/overleash/overleash"
)

func (c *Config) registerClientApi(s *http.ServeMux) {
	s.Handle("GET /api/client/features", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ifNoneMatch := strings.Trim(strings.TrimPrefix(r.Header.Get("If-None-Match"), "W/"), "\"")

		if ifNoneMatch != "" && ifNoneMatch == c.Overleash.ActiveFeatureEnvironment().EtagOfCachedJson() {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", fmt.Sprintf("W/\"%s\"", c.Overleash.ActiveFeatureEnvironment().EtagOfCachedJson()))
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(c.Overleash.ActiveFeatureEnvironment().CachedJson())
	}))

	s.Handle("GET /api/client/features/{key}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		flag, _ := c.Overleash.ActiveFeatureEnvironment().FeatureFile().Features.Get(key)

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

		metric.Environment = strings.SplitN(c.Overleash.ActiveFeatureEnvironment().Name(), ":", 2)[1]

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

		metric.Environment = strings.SplitN(c.Overleash.ActiveFeatureEnvironment().Name(), ":", 2)[1]

		c.Overleash.AddRegistration(&metric)

		w.WriteHeader(http.StatusOK)
	}))
}
