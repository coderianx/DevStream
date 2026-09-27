package middlewares

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"backend/internal/store"
)

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		if host, _, err := net.SplitHostPort(fwd); err == nil {
			return host
		}
		return fwd
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

func RateLimitByIP(maxRequests int64, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			key := fmt.Sprintf("ratelimit:%s", clientIP(r))

			count, err := store.Redis.Incr(ctx, key).Result()
			if err != nil {
				log.Println("[ERROR] rate limit incr:", err)
				next.ServeHTTP(w, r)
				return
			}

			if count == 1 {
				store.Redis.Expire(ctx, key, window)
			}

			if count > maxRequests {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]any{
					"error": "Too many requests",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
