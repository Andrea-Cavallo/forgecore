package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

// RequestIDMiddleware injects a unique request ID into every request.
// It preserves an existing X-Request-ID if the client provides one.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = uuid.New().String()
			r.Header.Set(HeaderRequestID, id)
		}
		w.Header().Set(HeaderRequestID, id)
		next.ServeHTTP(w, r)
	})
}
