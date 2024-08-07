package server

import (
	"net/http"
	"overleash/overleash"
)

func createNewDynamicModeMiddleware(o *overleash.OverleashContext) func(http.Handler) http.Handler {
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
