package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/trace/trace/pkg/auth"
)

type contextKey string

const ContextKeyUserID contextKey = "user_id"

// Auth validates the JWT on every request and injects userID into the context.
// Every API call and MCP tool call requires a valid JWT (spec requirement).
func Auth(jwtSvc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractBearerToken(r)
			if token == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}
			userID, err := jwtSvc.Validate(token)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the validated userID from the request context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ContextKeyUserID).(string)
	return v, ok
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(h, "Bearer ")
}
