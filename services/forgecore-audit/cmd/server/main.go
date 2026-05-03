package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/infrastructure/postgres"
	transportNATS "github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/transport/nats"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/jackc/pgx/v5/pgxpool"
	natsclient "github.com/nats-io/nats.go"
)

const (
	keyDatabaseURL       = "DATABASE_URL"
	keyNATSURL           = "NATS_URL"
	defaultAuditDatabase = "postgres://postgres:postgres@localhost:5432/audit?sslmode=disable"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyDatabaseURL, Default: defaultAuditDatabase, Kind: configschema.String},
	{Key: keyNATSURL, Default: natsclient.DefaultURL, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-audit fallito", "errore", err)
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

	slog.Info("forgecore-audit avviato")
	return consumer.Start(ctx)
}

type config struct {
	databaseURL string
	natsURL     string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		databaseURL: values.String(keyDatabaseURL),
		natsURL:     values.String(keyNATSURL),
	}, nil
}
