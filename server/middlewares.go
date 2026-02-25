package server

import (
	"net/http"

	"github.com/Iandenh/overleash/internal/version"
)

type Middleware func(http.Handler) http.Handler

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !version.IsDevelopMode() {
			w.Header().Set("Cache-Control", "max-age=31536000")
		}

		next.ServeHTTP(w, r)
	})
}
