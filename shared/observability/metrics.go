package observability

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// HTTPMetrics holds Prometheus counters and histograms for HTTP request tracking.
type HTTPMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
}

// NewHTTPMetrics creates and registers Prometheus metrics for the given service.
func NewHTTPMetrics(service string) *HTTPMetrics {
	m := &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "app",
			Subsystem: service,
			Name:      "http_requests_total",
		}, []string{"method", "path", "status"}),
		RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: service,
			Name:      "http_request_duration_seconds",
			Buckets:   prometheus.DefBuckets,
		}, []string{"method", "path"}),
	}
	prometheus.MustRegister(m.RequestsTotal, m.RequestDuration)
	return m
}

// MetricsMiddleware wraps an http.Handler to record request metrics.
func MetricsMiddleware(m *HTTPMetrics, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rw, r)
		m.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.status)).Inc()
		m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	})
}

// MetricsHandler returns the Prometheus HTTP handler for scraping.
func MetricsHandler() http.Handler { return promhttp.Handler() }

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
