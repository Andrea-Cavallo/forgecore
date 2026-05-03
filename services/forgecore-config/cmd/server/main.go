package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-config/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-config/internal/infrastructure/postgres"
	redisCache "github.com/Andrea-Cavallo/golang-modules/services/forgecore-config/internal/infrastructure/redis"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/forgecore-config/internal/transport/rest"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	defaultAddr         = ":8088"
	keyPort             = "PORT"
	keyDatabaseURL      = "DATABASE_URL"
	keyRedisAddr        = "REDIS_ADDR"
	defaultDatabaseURL  = "postgres://postgres:postgres@localhost:5432/config?sslmode=disable"
	defaultRedisAddr    = "localhost:6379"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyDatabaseURL, Default: defaultDatabaseURL, Kind: configschema.String},
	{Key: keyRedisAddr, Default: defaultRedisAddr, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-config fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig(ctx)
	if err != nil {
		return fmt.Errorf("caricamento configurazione: %w", err)
	}

	pool, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connessione database: %w", err)
	}
	defer pool.Close()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.redisAddr})
	defer rdb.Close()

	repo := pgRepo.NewConfigRepository(pool)
	cache := redisCache.NewConfigCache(rdb)

	getUC := application.NewGetConfigUseCase(repo, cache)
	setUC := application.NewSetConfigUseCase(repo, cache)

	h := transportREST.NewHandler(getUC, setUC)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	health.Register(mux, "forgecore-config", map[string]health.Check{
		"postgres": pool.Ping,
		"redis": func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
	})

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("forgecore-config avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown forgecore-config fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

type config struct {
	addr        string
	databaseURL string
	redisAddr   string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:        values.String(keyPort),
		databaseURL: values.String(keyDatabaseURL),
		redisAddr:   values.String(keyRedisAddr),
	}, nil
}
