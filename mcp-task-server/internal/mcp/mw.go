package mcp

import (
	"encoding/json"
	"net/http"
)

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SSE connections and health checks skip auth
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-API-Key")
		if key == "" || key != s.apiKey {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(ErrorResponse(nil, ErrInvalidRequest, "unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
