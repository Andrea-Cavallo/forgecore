package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type requestIDKey struct{}

const HeaderRequestID = "X-Request-ID"

// RequestIDFromContext retrieves the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// RequestIDMiddleware injects a unique request ID into every request context and response header.
// It uses the incoming X-Request-ID header if present, otherwise generates a new UUID.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), requestIDKey{}, reqID)
		w.Header().Set(HeaderRequestID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
