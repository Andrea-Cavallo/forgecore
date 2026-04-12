package router

import (
	"net/http"

	"github.com/yourorg/golang-modules/services/api-gateway/internal/proxy"
)

// Router maps URL prefixes to upstream service proxies.
type Router struct {
	mux *http.ServeMux
}

func New() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (r *Router) Register(prefix string, p *proxy.ServiceProxy) {
	r.mux.Handle(prefix, http.StripPrefix(prefix, p))
}

func (r *Router) Health(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`)) // network errors after headers are sent are non-actionable
}

func (r *Router) Build() http.Handler {
	r.mux.HandleFunc("/health", r.Health)
	return r.mux
}
