package middleware

import (
	"encoding/json"
	"net/http"
)

type AuthMiddleware struct {
	apiKey string
}

func NewAuthMiddleware(apiKey string) *AuthMiddleware {
	return &AuthMiddleware{apiKey: apiKey}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			respondWithError(w, http.StatusUnauthorized, "Missing API key")
			return
		}

		if apiKey != m.apiKey {
			respondWithError(w, http.StatusUnauthorized, "Invalid API key")
			return
		}

		next(w, r)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
