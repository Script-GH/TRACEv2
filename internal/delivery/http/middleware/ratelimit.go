package middleware

import (
	"net/http"
)

// RateLimit enforces the spec requirements:
//   - Max 3 active sessions per user (prevents duplicate state)
//   - Max 100 API calls per hour per user
//
// Uses Redis as the counter backend (atomic INCR + TTL).
// The Redis client is injected via closure.
func RateLimit(checkFn func(userID string) (allowed bool, err error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "missing user context", http.StatusInternalServerError)
				return
			}
			allowed, err := checkFn(userID)
			if err != nil {
				http.Error(w, "rate limit check error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				w.Header().Set("Retry-After", "3600")
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
