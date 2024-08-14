package server

import (
	"encoding/json"
	"github.com/Iandenh/overleash/proxy"
	"net/http"
)

func (c *Config) registerClientApi(s *http.ServeMux, middleware Middleware) {
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
		if c.proxyMetrics == false {
			w.WriteHeader(http.StatusOK)

			return
		}
		p := proxy.New(c.Overleash.Url())

		err := p.ProxyRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})))

	s.Handle("POST /api/client/register", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
}
