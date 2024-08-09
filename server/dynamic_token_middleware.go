package server

import (
	"net/http"
	"overleash/overleash"
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
