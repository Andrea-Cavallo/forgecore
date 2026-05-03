package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-jobs/internal/jobs"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-jobs/internal/scheduler"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/redis/go-redis/v9"
)

const (
	keyRedisAddr      = "REDIS_ADDR"
	keyHealthAddr     = "HEALTH_ADDR"
	defaultRedisAddr  = "localhost:6379"
	defaultHealthAddr = ":8092"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyRedisAddr, Default: defaultRedisAddr, Kind: configschema.String},
	{Key: keyHealthAddr, Default: defaultHealthAddr, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-jobs fallito", "errore", err)
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
	stopHealth := startHealthServer(ctx, cfg.healthAddr, rdb)
	defer stopHealth()

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

	slog.Info("forgecore-jobs worker avviato", "health_addr", cfg.healthAddr, "redis_addr", cfg.redisAddr)
	sched.Start(ctx)
	<-ctx.Done()
	slog.Info("forgecore-jobs worker terminato")
	return nil
}

type config struct {
	redisAddr  string
	healthAddr string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		redisAddr:  values.String(keyRedisAddr),
		healthAddr: values.String(keyHealthAddr),
	}, nil
}

func startHealthServer(ctx context.Context, addr string, rdb *redis.Client) func() {
	mux := http.NewServeMux()
	health.Register(mux, "forgecore-jobs", map[string]health.Check{
		"redis": func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	})
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("server health avviato", "servizio", "forgecore-jobs", "addr", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server health fallito", "servizio", "forgecore-jobs", "addr", addr, "errore", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown server health fallito", "servizio", "forgecore-jobs", "addr", addr, "errore", err)
		}
	}()
	return func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}
}
