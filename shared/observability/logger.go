package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type contextKey string

const loggerKey contextKey = "logger"

type ServiceInfo struct {
	Service string
	Version string
	Env     string
}

func NewLogger(service, version, env string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})).With(
		"service", service,
		"version", version,
		"env", env,
	)
}

func NewServiceLogger(info ServiceInfo, w io.Writer) *slog.Logger {
	return NewLogger(info.Service, info.Version, info.Env, w)
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
