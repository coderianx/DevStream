package middlewars

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsContextKey contextKey = "claims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		tokenString, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Missing bearer token",
			})
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Internal server error",
			})
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&models.Claims{},
			func(t *jwt.Token) (any, error) {
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		)
		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(*models.Claims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{
				"error": "Invalid token claims",
			})
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetClaims(r *http.Request) (*models.Claims, bool) {
	claims, ok := r.Context().Value(claimsContextKey).(*models.Claims)
	return claims, ok
}
