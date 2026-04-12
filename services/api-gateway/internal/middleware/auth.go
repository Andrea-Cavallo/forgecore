package middleware

import (
	"net/http"
	"strings"

	"github.com/yourorg/golang-modules/services/api-gateway/internal/clients/authgrpc"
)

// publicPaths are URL path prefixes that bypass JWT authentication.
var publicPaths = []string{
	"/health",
	"/v1/auth/register",
	"/v1/auth/login",
	"/v1/auth/refresh",
	"/v1/auth/forgot-password",
	"/v1/auth/reset-password",
	"/v1/auth/verify-email",
	"/v1/auth/resend-verification",
	"/v1/auth/oauth/",
}

// AuthMiddleware validates JWTs via the auth-service gRPC endpoint.
type AuthMiddleware struct {
	client *authgrpc.Client
}

// NewAuthMiddleware creates an AuthMiddleware backed by an authgrpc.Client.
func NewAuthMiddleware(client *authgrpc.Client) *AuthMiddleware {
	return &AuthMiddleware{client: client}
}

// Middleware returns an http.Handler that enforces JWT authentication
// on all paths not listed in publicPaths.
func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, `{"error":"token mancante"}`, http.StatusUnauthorized)
			return
		}
		resp, err := m.client.ValidateToken(r.Context(), token)
		if err != nil || !resp.Valid {
			http.Error(w, `{"error":"token non valido"}`, http.StatusUnauthorized)
			return
		}
		// Propagate identity headers to upstream services.
		r.Header.Set("X-User-ID", resp.UserID)
		r.Header.Set("X-Tenant-ID", resp.TenantID)
		r.Header.Set("X-User-Roles", strings.Join(resp.Roles, ","))
		next.ServeHTTP(w, r)
	})
}

func isPublicPath(path string) bool {
	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
}
