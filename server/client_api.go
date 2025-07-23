package server

import (
	"encoding/json"
	"fmt"
	"github.com/Iandenh/overleash/proxy"
	"net/http"
	"strings"
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
		if !c.proxyMetrics {
			w.WriteHeader(http.StatusOK)

			return
		}
		p := proxy.New(c.Overleash.Upstream())

		err := p.ProxyRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))

	s.Handle("POST /api/client/register", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !c.proxyMetrics {
			w.WriteHeader(http.StatusOK)

			return
		}

		p := proxy.New(c.Overleash.Upstream())

		err := p.ProxyRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
}
