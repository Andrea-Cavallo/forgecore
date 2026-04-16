package common

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// OTELTransport inietta gli header di tracciamento OTEL in ogni richiesta HTTP.
type OTELTransport struct {
	base http.RoundTripper
}

func NewOTELTransport(base http.RoundTripper) *OTELTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &OTELTransport{base: base}
}

func (t *OTELTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	carrier := propagation.HeaderCarrier(req.Header)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return t.base.RoundTrip(req)
}

// ContextWithToken aggiunge il bearer token all'header Authorization del context.
func InjectBearer(ctx context.Context, req *http.Request, token string) {
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if reqID := ctx.Value(ctxKeyRequestID{}); reqID != nil {
		if s, ok := reqID.(string); ok {
			req.Header.Set("X-Request-ID", s)
		}
	}
}

type ctxKeyRequestID struct{}

// WithRequestID arricchisce il context con un request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKeyRequestID{}, id)
}
