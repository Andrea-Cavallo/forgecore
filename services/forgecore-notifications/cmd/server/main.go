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

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-notifications/internal/application"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-notifications/internal/infrastructure/email"
	pgRepo "github.com/Andrea-Cavallo/golang-modules/services/forgecore-notifications/internal/infrastructure/postgres"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-notifications/internal/infrastructure/sms"
	transportNATS "github.com/Andrea-Cavallo/golang-modules/services/forgecore-notifications/internal/transport/nats"
	"github.com/Andrea-Cavallo/golang-modules/shared/configloader"
	"github.com/Andrea-Cavallo/golang-modules/shared/configschema"
	"github.com/Andrea-Cavallo/golang-modules/shared/health"
	"github.com/jackc/pgx/v5/pgxpool"
	natsclient "github.com/nats-io/nats.go"
)

const (
	keyDatabaseURL     = "DATABASE_URL"
	keyPort            = "PORT"
	keyNATSURL         = "NATS_URL"
	keySendGridAPIKey  = "SENDGRID_API_KEY"
	keyFromEmail       = "FROM_EMAIL"
	keyFromName        = "FROM_NAME"
	keyTwilioSID       = "TWILIO_ACCOUNT_SID"
	keyTwilioToken     = "TWILIO_AUTH_TOKEN"
	keyTwilioFrom      = "TWILIO_FROM_NUMBER"
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/notifications?sslmode=disable"
	defaultFromEmail   = "noreply@example.com"
	defaultFromName    = "ForgeCore"
	defaultAddr        = ":8083"
)

var runtimeConfigSchema = configschema.Schema{
	{Key: keyPort, Default: defaultAddr, Kind: configschema.String},
	{Key: keyDatabaseURL, Default: defaultDatabaseURL, Kind: configschema.String},
	{Key: keyNATSURL, Default: natsclient.DefaultURL, Kind: configschema.String},
	{Key: keySendGridAPIKey, Kind: configschema.String, Secret: true},
	{Key: keyFromEmail, Default: defaultFromEmail, Kind: configschema.String},
	{Key: keyFromName, Default: defaultFromName, Kind: configschema.String},
	{Key: keyTwilioSID, Kind: configschema.String, Secret: true},
	{Key: keyTwilioToken, Kind: configschema.String, Secret: true},
	{Key: keyTwilioFrom, Kind: configschema.String},
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		slog.Error("avvio forgecore-notifications fallito", "errore", err)
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

	repo := pgRepo.NewNotificationRepository(pool)
	emailProvider := email.NewSendGridProvider(cfg.sendGridAPIKey, cfg.fromEmail, cfg.fromName)
	smsProvider := sms.NewTwilioProvider(cfg.twilioSID, cfg.twilioToken, cfg.twilioFrom)

	sendUC := application.NewSendUseCase(repo, emailProvider, smsProvider)

	consumers, err := transportNATS.NewConsumers(nc, sendUC)
	if err != nil {
		return fmt.Errorf("init consumers NATS: %w", err)
	}

	stopHealth := startHealthServer(ctx, cfg.addr, "forgecore-notifications", map[string]health.Check{
		"postgres": pool.Ping,
		"nats": func(context.Context) error {
			if !nc.IsConnected() {
				return fmt.Errorf("nats non connesso")
			}
			return nil
		},
	})
	defer stopHealth()

	slog.Info("forgecore-notifications avviato", "addr", cfg.addr)
	return consumers.Start(ctx)
}

type config struct {
	addr           string
	databaseURL    string
	natsURL        string
	sendGridAPIKey string
	fromEmail      string
	fromName       string
	twilioSID      string
	twilioToken    string
	twilioFrom     string
}

func loadConfig(ctx context.Context) (config, error) {
	values, err := configloader.NewDefault(runtimeConfigSchema).Load(ctx)
	if err != nil {
		return config{}, err
	}
	return config{
		addr:           values.String(keyPort),
		databaseURL:    values.String(keyDatabaseURL),
		natsURL:        values.String(keyNATSURL),
		sendGridAPIKey: values.Secret(keySendGridAPIKey).Value(),
		fromEmail:      values.String(keyFromEmail),
		fromName:       values.String(keyFromName),
		twilioSID:      values.Secret(keyTwilioSID).Value(),
		twilioToken:    values.Secret(keyTwilioToken).Value(),
		twilioFrom:     values.String(keyTwilioFrom),
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
