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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/application"
	minioProvider "github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/infrastructure/minio"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/infrastructure/postgres"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/transport/rest"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultAddr         = ":8090"
	keyPort             = "PORT"
	keyDatabaseURL      = "DATABASE_URL"
	keyMinioEndpoint    = "MINIO_ENDPOINT"
	keyMinioAccessKey   = "MINIO_ACCESS_KEY"
	keyMinioSecretKey   = "MINIO_SECRET_KEY"
	keyMinioSSL         = "MINIO_SSL"
	defaultDatabaseURL  = "postgres://postgres:postgres@localhost:5432/storage?sslmode=disable"
	defaultMinioAddr    = "localhost:9000"
	defaultMinioAccess  = "minioadmin"
	defaultMinioSecret  = "minioadmin"
	readTimeoutSeconds  = 30 // longer for file uploads
	writeTimeoutSeconds = 30
	idleTimeoutSeconds  = 120
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyDatabaseURL, Default: defaultDatabaseURL, Kind: configschema.String},
	{Key: keyMinioEndpoint, Default: defaultMinioAddr, Kind: configschema.String},
	{Key: keyMinioAccessKey, Default: defaultMinioAccess, Kind: configschema.String, Secret: true},
	{Key: keyMinioSecretKey, Default: defaultMinioSecret, Kind: configschema.String, Secret: true},
	{Key: keyMinioSSL, Default: "false", Kind: configschema.Bool},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-storage fallito", "errore", err)
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

	storage, err := minioProvider.NewProvider(cfg.minioEndpoint, cfg.minioAccessKey, cfg.minioSecretKey, cfg.minioSSL)
	if err != nil {
		return fmt.Errorf("init minio: %w", err)
	}

	fileRepo := pgRepo.NewFileRepository(pool)
	uploadUC := application.NewUploadUseCase(fileRepo, storage)
	presignUC := application.NewGeneratePresignedUseCase(fileRepo, storage)

	h := transportREST.NewHandler(uploadUC, presignUC)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	health.Register(mux, "forgecore-storage", map[string]health.Check{
		"postgres": pool.Ping,
	})

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      mux,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("forgecore-storage avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown forgecore-storage fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

type config struct {
	addr           string
	databaseURL    string
	minioEndpoint  string
	minioAccessKey string
	minioSecretKey string
	minioSSL       bool
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:           values.String(keyPort),
		databaseURL:    values.String(keyDatabaseURL),
		minioEndpoint:  values.String(keyMinioEndpoint),
		minioAccessKey: values.Secret(keyMinioAccessKey).Value(),
		minioSecretKey: values.Secret(keyMinioSecretKey).Value(),
		minioSSL:       values.Bool(keyMinioSSL),
	}, nil
}
