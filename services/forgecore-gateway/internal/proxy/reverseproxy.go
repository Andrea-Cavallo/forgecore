package proxy

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ServiceProxy wraps a reverse proxy for a single upstream service.
type ServiceProxy struct {
	name   string
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewServiceProxy(name, targetURL string) (*ServiceProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("URL servizio non valido %s: %w", name, err)
	}
	return &ServiceProxy{
		name:   name,
		target: target,
		proxy:  httputil.NewSingleHostReverseProxy(target),
	}, nil
}

func (sp *ServiceProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("proxy richiesta", "servizio", sp.name, "path", r.URL.Path)
	sp.proxy.ServeHTTP(w, r)
}
