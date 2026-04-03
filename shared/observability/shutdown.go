package observability

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(ctx context.Context, logger *slog.Logger, cleanup func(ctx context.Context)) {
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-ctx.Done()
	logger.Info("segnale di shutdown ricevuto, in attesa drain...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cleanup(shutdownCtx)
	logger.Info("shutdown completato")
	os.Exit(0)
}
