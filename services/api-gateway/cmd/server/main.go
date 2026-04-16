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

	"github.com/Andrea-Cavallo/golang-modules/services/api-gateway/internal/clients/authgrpc"
	"github.com/Andrea-Cavallo/golang-modules/services/api-gateway/internal/middleware"
	"github.com/Andrea-Cavallo/golang-modules/services/api-gateway/internal/proxy"
	"github.com/Andrea-Cavallo/golang-modules/services/api-gateway/internal/router"
)

const (
	defaultAddr         = ":8080"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio gateway fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("caricamento configurazione: %w", err)
	}

	authClient, err := authgrpc.NewClient(cfg.authGRPCAddr)
	if err != nil {
		return fmt.Errorf("client gRPC auth: %w", err)
	}
	defer authClient.Close()

	handler, err := buildHandler(cfg, authClient)
	if err != nil {
		return fmt.Errorf("costruzione handler: %w", err)
	}

	srv := &http.Server{
		Addr:         cfg.addr,
		Handler:      handler,
		ReadTimeout:  readTimeoutSeconds * time.Second,
		WriteTimeout: writeTimeoutSeconds * time.Second,
		IdleTimeout:  idleTimeoutSeconds * time.Second,
	}
	slog.Info("api-gateway avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown gateway fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// config holds all runtime configuration read from environment variables.
type config struct {
	addr                string
	authGRPCAddr        string
	authServiceURL      string
	paymentServiceURL   string
	notifServiceURL     string
	permissionServiceURL string
	configServiceURL    string
	webhookServiceURL   string
	storageServiceURL   string
	subsServiceURL      string
	adminServiceURL     string
	auditServiceURL     string
	corsOrigin          string
	rateLimit           int
}

func loadConfig() (config, error) {
	return config{
		addr:                 envOr("PORT", defaultAddr),
		authGRPCAddr:         envOr("AUTH_GRPC_ADDR", "localhost:9091"),
		authServiceURL:       envOr("AUTH_SERVICE_URL", "http://localhost:8081"),
		paymentServiceURL:    envOr("PAYMENT_SERVICE_URL", "http://localhost:8082"),
		notifServiceURL:      envOr("NOTIF_SERVICE_URL", "http://localhost:8083"),
		adminServiceURL:      envOr("ADMIN_SERVICE_URL", "http://localhost:8084"),
		auditServiceURL:      envOr("AUDIT_SERVICE_URL", "http://localhost:8085"),
		permissionServiceURL: envOr("PERMISSION_SERVICE_URL", "http://localhost:8087"),
		configServiceURL:     envOr("CONFIG_SERVICE_URL", "http://localhost:8088"),
		webhookServiceURL:    envOr("WEBHOOK_SERVICE_URL", "http://localhost:8089"),
		storageServiceURL:    envOr("STORAGE_SERVICE_URL", "http://localhost:8090"),
		subsServiceURL:       envOr("SUBS_SERVICE_URL", "http://localhost:8091"),
		corsOrigin:           envOr("CORS_ORIGIN", "*"),
		rateLimit:            100,
	}, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func buildHandler(cfg config, authClient *authgrpc.Client) (http.Handler, error) {
	proxies, err := buildProxies(cfg)
	if err != nil {
		return nil, err
	}

	r := router.New()
	for prefix, p := range proxies {
		r.Register(prefix, p)
	}

	rateLimiter := middleware.NewRateLimiter(cfg.rateLimit)
	authMW := middleware.NewAuthMiddleware(authClient)

	// Middleware chain (outermost → innermost):
	// RequestID → Logger → SecurityHeaders → CORS → RateLimit → Auth → Router
	chain := chain(
		middleware.RequestIDMiddleware,
		middleware.LoggerMiddleware,
		middleware.SecurityHeadersMiddleware,
		middleware.CORSMiddleware(cfg.corsOrigin),
		middleware.RateLimitMiddleware(rateLimiter),
		authMW.Middleware,
	)

	return chain(r.Build()), nil
}

func buildProxies(cfg config) (map[string]*proxy.ServiceProxy, error) {
	routes := map[string]string{
		"/v1/auth/":         cfg.authServiceURL,
		"/v1/payments/":     cfg.paymentServiceURL,
		"/v1/notifications/": cfg.notifServiceURL,
		"/v1/admin/":        cfg.adminServiceURL,
		"/v1/audit/":        cfg.auditServiceURL,
		"/v1/permissions/":  cfg.permissionServiceURL,
		"/v1/config/":       cfg.configServiceURL,
		"/v1/webhooks/":     cfg.webhookServiceURL,
		"/v1/storage/":      cfg.storageServiceURL,
		"/v1/subscriptions/": cfg.subsServiceURL,
	}
	proxies := make(map[string]*proxy.ServiceProxy, len(routes))
	for prefix, url := range routes {
		p, err := proxy.NewServiceProxy(prefix, url)
		if err != nil {
			return nil, fmt.Errorf("proxy %s: %w", prefix, err)
		}
		proxies[prefix] = p
	}
	return proxies, nil
}

// chain composes middleware functions right-to-left.
func chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
