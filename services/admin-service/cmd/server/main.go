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

	"github.com/Andrea-Cavallo/golang-modules/services/admin-service/internal/application"
	transportREST "github.com/Andrea-Cavallo/golang-modules/services/admin-service/internal/transport/rest"
)

const (
	defaultAddr         = ":8084"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio admin-service fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := loadConfig()

	// Stub clients — replace with real gRPC clients in Phase 7.
	tenantClient := &stubTenantClient{}
	userClient := &stubUserClient{}
	statsClient := &stubStatsClient{}

	listTenantsUC := application.NewListTenantsUseCase(tenantClient)
	manageUsersUC := application.NewManageUsersUseCase(userClient)
	statsUC := application.NewGetStatsUseCase(statsClient)

	h := transportREST.NewHandler(listTenantsUC, manageUsersUC, statsUC)
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
	slog.Info("admin-service avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown admin-service fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

type config struct {
	addr string
}

func loadConfig() config {
	return config{addr: envOr("PORT", defaultAddr)}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
