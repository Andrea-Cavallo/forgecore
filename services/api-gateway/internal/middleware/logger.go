package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseRecorder wraps ResponseWriter to capture the status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
	}
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) status() int {
	if !rr.written {
		return http.StatusOK
	}
	return rr.statusCode
}

// LoggerMiddleware logs method, path, status code and duration for every request.
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)
		slog.Info("richiesta",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", r.Header.Get(HeaderRequestID),
			"ip", r.RemoteAddr,
		)
	})
}
