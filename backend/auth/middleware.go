package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "auth_user_id"

// RequireAuth is middleware that extracts and validates a Bearer token from the
// Authorization header. If the token is missing or invalid the request is
// rejected with 401 Unauthorized. On success the authenticated user ID is
// stored in the request context (retrievable via GetUserID).
func RequireAuth(authSvc *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				http.Error(w, "missing authorization token", http.StatusUnauthorized)
				return
			}

			userID, err := authSvc.ValidateToken(tokenStr)
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth is middleware that extracts and validates a Bearer token when
// present, but does NOT reject the request if it is missing. This allows
// endpoints to serve both authenticated and guest users during the transition
// period.
func OptionalAuth(authSvc *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr != "" {
				if userID, err := authSvc.ValidateToken(tokenStr); err == nil {
					ctx := context.WithValue(r.Context(), userIDKey, userID)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID returns the authenticated user ID from the request context.
// The second return value indicates whether a user ID was present.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
