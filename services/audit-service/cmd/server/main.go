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
	"github.com/yourorg/golang-modules/services/audit-service/internal/application"
	pgRepo "github.com/yourorg/golang-modules/services/audit-service/internal/infrastructure/postgres"
	transportNATS "github.com/yourorg/golang-modules/services/audit-service/internal/transport/nats"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio audit-service fallito", "errore", err)
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

	repo := pgRepo.NewAuditRepository(pool)
	recordUC := application.NewRecordEventUseCase(repo)

	consumer, err := transportNATS.NewConsumer(nc, recordUC)
	if err != nil {
		return fmt.Errorf("init consumer audit NATS: %w", err)
	}

	slog.Info("audit-service avviato")
	return consumer.Start(ctx)
}

type config struct {
	databaseURL string
	natsURL     string
}

func loadConfig() config {
	return config{
		databaseURL: envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/audit?sslmode=disable"),
		natsURL:     envOr("NATS_URL", natsclient.DefaultURL),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
