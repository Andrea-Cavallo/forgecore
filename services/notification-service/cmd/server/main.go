package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	natsclient "github.com/nats-io/nats.go"
	"github.com/yourorg/golang-modules/services/notification-service/internal/application"
	pgRepo "github.com/yourorg/golang-modules/services/notification-service/internal/infrastructure/postgres"
	"github.com/yourorg/golang-modules/services/notification-service/internal/infrastructure/email"
	"github.com/yourorg/golang-modules/services/notification-service/internal/infrastructure/sms"
	transportNATS "github.com/yourorg/golang-modules/services/notification-service/internal/transport/nats"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio notification-service fallito", "errore", err)
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

	nc, err := natsclient.Connect(cfg.natsURL)
	if err != nil {
		return fmt.Errorf("connessione NATS: %w", err)
	}
	defer nc.Close()

	repo := pgRepo.NewNotificationRepository(pool)
	emailProvider := email.NewSendGridProvider(cfg.sendGridAPIKey, cfg.fromEmail, cfg.fromName)
	smsProvider := sms.NewTwilioProvider(cfg.twilioSID, cfg.twilioToken, cfg.twilioFrom)

	sendUC := application.NewSendUseCase(repo, emailProvider, smsProvider)

	consumers, err := transportNATS.NewConsumers(nc, sendUC)
	if err != nil {
		return fmt.Errorf("init consumers NATS: %w", err)
	}

	slog.Info("notification-service avviato")
	return consumers.Start(ctx)
}

type config struct {
	databaseURL    string
	natsURL        string
	sendGridAPIKey string
	fromEmail      string
	fromName       string
	twilioSID      string
	twilioToken    string
	twilioFrom     string
}

func loadConfig() config {
	return config{
		databaseURL:    envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable"),
		natsURL:        envOr("NATS_URL", natsclient.DefaultURL),
		sendGridAPIKey: envOr("SENDGRID_API_KEY", ""),
		fromEmail:      envOr("FROM_EMAIL", "noreply@example.com"),
		fromName:       envOr("FROM_NAME", "Superpowers"),
		twilioSID:      envOr("TWILIO_ACCOUNT_SID", ""),
		twilioToken:    envOr("TWILIO_AUTH_TOKEN", ""),
		twilioFrom:     envOr("TWILIO_FROM_NUMBER", ""),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
