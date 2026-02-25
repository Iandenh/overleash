package server

import (
	"net/http"
	"strings"

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

func compress(next http.Handler, c func(http.Handler) http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/client/streaming") {
			next.ServeHTTP(w, r)
			return
		}

		c(next).ServeHTTP(w, r)
	})
}
