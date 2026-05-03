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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/infrastructure/postgres"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/infrastructure/providers/stripe"
	transportHTTP "github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/transport/http"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	defaultAddr         = ":8082"
	keyPort             = "PORT"
	keyDatabaseURL      = "DATABASE_URL"
	keyNATSURL          = "NATS_URL"
	keyStripeSecret     = "STRIPE_SECRET_KEY"
	keyStripeWebhook    = "STRIPE_WEBHOOK_SECRET"
	defaultDatabaseURL  = "postgres://postgres:postgres@localhost:5432/payments?sslmode=disable"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyDatabaseURL, Default: defaultDatabaseURL, Kind: configschema.String},
	{Key: keyNATSURL, Default: nats.DefaultURL, Kind: configschema.String},
	{Key: keyStripeSecret, Kind: configschema.String, Secret: true},
	{Key: keyStripeWebhook, Kind: configschema.String, Secret: true},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-payments fallito", "errore", err)
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
	health.Register(mux, "forgecore-payments", map[string]health.Check{
		"postgres": pool.Ping,
		"nats": func(context.Context) error {
			if !nc.IsConnected() {
				return fmt.Errorf("nats non connesso")
			}
			return nil
		},
	})

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("forgecore-payments avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown forgecore-payments fallito", "errore", err)
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

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:                values.String(keyPort),
		databaseURL:         values.String(keyDatabaseURL),
		natsURL:             values.String(keyNATSURL),
		stripeSecretKey:     values.Secret(keyStripeSecret).Value(),
		stripeWebhookSecret: values.Secret(keyStripeWebhook).Value(),
	}, nil
}
