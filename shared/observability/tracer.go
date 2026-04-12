package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer initializes an OTLP HTTP trace exporter and registers it globally.
// Set insecure=true only for local/dev environments; production must use TLS.
func InitTracer(ctx context.Context, service, version, otelEndpoint string, insecure bool) (*sdktrace.TracerProvider, error) {
	opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(otelEndpoint)}
	if insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(service),
		semconv.ServiceVersionKey.String(version),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}
