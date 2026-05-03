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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/infrastructure/postgres"
	transportNATS "github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/transport/nats"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
	natsclient "github.com/nats-io/nats.go"
)

const (
	keyDatabaseURL       = "DATABASE_URL"
	keyNATSURL           = "NATS_URL"
	keyPort              = "PORT"
	defaultAuditDatabase = "postgres://postgres:postgres@localhost:5432/audit?sslmode=disable"
	defaultAddr          = ":8085"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyDatabaseURL, Default: defaultAuditDatabase, Kind: configschema.String},
	{Key: keyNATSURL, Default: natsclient.DefaultURL, Kind: configschema.String},
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
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

	stopHealth := startHealthServer(ctx, cfg.addr, "forgecore-audit", map[string]health.Check{
		"postgres": pool.Ping,
		"nats": func(context.Context) error {
			if !nc.IsConnected() {
				return fmt.Errorf("nats non connesso")
			}
			return nil
		},
	})
	defer stopHealth()

	slog.Info("forgecore-audit avviato", "addr", cfg.addr)
	return consumer.Start(ctx)
}

type config struct {
	databaseURL string
	natsURL     string
	addr        string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		databaseURL: values.String(keyDatabaseURL),
		natsURL:     values.String(keyNATSURL),
		addr:        values.String(keyPort),
	}, nil
}

func startHealthServer(ctx context.Context, addr string, service string, checks map[string]health.Check) func() {
	mux := http.NewServeMux()
	health.Register(mux, service, checks)
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	slog.Info("server health avviato", "servizio", service, "addr", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server health fallito", "servizio", service, "addr", addr, "errore", err)
		}
	}()
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown server health fallito", "servizio", service, "addr", addr, "errore", err)
		}
	}()
	return func() {
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}
}
