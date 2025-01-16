package server

import (
	"github.com/Iandenh/overleash/internal/version"
	"github.com/Iandenh/overleash/overleash"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func createNewDynamicModeMiddleware(o *overleash.OverleashContext) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if o.ShouldDoDynamicCheck() {
				token := r.Header.Get("Authorization")
				if o.AddDynamicToken(token) == false {
					http.Error(w, "No valid token available", http.StatusUnauthorized)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !version.IsDevelopMode() {
			w.Header().Set("Cache-Control", "max-age=31536000")
		}

		next.ServeHTTP(w, r)
	})
}
