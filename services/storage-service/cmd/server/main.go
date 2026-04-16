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
	"github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/application"
	minioProvider "github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/infrastructure/minio"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/infrastructure/postgres"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/storage-service/internal/transport/rest"
)

const (
	defaultAddr         = ":8090"
	readTimeoutSeconds  = 30 // longer for file uploads
	writeTimeoutSeconds = 30
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio storage-service fallito", "errore", err)
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
	slog.Info("storage-service avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown storage-service fallito", "errore", err)
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

func loadConfig() config {
	return config{
		addr:           envOr("PORT", defaultAddr),
		databaseURL:    envOr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/storage?sslmode=disable"),
		minioEndpoint:  envOr("MINIO_ENDPOINT", "localhost:9000"),
		minioAccessKey: envOr("MINIO_ACCESS_KEY", "minioadmin"),
		minioSecretKey: envOr("MINIO_SECRET_KEY", "minioadmin"),
		minioSSL:       os.Getenv("MINIO_SSL") == "true",
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
