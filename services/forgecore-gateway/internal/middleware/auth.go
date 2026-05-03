package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/clients/authgrpc"
)

// publicPaths are URL path prefixes that bypass JWT authentication.
var publicPaths = []string{
	"/health",
	"/healthz",
	"/readyz",
	"/v1/auth/register",
	"/v1/auth/login",
	"/v1/auth/refresh",
	"/v1/auth/forgot-password",
	"/v1/auth/reset-password",
	"/v1/auth/verify-email",
	"/v1/auth/resend-verification",
	"/v1/auth/oauth/",
}

// AuthMiddleware validates JWTs via the forgecore-auth gRPC endpoint.
type AuthMiddleware struct {
	client *authgrpc.Client
}

type errorResponse struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
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
			writeAuthError(w, r, "missing_token", "token mancante")
			return
		}
		resp, err := m.client.ValidateToken(r.Context(), token)
		if err != nil || !resp.Valid {
			writeAuthError(w, r, "invalid_token", "token non valido")
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

func writeAuthError(w http.ResponseWriter, r *http.Request, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	resp := errorResponse{
		Code:      code,
		Message:   message,
		RequestID: r.Header.Get(HeaderRequestID),
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}
