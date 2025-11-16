package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"insider-messaging/internal/config"
)

// APIKeyMiddleware API key kontrolü yapan middleware
func APIKeyMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

			if cfg.APIKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.Header.Get("Authorization")
				if strings.HasPrefix(apiKey, "Bearer ") {
					apiKey = strings.TrimPrefix(apiKey, "Bearer ")
				}
			}

			if apiKey != cfg.APIKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(ErrorResponse{
					Error: "Invalid or missing API key. Provide X-API-Key header.",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
