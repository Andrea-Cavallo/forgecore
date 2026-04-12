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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/yourorg/golang-modules/services/config-service/internal/application"
	pgRepo "github.com/yourorg/golang-modules/services/config-service/internal/infrastructure/postgres"
	redisCache "github.com/yourorg/golang-modules/services/config-service/internal/infrastructure/redis"
	transportREST "github.com/yourorg/golang-modules/services/config-service/internal/transport/rest"
)

const (
	defaultAddr         = ":8088"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio config-service fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := loadConfig()

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
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("config-service avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown config-service fallito", "errore", err)
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

func loadConfig() config {
	return config{
		addr:        envOr("PORT", defaultAddr),
		databaseURL: envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/config?sslmode=disable"),
		redisAddr:   envOr("REDIS_ADDR", "localhost:6379"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
