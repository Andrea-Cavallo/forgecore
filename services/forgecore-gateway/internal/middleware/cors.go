package middleware

import "net/http"

const (
	headerOrigin  = "Access-Control-Allow-Origin"
	headerMethods = "Access-Control-Allow-Methods"
	headerHeaders = "Access-Control-Allow-Headers"
)

// CORSMiddleware adds permissive CORS headers for API access.
// In production, configure AllowedOrigins explicitly.
func CORSMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerOrigin, allowedOrigin)
			w.Header().Set(headerMethods, "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set(headerHeaders, "Content-Type, Authorization, X-Tenant-ID, X-Request-ID")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
