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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/clients/authgrpc"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/middleware"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/proxy"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-gateway/internal/router"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
)

const (
	defaultAddr         = ":8080"
	keyPort             = "PORT"
	keyAuthGRPCAddr     = "AUTH_GRPC_ADDR"
	keyAuthServiceURL   = "AUTH_SERVICE_URL"
	keyPayServiceURL    = "PAYMENT_SERVICE_URL"
	keyNotifServiceURL  = "NOTIF_SERVICE_URL"
	keyAdminServiceURL  = "ADMIN_SERVICE_URL"
	keyAuditServiceURL  = "AUDIT_SERVICE_URL"
	keyPermServiceURL   = "PERMISSION_SERVICE_URL"
	keyConfigServiceURL = "CONFIG_SERVICE_URL"
	keyHookServiceURL   = "WEBHOOK_SERVICE_URL"
	keyStorageURL       = "STORAGE_SERVICE_URL"
	keySubsServiceURL   = "SUBS_SERVICE_URL"
	keyCORSOrigin       = "CORS_ORIGIN"
	keyRateLimit        = "RATE_LIMIT"
	readTimeoutSeconds  = 5
	writeTimeoutSeconds = 10
	idleTimeoutSeconds  = 120
	defaultRateLimit    = "100"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyAuthGRPCAddr, Default: "localhost:9091", Kind: configschema.String},
	{Key: keyAuthServiceURL, Default: "http://localhost:8081", Kind: configschema.String},
	{Key: keyPayServiceURL, Default: "http://localhost:8082", Kind: configschema.String},
	{Key: keyNotifServiceURL, Default: "http://localhost:8083", Kind: configschema.String},
	{Key: keyAdminServiceURL, Default: "http://localhost:8084", Kind: configschema.String},
	{Key: keyAuditServiceURL, Default: "http://localhost:8085", Kind: configschema.String},
	{Key: keyPermServiceURL, Default: "http://localhost:8087", Kind: configschema.String},
	{Key: keyConfigServiceURL, Default: "http://localhost:8088", Kind: configschema.String},
	{Key: keyHookServiceURL, Default: "http://localhost:8089", Kind: configschema.String},
	{Key: keyStorageURL, Default: "http://localhost:8090", Kind: configschema.String},
	{Key: keySubsServiceURL, Default: "http://localhost:8091", Kind: configschema.String},
	{Key: keyCORSOrigin, Default: "*", Kind: configschema.String},
	{Key: keyRateLimit, Default: defaultRateLimit, Kind: configschema.Int},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-gateway fallito", "errore", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig(ctx)
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
	slog.Info("forgecore-gateway avviato", "addr", cfg.addr)

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			slog.Error("shutdown forgecore-gateway fallito", "errore", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

type config struct {
	addr                 string
	authGRPCAddr         string
	authServiceURL       string
	paymentServiceURL    string
	notifServiceURL      string
	permissionServiceURL string
	configServiceURL     string
	webhookServiceURL    string
	storageServiceURL    string
	subsServiceURL       string
	adminServiceURL      string
	auditServiceURL      string
	corsOrigin           string
	rateLimit            int
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:                 values.String(keyPort),
		authGRPCAddr:         values.String(keyAuthGRPCAddr),
		authServiceURL:       values.String(keyAuthServiceURL),
		paymentServiceURL:    values.String(keyPayServiceURL),
		notifServiceURL:      values.String(keyNotifServiceURL),
		adminServiceURL:      values.String(keyAdminServiceURL),
		auditServiceURL:      values.String(keyAuditServiceURL),
		permissionServiceURL: values.String(keyPermServiceURL),
		configServiceURL:     values.String(keyConfigServiceURL),
		webhookServiceURL:    values.String(keyHookServiceURL),
		storageServiceURL:    values.String(keyStorageURL),
		subsServiceURL:       values.String(keySubsServiceURL),
		corsOrigin:           values.String(keyCORSOrigin),
		rateLimit:            values.Int(keyRateLimit),
	}, nil
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
		middleware.RBACMiddleware,
		middleware.AuditMiddleware,
	)

	return chain(r.Build()), nil
}

func buildProxies(cfg config) (map[string]*proxy.ServiceProxy, error) {
	routes := map[string]string{
		"/v1/auth/":          cfg.authServiceURL,
		"/v1/payments/":      cfg.paymentServiceURL,
		"/v1/notifications/": cfg.notifServiceURL,
		"/v1/admin/":         cfg.adminServiceURL,
		"/v1/audit/":         cfg.auditServiceURL,
		"/v1/permissions/":   cfg.permissionServiceURL,
		"/v1/config/":        cfg.configServiceURL,
		"/v1/webhooks/":      cfg.webhookServiceURL,
		"/v1/storage/":       cfg.storageServiceURL,
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
