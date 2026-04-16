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
	"github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/application"
	stripeBilling "github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/infrastructure/billing"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/infrastructure/postgres"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/transport/rest"
)

const (
	defaultAddr         = ":8091"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio subscription-service fallito", "errore", err)
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

	subsRepo := pgRepo.NewSubscriptionRepository(pool)
	planRepo := pgRepo.NewPlanRepository(pool)
	billing := stripeBilling.NewStripeProvider(cfg.stripeSecretKey)

	subscribeUC := application.NewSubscribeUseCase(subsRepo, planRepo, billing)
	cancelUC := application.NewCancelUseCase(subsRepo, billing)

	h := transportREST.NewHandler(subscribeUC, cancelUC)
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
	slog.Info("subscription-service avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown subscription-service fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

type config struct {
	addr            string
	databaseURL     string
	stripeSecretKey string
}

func loadConfig() config {
	return config{
		addr:            envOr("PORT", defaultAddr),
		databaseURL:     envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable"),
		stripeSecretKey: envOr("STRIPE_SECRET_KEY", ""),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
