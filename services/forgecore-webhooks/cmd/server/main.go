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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/application"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/infrastructure/postgres"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/transport/rest"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultAddr         = ":8089"
	keyPort             = "PORT"
	keyDatabaseURL      = "DATABASE_URL"
	defaultDatabaseURL  = "postgres://postgres:postgres@localhost:5432/webhooks?sslmode=disable"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyDatabaseURL, Default: defaultDatabaseURL, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-webhooks fallito", "errore", err)
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

	endpointRepo := pgRepo.NewEndpointRepository(pool)
	deliveryRepo := pgRepo.NewDeliveryRepository(pool)

	registerUC := application.NewRegisterEndpointUseCase(endpointRepo)
	deliverUC := application.NewDeliverUseCase(endpointRepo, deliveryRepo)

	h := transportREST.NewHandler(registerUC, deliverUC)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	health.Register(mux, "forgecore-webhooks", map[string]health.Check{
		"postgres": pool.Ping,
	})

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("forgecore-webhooks avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown forgecore-webhooks fallito", "errore", err)
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
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:        values.String(keyPort),
		databaseURL: values.String(keyDatabaseURL),
	}, nil
}
