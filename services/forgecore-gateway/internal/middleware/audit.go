package middleware

import (
	"log/slog"
	"net/http"
	"strings"
)

func AuditMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSensitiveAction(r.Method, r.URL.Path) {
			slog.Info("audit sicurezza gateway",
				"servizio", "forgecore-gateway",
				"metodo", r.Method,
				"path", r.URL.Path,
				"tenant_id", r.Header.Get("X-Tenant-ID"),
				"user_id", r.Header.Get("X-User-ID"),
				"request_id", r.Header.Get(HeaderRequestID),
			)
		}
		next.ServeHTTP(w, r)
	})
}

func isSensitiveAction(method string, path string) bool {
	if method == http.MethodGet || method == http.MethodOptions {
		return false
	}
	for _, prefix := range []string{
		"/v1/admin",
		"/v1/config",
		"/v1/payments",
		"/v1/permissions",
		"/v1/storage",
		"/v1/subscriptions",
		"/v1/webhooks",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
