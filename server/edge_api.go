package server

import (
	"net/http"

	"github.com/Iandenh/overleash/proxy"
)

func (c *Config) registerEdgeApi(s *http.ServeMux) {
	s.Handle("POST /edge/validate", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := proxy.New(c.Overleash.Upstream())

		err := p.ProxyRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
}
