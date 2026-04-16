package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/Andrea-Cavallo/golang-modules/services/job-service/internal/jobs"
	"github.com/Andrea-Cavallo/golang-modules/services/job-service/internal/scheduler"
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
	cfg := loadConfig()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.redisAddr})
	defer rdb.Close()

	registry := jobs.NewRegistry()
	tokenCleaner := &redisTokenCleaner{rdb: rdb}
	registry.Register(jobs.TypeCleanupTokens, jobs.NewCleanupTokensHandler(tokenCleaner))

	sched := scheduler.NewScheduler()
	sched.Register(scheduler.Task{
		Name:     jobs.TypeCleanupTokens,
		Interval: 1 * time.Hour,
		Run: func(ctx context.Context) error {
			return registry.Dispatch(ctx, jobs.TypeCleanupTokens, []byte(`{"tenant_id":"*"}`))
		},
	})

	slog.Info("job-service worker avviato")
	sched.Start(ctx)
	<-ctx.Done()
	slog.Info("job-service worker terminato")
	return nil
}

type config struct {
	redisAddr string
}

func loadConfig() config {
	return config{redisAddr: envOr("REDIS_ADDR", "localhost:6379")}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
