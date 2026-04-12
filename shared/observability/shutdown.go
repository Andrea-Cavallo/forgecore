package observability

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"
)

// WaitForShutdown registers its own signal handler on a fresh background context,
// avoiding double-registration when main also uses signal.NotifyContext.
// It blocks until SIGTERM or SIGINT, then calls cleanup with a 30s timeout.
// Returns normally so callers can defer resource cleanup and exit naturally.
func WaitForShutdown(logger *slog.Logger, cleanup func(ctx context.Context)) {
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-sigCtx.Done()
	logger.Info("segnale di shutdown ricevuto, in attesa drain...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cleanup(shutdownCtx)
	logger.Info("shutdown completato")
}
