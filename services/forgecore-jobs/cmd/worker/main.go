package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-jobs/internal/jobs"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-jobs/internal/scheduler"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/redis/go-redis/v9"
)

const (
	keyRedisAddr     = "REDIS_ADDR"
	defaultRedisAddr = "localhost:6379"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyRedisAddr, Default: defaultRedisAddr, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio worker fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig(ctx)
	if err != nil {
		return err
	}

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

	slog.Info("forgecore-jobs worker avviato")
	sched.Start(ctx)
	<-ctx.Done()
	slog.Info("forgecore-jobs worker terminato")
	return nil
}

type config struct {
	redisAddr string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{redisAddr: values.String(keyRedisAddr)}, nil
}
