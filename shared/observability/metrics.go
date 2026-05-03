package observability

import (
	"errors"
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

type OperationalMetrics struct {
	OutboxMessagesTotal      *prometheus.CounterVec
	IdempotencyRequestsTotal *prometheus.CounterVec
	ProviderCallsTotal       *prometheus.CounterVec
	JobRunsTotal             *prometheus.CounterVec
}

// NewHTTPMetrics creates and registers Prometheus metrics for the given service.
// Safe to call multiple times (e.g., in tests): returns the previously registered collector on conflict.
func NewHTTPMetrics(service string) (*HTTPMetrics, error) {
	reqTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "app",
		Subsystem: service,
		Name:      "http_requests_total",
	}, []string{"method", "path", "status"})
	reqDur := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: service,
		Name:      "http_request_duration_seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path"})
	reqTotal = mustRegisterOrExisting(reqTotal).(*prometheus.CounterVec)
	reqDur = mustRegisterOrExisting(reqDur).(*prometheus.HistogramVec)
	return &HTTPMetrics{RequestsTotal: reqTotal, RequestDuration: reqDur}, nil
}

// mustRegisterOrExisting registers c; if already registered returns the existing collector.
func mustRegisterOrExisting(c prometheus.Collector) prometheus.Collector {
	if err := prometheus.Register(c); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			return are.ExistingCollector
		}
		panic(err)
	}
	return c
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

func NewOperationalMetrics(service string) *OperationalMetrics {
	return &OperationalMetrics{
		OutboxMessagesTotal: mustRegisterOrExisting(prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "forgecore",
			Subsystem: service,
			Name:      "outbox_messages_total",
		}, []string{"status"})).(*prometheus.CounterVec),
		IdempotencyRequestsTotal: mustRegisterOrExisting(prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "forgecore",
			Subsystem: service,
			Name:      "idempotency_requests_total",
		}, []string{"operation", "result"})).(*prometheus.CounterVec),
		ProviderCallsTotal: mustRegisterOrExisting(prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "forgecore",
			Subsystem: service,
			Name:      "provider_calls_total",
		}, []string{"provider", "operation", "status"})).(*prometheus.CounterVec),
		JobRunsTotal: mustRegisterOrExisting(prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "forgecore",
			Subsystem: service,
			Name:      "job_runs_total",
		}, []string{"job", "status"})).(*prometheus.CounterVec),
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
