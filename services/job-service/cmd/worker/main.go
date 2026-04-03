package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio worker fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	slog.Info("job-service worker avviato")
	<-ctx.Done()
	slog.Info("job-service worker terminato")
	return nil
}
