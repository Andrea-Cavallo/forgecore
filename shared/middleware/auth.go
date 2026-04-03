package middleware

import (
	"context"
	"net/http"
	"strings"
)

type claimsKey struct{}

// Claims holds the decoded JWT claims stored in the request context.
type Claims struct {
	UserID   string
	TenantID string
	Roles    []string
}

// ClaimsFromContext retrieves JWT claims from the context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	v, ok := ctx.Value(claimsKey{}).(*Claims)
	return v, ok
}

// TokenVerifier is implemented by the auth-service gRPC client.
type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*Claims, error)
}

// AuthMiddleware validates the Bearer token and stores decoded claims in the context.
// Returns 401 if the token is missing or invalid.
func AuthMiddleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "token di autenticazione mancante", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := verifier.VerifyToken(r.Context(), token)
			if err != nil {
				http.Error(w, "token non valido", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
