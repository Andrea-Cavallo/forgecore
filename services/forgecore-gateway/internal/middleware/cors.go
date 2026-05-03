package middleware

import (
	"net/http"
	"strings"
)

const (
	headerOrigin      = "Access-Control-Allow-Origin"
	headerMethods     = "Access-Control-Allow-Methods"
	headerHeaders     = "Access-Control-Allow-Headers"
	headerCredentials = "Access-Control-Allow-Credentials"
	headerVary        = "Vary"
	allowedMethods    = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	allowedHeaders    = "Content-Type, Authorization, X-Tenant-ID, X-Request-ID, Idempotency-Key"
)

// CORSMiddleware adds frontend-safe CORS headers for API access.
// allowedOrigins supports "*" or a comma-separated origin allowlist.
func CORSMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	allowlist := parseOrigins(allowedOrigins)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigin, ok := resolveOrigin(r.Header.Get("Origin"), allowlist)
			if !ok && r.Method == http.MethodOptions {
				http.Error(w, `{"code":"cors_forbidden","message":"origin non consentita"}`, http.StatusForbidden)
				return
			}
			setCORSHeaders(w, allowedOrigin)
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseOrigins(value string) map[string]struct{} {
	origins := make(map[string]struct{})
	for _, part := range strings.Split(value, ",") {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins[origin] = struct{}{}
		}
	}
	if len(origins) == 0 {
		origins["*"] = struct{}{}
	}
	return origins
}

func resolveOrigin(requestOrigin string, allowlist map[string]struct{}) (string, bool) {
	if _, ok := allowlist["*"]; ok {
		if requestOrigin != "" {
			return requestOrigin, true
		}
		return "*", true
	}
	if requestOrigin == "" {
		return "", true
	}
	if _, ok := allowlist[requestOrigin]; ok {
		return requestOrigin, true
	}
	return "", false
}

func setCORSHeaders(w http.ResponseWriter, origin string) {
	h := w.Header()
	if origin != "" {
		h.Set(headerOrigin, origin)
	}
	h.Set(headerMethods, allowedMethods)
	h.Set(headerHeaders, allowedHeaders)
	h.Set(headerCredentials, "true")
	h.Add(headerVary, "Origin")
}
