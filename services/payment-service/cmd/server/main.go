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
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/Andrea-Cavallo/golang-modules/services/payment-service/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/payment-service/internal/infrastructure/postgres"
	"github.com/Andrea-Cavallo/golang-modules/services/payment-service/internal/infrastructure/providers/stripe"
	transportHTTP "github.com/Andrea-Cavallo/golang-modules/services/payment-service/internal/transport/http"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

const (
	defaultAddr         = ":8082"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio payment-service fallito", "errore", err)
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

	nc, err := nats.Connect(cfg.natsURL)
	if err != nil {
		return fmt.Errorf("connessione NATS: %w", err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("jetstream init: %w", err)
	}
	publisher := events.NewPublisher(js)
	repo := pgRepo.NewPaymentRepository(pool)
	provider := stripe.NewProvider(cfg.stripeSecretKey, cfg.stripeWebhookSecret)

	createUC := application.NewCreatePaymentUseCase(repo, provider, publisher)
	refundUC := application.NewRefundUseCase(repo, provider, publisher)
	webhookUC := application.NewHandleStripeWebhookUseCase(provider, repo, publisher)

	h := transportHTTP.NewHandler(createUC, refundUC, webhookUC)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	mux.HandleFunc("GET /health", healthHandler)

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("payment-service avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown payment-service fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

type config struct {
	addr                string
	databaseURL         string
	natsURL             string
	stripeSecretKey     string
	stripeWebhookSecret string
}

func loadConfig() config {
	return config{
		addr:                envOr("PORT", defaultAddr),
		databaseURL:         envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/payments?sslmode=disable"),
		natsURL:             envOr("NATS_URL", nats.DefaultURL),
		stripeSecretKey:     envOr("STRIPE_SECRET_KEY", ""),
		stripeWebhookSecret: envOr("STRIPE_WEBHOOK_SECRET", ""),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
