package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

const (
	statusOK   = "ok"
	statusFail = "fail"
)

type Check func(context.Context) error

type Response struct {
	Service string            `json:"service"`
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks,omitempty"`
}

func Register(mux *http.ServeMux, service string, checks map[string]Check) {
	handler := Handler(service, checks)
	mux.Handle("GET /health", handler)
	mux.Handle("GET /healthz", handler)
	mux.Handle("GET /readyz", Handler(service, checks))
}

func Handler(service string, checks map[string]Check) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		resp, status := runChecks(ctx, service, checks)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			return
		}
	})
}

func runChecks(ctx context.Context, service string, checks map[string]Check) (Response, int) {
	resp := Response{Service: service, Status: statusOK, Checks: map[string]string{}}
	for name, check := range checks {
		if err := check(ctx); err != nil {
			resp.Status = statusFail
			resp.Checks[name] = err.Error()
			continue
		}
		resp.Checks[name] = statusOK
	}
	if resp.Status == statusFail {
		return resp, http.StatusServiceUnavailable
	}
	return resp, http.StatusOK
}
