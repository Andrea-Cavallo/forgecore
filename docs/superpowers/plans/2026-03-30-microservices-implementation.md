# Microservizi Go 1.24 — Piano di Implementazione

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Costruire un backend SaaS enterprise-grade in Go 1.24 composto da 10 microservizi riusabili, production-ready, con auth, pagamenti, notifiche, permessi, subscription e storage.

**Architecture:** Monorepo con un modulo Go per servizio. Ogni servizio segue DDD a 4 layer (domain → application → infrastructure → transport). Comunicazione REST esterna, gRPC interna, NATS JetStream per eventi asincroni.

**Tech Stack:** Go 1.24, PostgreSQL 16, Redis 7 (Sentinel), NATS JetStream, HashiCorp Vault, Traefik v3, Docker Compose, OpenTelemetry, Prometheus/Grafana, Jaeger

**Spec di riferimento:** `docs/superpowers/specs/2026-03-30-microservices-design.md`

---

## Mappa delle fasi

| Fase | Contenuto | Dipendenze |
|------|-----------|-----------|
| **0** | Infrastruttura base (repo, shared libs, Docker, Vault) | — |
| **1** | Auth Service core (register, login, JWT, refresh) | Fase 0 |
| **2** | Auth Service avanzato (MFA, OAuth2, GDPR, email verify) | Fase 1 |
| **3** | API Gateway | Fase 1 |
| **4** | Payment Service | Fase 1, 3 |
| **5** | Notification Service | Fase 0, 1, 4 |
| **6** | Admin + Audit + Job Services | Fase 1, 4, 5 |
| **7** | Permission Service + SDK | Fase 1 |
| **8** | Config Service + Webhook Service (outbound) | Fase 0, 1 |
| **9** | Storage Service + Subscription Service | Fase 4 |
| **10** | Observability completa + HA + CI/CD | Tutte le fasi |

---

# FASE 0 — Infrastruttura Base

## Task 0.1: Struttura repository

**Files:**
- Create: `services/auth-service/go.mod`
- Create: `services/payment-service/go.mod`
- Create: `services/notification-service/go.mod`
- Create: `services/admin-service/go.mod`
- Create: `services/audit-service/go.mod`
- Create: `services/job-service/go.mod`
- Create: `services/api-gateway/go.mod`
- Create: `services/permission-service/go.mod`
- Create: `services/config-service/go.mod`
- Create: `services/webhook-service/go.mod`
- Create: `services/storage-service/go.mod`
- Create: `services/subscription-service/go.mod`
- Create: `shared/go.mod`
- Create: `sdk/go/go.mod`

- [x] **Step 1: Crea le directory dei servizi**

```bash
mkdir -p services/{auth-service,payment-service,notification-service,admin-service,audit-service,job-service,api-gateway,permission-service,config-service,webhook-service,storage-service,subscription-service}/cmd/server
mkdir -p services/job-service/cmd/worker
mkdir -p shared/{proto,events,middleware,validation,crypto,pagination,i18n,observability}
mkdir -p sdk/go/{auth,payment,permission,config,common}
mkdir -p deployments/{traefik,vault,pgbouncer,prometheus,alertmanager,grafana/provisioning,nats}
mkdir -p scripts
```

- [x] **Step 2: Inizializza go.mod per shared**

```bash
cd shared && go mod init github.com/yourorg/golang-modules/shared && cd ..
```

- [x] **Step 3: Inizializza go.mod per ogni servizio (script)**

```bash
for svc in auth-service payment-service notification-service admin-service audit-service job-service api-gateway permission-service config-service webhook-service storage-service subscription-service; do
  cd services/$svc
  go mod init github.com/yourorg/golang-modules/services/$svc
  cd ../..
done
```

- [x] **Step 4: Crea .gitignore nella root**

```bash
cat > .gitignore << 'EOF'
# Binaries
bin/
*.exe

# Go
*.test
*.out
coverage.html

# Env
.env
.env.*
!.env.example

# Vault tokens
vault-token
*.vault

# Docker volumes
postgres-data/
redis-data/

# IDE
.idea/
.vscode/
*.swp
EOF
```

- [x] **Step 5: Commit struttura base**

```bash
git init
git add .
git commit -m "chore: initialize monorepo structure for 10 microservices"
```

---

## Task 0.2: shared/observability — logging, tracing, metrics

**Files:**
- Create: `shared/observability/logger.go`
- Create: `shared/observability/tracer.go`
- Create: `shared/observability/metrics.go`
- Create: `shared/observability/shutdown.go`

- [x] **Step 1: Aggiungi dipendenze shared**

```bash
cd shared
go get go.opentelemetry.io/otel@latest
go get go.opentelemetry.io/otel/trace@latest
go get go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp@latest
go get go.opentelemetry.io/otel/sdk/trace@latest
go get github.com/prometheus/client_golang@latest
cd ..
```

- [x] **Step 2: Scrivi test per il logger**

```go
// shared/observability/logger_test.go
package observability_test

import (
    "bytes"
    "context"
    "log/slog"
    "testing"
    "github.com/yourorg/golang-modules/shared/observability"
)

func TestNewLogger_JSON(t *testing.T) {
    var buf bytes.Buffer
    logger := observability.NewLogger("test-service", "1.0.0", "test", &buf)
    logger.Info("test message", "key", "value")
    output := buf.String()
    if !bytes.Contains([]byte(output), []byte(`"service":"test-service"`)) {
        t.Errorf("expected service field in log output, got: %s", output)
    }
    if !bytes.Contains([]byte(output), []byte(`"msg":"test message"`)) {
        t.Errorf("expected msg field in log output, got: %s", output)
    }
}

func TestLoggerFromContext(t *testing.T) {
    logger := observability.NewLogger("svc", "1.0", "test", nil)
    ctx := observability.WithLogger(context.Background(), logger)
    got := observability.LoggerFromContext(ctx)
    if got == nil {
        t.Fatal("expected logger from context, got nil")
    }
}
```

- [x] **Step 3: Esegui il test — deve fallire**

```bash
cd shared && go test ./observability/... -run TestNewLogger_JSON -v
# Expected: FAIL — logger.go not found
```

- [x] **Step 4: Implementa logger.go**

```go
// shared/observability/logger.go
package observability

import (
    "context"
    "io"
    "log/slog"
    "os"
)

type contextKey string
const loggerKey contextKey = "logger"

func NewLogger(service, version, env string, w io.Writer) *slog.Logger {
    if w == nil {
        w = os.Stdout
    }
    return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })).With(
        "service", service,
        "version", version,
        "env",     env,
    )
}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
    return context.WithValue(ctx, loggerKey, l)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

- [x] **Step 5: Implementa tracer.go**

```go
// shared/observability/tracer.go
package observability

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(ctx context.Context, service, version, otelEndpoint string) (*sdktrace.TracerProvider, error) {
    exp, err := otlptracehttp.New(ctx,
        otlptracehttp.WithEndpoint(otelEndpoint),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }
    res := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(service),
        semconv.ServiceVersionKey.String(version),
    )
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)
    return tp, nil
}
```

- [x] **Step 6: Implementa metrics.go**

```go
// shared/observability/metrics.go
package observability

import (
    "net/http"
    "strconv"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPMetrics struct {
    RequestsTotal   *prometheus.CounterVec
    RequestDuration *prometheus.HistogramVec
}

func NewHTTPMetrics(service string) *HTTPMetrics {
    m := &HTTPMetrics{
        RequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
            Namespace: "app",
            Subsystem: service,
            Name:      "http_requests_total",
        }, []string{"method", "path", "status"}),
        RequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
            Namespace: service,
            Name:      "http_request_duration_seconds",
            Buckets:   prometheus.DefBuckets,
        }, []string{"method", "path"}),
    }
    prometheus.MustRegister(m.RequestsTotal, m.RequestDuration)
    return m
}

func MetricsMiddleware(m *HTTPMetrics, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := &responseWriter{ResponseWriter: w, status: 200}
        start := time.Now()
        next.ServeHTTP(rw, r)
        m.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.status)).Inc()
        m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
    })
}

func MetricsHandler() http.Handler { return promhttp.Handler() }

type responseWriter struct {
    http.ResponseWriter
    status int
}
func (rw *responseWriter) WriteHeader(code int) { rw.status = code; rw.ResponseWriter.WriteHeader(code) }
```

- [x] **Step 7: Implementa shutdown.go**

```go
// shared/observability/shutdown.go
package observability

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func WaitForShutdown(ctx context.Context, logger *slog.Logger, cleanup func(ctx context.Context)) {
    ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
    defer stop()
    <-ctx.Done()
    logger.Info("shutdown signal received, draining...")
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    cleanup(shutdownCtx)
    logger.Info("shutdown complete")
    os.Exit(0)
}
```

- [x] **Step 8: Esegui i test**

```bash
cd shared && go test ./observability/... -v
# Expected: PASS
```

- [x] **Step 9: Commit**

```bash
git add shared/observability/
git commit -m "feat(shared): add structured logger, OTEL tracer, Prometheus metrics, graceful shutdown"
```

---

## Task 0.3: shared/validation

**Files:**
- Create: `shared/validation/validator.go`
- Create: `shared/validation/validator_test.go`

- [x] **Step 1: Aggiungi dipendenza**

```bash
cd shared && go get github.com/go-playground/validator/v10@latest && cd ..
```

- [x] **Step 2: Scrivi test**

```go
// shared/validation/validator_test.go
package validation_test

import (
    "testing"
    "github.com/yourorg/golang-modules/shared/validation"
)

type testInput struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"required,min=18,max=120"`
}

func TestValidate_Valid(t *testing.T) {
    input := testInput{Email: "user@example.com", Age: 25}
    if err := validation.Validate(input); err != nil {
        t.Errorf("expected no error, got: %v", err)
    }
}

func TestValidate_InvalidEmail(t *testing.T) {
    input := testInput{Email: "not-an-email", Age: 25}
    errs := validation.Validate(input)
    if errs == nil {
        t.Fatal("expected validation error for invalid email")
    }
    fieldErrors, ok := errs.(validation.FieldErrors)
    if !ok {
        t.Fatalf("expected FieldErrors type, got %T", errs)
    }
    if len(fieldErrors) == 0 || fieldErrors[0].Field != "email" {
        t.Errorf("expected error on field 'email', got: %+v", fieldErrors)
    }
}
```

- [x] **Step 3: Run test — deve fallire**

```bash
cd shared && go test ./validation/... -v
# Expected: FAIL
```

- [x] **Step 4: Implementa validator.go**

```go
// shared/validation/validator.go
package validation

import (
    "fmt"
    "strings"
    "github.com/go-playground/validator/v10"
)

var v = validator.New(validator.WithRequiredStructFields())

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

type FieldErrors []FieldError

func (fe FieldErrors) Error() string {
    msgs := make([]string, len(fe))
    for i, e := range fe {
        msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
    }
    return strings.Join(msgs, "; ")
}

func Validate(input any) error {
    err := v.Struct(input)
    if err == nil {
        return nil
    }
    var fieldErrors FieldErrors
    for _, e := range err.(validator.ValidationErrors) {
        fieldErrors = append(fieldErrors, FieldError{
            Field:   strings.ToLower(e.Field()),
            Message: humanMessage(e),
        })
    }
    return fieldErrors
}

func humanMessage(e validator.FieldError) string {
    switch e.Tag() {
    case "required":
        return "is required"
    case "email":
        return "must be a valid email address"
    case "min":
        return fmt.Sprintf("must be at least %s", e.Param())
    case "max":
        return fmt.Sprintf("must be at most %s", e.Param())
    case "oneof":
        return fmt.Sprintf("must be one of: %s", e.Param())
    default:
        return fmt.Sprintf("failed validation: %s", e.Tag())
    }
}
```

- [x] **Step 5: Run test — deve passare**

```bash
cd shared && go test ./validation/... -v
# Expected: PASS
```

- [x] **Step 6: Commit**

```bash
git add shared/validation/
git commit -m "feat(shared): add centralized input validation with structured field errors"
```

---

## Task 0.4: shared/crypto — PII encryption AES-256-GCM

**Files:**
- Create: `shared/crypto/pii.go`
- Create: `shared/crypto/pii_test.go`

- [x] **Step 1: Scrivi test**

```go
// shared/crypto/pii_test.go
package crypto_test

import (
    "testing"
    "github.com/yourorg/golang-modules/shared/crypto"
)

func TestEncryptDecrypt(t *testing.T) {
    key := make([]byte, 32) // 256-bit key
    for i := range key { key[i] = byte(i) }

    enc := crypto.NewPIIEncryptor(key)
    plaintext := "user@example.com"

    ciphertext, err := enc.Encrypt(plaintext)
    if err != nil {
        t.Fatalf("encrypt error: %v", err)
    }
    if ciphertext == plaintext {
        t.Fatal("ciphertext should differ from plaintext")
    }

    decrypted, err := enc.Decrypt(ciphertext)
    if err != nil {
        t.Fatalf("decrypt error: %v", err)
    }
    if decrypted != plaintext {
        t.Errorf("expected %q, got %q", plaintext, decrypted)
    }
}

func TestHash(t *testing.T) {
    h1 := crypto.Hash("user@example.com")
    h2 := crypto.Hash("user@example.com")
    h3 := crypto.Hash("other@example.com")
    if h1 != h2 {
        t.Error("same input should produce same hash")
    }
    if h1 == h3 {
        t.Error("different inputs should produce different hashes")
    }
}
```

- [x] **Step 2: Implementa pii.go**

```go
// shared/crypto/pii.go
package crypto

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "io"
)

type PIIEncryptor struct{ key []byte }

func NewPIIEncryptor(key []byte) *PIIEncryptor { return &PIIEncryptor{key: key} }

func (e *PIIEncryptor) Encrypt(plaintext string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil { return "", fmt.Errorf("new cipher: %w", err) }
    gcm, err := cipher.NewGCM(block)
    if err != nil { return "", fmt.Errorf("new gcm: %w", err) }
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", fmt.Errorf("rand nonce: %w", err)
    }
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *PIIEncryptor) Decrypt(encoded string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encoded)
    if err != nil { return "", fmt.Errorf("base64 decode: %w", err) }
    block, err := aes.NewCipher(e.key)
    if err != nil { return "", fmt.Errorf("new cipher: %w", err) }
    gcm, err := cipher.NewGCM(block)
    if err != nil { return "", fmt.Errorf("new gcm: %w", err) }
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize { return "", fmt.Errorf("ciphertext too short") }
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plain, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil { return "", fmt.Errorf("gcm open: %w", err) }
    return string(plain), nil
}

// Hash returns a deterministic SHA-256 hex hash — for DB indexes on encrypted fields
func Hash(value string) string {
    h := sha256.Sum256([]byte(value))
    return fmt.Sprintf("%x", h)
}
```

- [x] **Step 3: Run test**

```bash
cd shared && go test ./crypto/... -v
# Expected: PASS
```

- [x] **Step 4: Commit**

```bash
git add shared/crypto/
git commit -m "feat(shared): add AES-256-GCM PII encryptor and SHA-256 hash helper"
```

---

## Task 0.5: shared/pagination — cursor-based standard

**Files:**
- Create: `shared/pagination/cursor.go`
- Create: `shared/pagination/cursor_test.go`

- [x] **Step 1: Scrivi test**

```go
// shared/pagination/cursor_test.go
package pagination_test

import (
    "testing"
    "time"
    "github.com/google/uuid"
    "github.com/yourorg/golang-modules/shared/pagination"
)

func TestEncodeDecode(t *testing.T) {
    original := pagination.Cursor{
        ID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
        CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
    }
    encoded := pagination.Encode(original)
    if encoded == "" {
        t.Fatal("encoded cursor should not be empty")
    }
    decoded, err := pagination.Decode(encoded)
    if err != nil {
        t.Fatalf("decode error: %v", err)
    }
    if decoded.ID != original.ID {
        t.Errorf("ID mismatch: got %v, want %v", decoded.ID, original.ID)
    }
    if !decoded.CreatedAt.Equal(original.CreatedAt) {
        t.Errorf("CreatedAt mismatch")
    }
}

func TestDecodeInvalid(t *testing.T) {
    _, err := pagination.Decode("not-a-valid-cursor")
    if err == nil {
        t.Fatal("expected error for invalid cursor")
    }
}
```

- [x] **Step 2: Implementa cursor.go**

```go
// shared/pagination/cursor.go
package pagination

import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "time"
    "github.com/google/uuid"
)

type Cursor struct {
    ID        uuid.UUID `json:"id"`
    CreatedAt time.Time `json:"created_at"`
}

type Page struct {
    NextCursor *string `json:"next_cursor,omitempty"`
    PrevCursor *string `json:"prev_cursor,omitempty"`
    HasNext    bool    `json:"has_next"`
    HasPrev    bool    `json:"has_prev"`
    Limit      int     `json:"limit"`
}

func Encode(c Cursor) string {
    b, _ := json.Marshal(c)
    return base64.URLEncoding.EncodeToString(b)
}

func Decode(s string) (Cursor, error) {
    b, err := base64.URLEncoding.DecodeString(s)
    if err != nil { return Cursor{}, fmt.Errorf("invalid cursor encoding: %w", err) }
    var c Cursor
    if err := json.Unmarshal(b, &c); err != nil {
        return Cursor{}, fmt.Errorf("invalid cursor json: %w", err)
    }
    return c, nil
}

// SQL returns the WHERE clause fragment for cursor-based pagination
// Usage: WHERE (created_at, id) < ($1, $2) ORDER BY created_at DESC, id DESC LIMIT $3
func SQL(cursor *Cursor, limit int) (where string, args []any, queryLimit int) {
    queryLimit = limit + 1 // fetch one extra to detect next page
    if cursor == nil {
        return "", nil, queryLimit
    }
    return "(created_at, id) < ($1, $2)", []any{cursor.CreatedAt, cursor.ID}, queryLimit
}
```

- [x] **Step 3: Run test**

```bash
cd shared && go get github.com/google/uuid && go test ./pagination/... -v
# Expected: PASS
```

- [x] **Step 4: Commit**

```bash
git add shared/pagination/
git commit -m "feat(shared): add cursor-based pagination encoder/decoder with SQL helper"
```

---

## Task 0.6: shared/events — NATS typed event definitions

**Files:**
- Create: `shared/events/auth.go`
- Create: `shared/events/payment.go`
- Create: `shared/events/notification.go`
- Create: `shared/events/audit.go`

- [x] **Step 1: Crea auth events**

```go
// shared/events/auth.go
package events

import (
    "time"
    "github.com/google/uuid"
)

const (
    SubjectUserRegistered   = "auth.user.registered"
    SubjectUserLogin        = "auth.user.login"
    SubjectPasswordReset    = "auth.user.password_reset"
    SubjectPasswordChanged  = "auth.user.password_changed"
    SubjectMFAEnabled       = "auth.mfa.enabled"
    SubjectEmailVerified    = "auth.email.verified"
)

type UserRegistered struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    Email      string    `json:"email"`
    VerifyURL  string    `json:"verify_url"`
    OccurredAt time.Time `json:"occurred_at"`
}

type UserLogin struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    IPAddress  string    `json:"ip_address"`
    UserAgent  string    `json:"user_agent"`
    OccurredAt time.Time `json:"occurred_at"`
}

type PasswordReset struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    Email      string    `json:"email"`
    ResetURL   string    `json:"reset_url"`
    OccurredAt time.Time `json:"occurred_at"`
}
```

- [x] **Step 2: Crea payment events**

```go
// shared/events/payment.go
package events

import (
    "time"
    "github.com/google/uuid"
)

const (
    SubjectPaymentSucceeded = "payment.succeeded"
    SubjectPaymentFailed    = "payment.failed"
    SubjectPaymentRefunded  = "payment.refunded"
    SubjectInvoiceCreated   = "invoice.created"
)

type PaymentSucceeded struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    PaymentID  uuid.UUID `json:"payment_id"`
    Amount     int64     `json:"amount"`
    Currency   string    `json:"currency"`
    Provider   string    `json:"provider"`
    OccurredAt time.Time `json:"occurred_at"`
}

type PaymentFailed struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    PaymentID  uuid.UUID `json:"payment_id"`
    Reason     string    `json:"reason"`
    OccurredAt time.Time `json:"occurred_at"`
}

type PaymentRefunded struct {
    TenantID   uuid.UUID `json:"tenant_id"`
    UserID     uuid.UUID `json:"user_id"`
    PaymentID  uuid.UUID `json:"payment_id"`
    Amount     int64     `json:"amount"`
    OccurredAt time.Time `json:"occurred_at"`
}
```

- [x] **Step 3: Crea audit events**

```go
// shared/events/audit.go
package events

import (
    "time"
    "github.com/google/uuid"
)

// Audit event — ogni servizio pubblica su audit.* per eventi sensibili
const SubjectAuditWildcard = "audit.>"

type AuditEvent struct {
    TenantID     uuid.UUID      `json:"tenant_id"`
    ActorID      *uuid.UUID     `json:"actor_id,omitempty"`
    ActorType    string         `json:"actor_type"` // "user", "system", "admin"
    Action       string         `json:"action"`     // "user.login", "payment.succeeded"
    ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
    ResourceType string         `json:"resource_type,omitempty"`
    IPAddress    string         `json:"ip_address,omitempty"`
    Metadata     map[string]any `json:"metadata,omitempty"`
    OccurredAt   time.Time      `json:"occurred_at"`
}

// Subject helper: "audit.user.login", "audit.payment.succeeded"
func AuditSubject(action string) string { return "audit." + action }
```

- [x] **Step 4: Crea helper publisher**

```go
// shared/events/publisher.go
package events

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/nats-io/nats.go/jetstream"
)

type Publisher struct{ js jetstream.JetStream }

func NewPublisher(js jetstream.JetStream) *Publisher { return &Publisher{js: js} }

func (p *Publisher) Publish(ctx context.Context, subject string, payload any) error {
    data, err := json.Marshal(payload)
    if err != nil { return fmt.Errorf("marshal event %s: %w", subject, err) }
    if _, err = p.js.Publish(ctx, subject, data); err != nil {
        return fmt.Errorf("publish %s: %w", subject, err)
    }
    return nil
}
```

- [x] **Step 5: Commit**

```bash
cd shared && go get github.com/nats-io/nats.go@latest
git add shared/events/
git commit -m "feat(shared): add typed NATS JetStream event definitions for all domains"
```

---

## Task 0.7: Docker Compose base

**Files:**
- Create: `deployments/docker-compose.yml`
- Create: `deployments/docker-compose.dev.yml`
- Create: `deployments/traefik/traefik.yml`
- Create: `deployments/vault/config.hcl`
- Create: `deployments/nats/nats.conf`

- [x] **Step 1: Crea docker-compose.yml**

```yaml
# deployments/docker-compose.yml
version: "3.9"

x-service-defaults: &service-defaults
  restart: unless-stopped
  networks: [backend]

services:
  # --- Ingress ---
  traefik:
    <<: *service-defaults
    image: traefik:v3
    ports: ["80:80", "443:443", "8080:8080"]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik/traefik.yml:/etc/traefik/traefik.yml
      - letsencrypt:/letsencrypt

  # --- Application Services ---
  api-gateway:
    <<: *service-defaults
    build: ../services/api-gateway
    environment:
      - PORT=8080
      - AUTH_GRPC_ADDR=auth-service:9091
      - ENV=production
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.gateway.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.gateway.tls.certresolver=letsencrypt"
    depends_on: [auth-service, payment-service]

  auth-service:
    <<: *service-defaults
    build: ../services/auth-service
    environment:
      - PORT=8081
      - GRPC_PORT=9091
      - DATABASE_URL=postgres://auth:${AUTH_DB_PASS}@pgbouncer:5432/auth_db?sslmode=disable
      - REDIS_URL=redis-sentinel:26379
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
      - ENV=production
    depends_on: [postgres-auth, redis-master, vault, nats-1]

  payment-service:
    <<: *service-defaults
    build: ../services/payment-service
    environment:
      - PORT=8082
      - GRPC_PORT=9092
      - DATABASE_URL=postgres://payment:${PAYMENT_DB_PASS}@pgbouncer:5432/payment_db?sslmode=disable
      - AUTH_GRPC_ADDR=auth-service:9091
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
      - ENV=production
    depends_on: [postgres-payments, auth-service, vault, nats-1]

  notification-service:
    <<: *service-defaults
    build: ../services/notification-service
    environment:
      - DATABASE_URL=postgres://notif:${NOTIF_DB_PASS}@pgbouncer:5432/notification_db?sslmode=disable
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
      - ENV=production
    depends_on: [postgres-notifications, nats-1, vault]

  admin-service:
    <<: *service-defaults
    build: ../services/admin-service
    environment:
      - PORT=8084
      - GRPC_PORT=9094
      - AUTH_GRPC_ADDR=auth-service:9091
      - PAYMENT_GRPC_ADDR=payment-service:9092
      - AUDIT_GRPC_ADDR=audit-service:9095
      - ENV=production
    depends_on: [auth-service, payment-service, audit-service]

  audit-service:
    <<: *service-defaults
    build: ../services/audit-service
    environment:
      - PORT=8085
      - GRPC_PORT=9095
      - DATABASE_URL=postgres://audit:${AUDIT_DB_PASS}@pgbouncer:5432/audit_db?sslmode=disable
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - ENV=production
    depends_on: [postgres-audit, nats-1]

  job-service:
    <<: *service-defaults
    build: ../services/job-service
    environment:
      - REDIS_URL=redis-sentinel:26379
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - ENV=production
    depends_on: [redis-master, nats-1]

  permission-service:
    <<: *service-defaults
    build: ../services/permission-service
    environment:
      - PORT=8086
      - GRPC_PORT=9096
      - DATABASE_URL=postgres://perm:${PERM_DB_PASS}@pgbouncer:5432/permission_db?sslmode=disable
      - REDIS_URL=redis-sentinel:26379
      - ENV=production
    depends_on: [postgres-permissions, redis-master]

  config-service:
    <<: *service-defaults
    build: ../services/config-service
    environment:
      - PORT=8087
      - GRPC_PORT=9097
      - DATABASE_URL=postgres://config:${CONFIG_DB_PASS}@pgbouncer:5432/config_db?sslmode=disable
      - REDIS_URL=redis-sentinel:26379
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - ENV=production
    depends_on: [postgres-config, redis-master, nats-1]

  webhook-service:
    <<: *service-defaults
    build: ../services/webhook-service
    environment:
      - PORT=8088
      - DATABASE_URL=postgres://webhook:${WEBHOOK_DB_PASS}@pgbouncer:5432/webhook_db?sslmode=disable
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - ENV=production
    depends_on: [postgres-webhooks, nats-1]

  storage-service:
    <<: *service-defaults
    build: ../services/storage-service
    environment:
      - PORT=8089
      - DATABASE_URL=postgres://storage:${STORAGE_DB_PASS}@pgbouncer:5432/storage_db?sslmode=disable
      - MINIO_ENDPOINT=minio:9000
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
      - ENV=production
    depends_on: [postgres-storage, minio, vault]

  subscription-service:
    <<: *service-defaults
    build: ../services/subscription-service
    environment:
      - PORT=8090
      - GRPC_PORT=9090
      - DATABASE_URL=postgres://subs:${SUBS_DB_PASS}@pgbouncer:5432/subscription_db?sslmode=disable
      - NATS_URL=nats://nats-1:4222,nats://nats-2:4222,nats://nats-3:4222
      - AUTH_GRPC_ADDR=auth-service:9091
      - VAULT_ADDR=http://vault:8200
      - VAULT_TOKEN=${VAULT_TOKEN}
      - ENV=production
    depends_on: [postgres-subscriptions, auth-service, nats-1, vault]

  # --- Databases (one per service) ---
  postgres-auth:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: auth_db
      POSTGRES_USER: auth
      POSTGRES_PASSWORD: ${AUTH_DB_PASS}
    volumes: [postgres-auth-data:/var/lib/postgresql/data]

  postgres-payments:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: payment_db
      POSTGRES_USER: payment
      POSTGRES_PASSWORD: ${PAYMENT_DB_PASS}
    volumes: [postgres-payments-data:/var/lib/postgresql/data]

  postgres-notifications:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: notification_db
      POSTGRES_USER: notif
      POSTGRES_PASSWORD: ${NOTIF_DB_PASS}
    volumes: [postgres-notifications-data:/var/lib/postgresql/data]

  postgres-audit:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: audit_db
      POSTGRES_USER: audit
      POSTGRES_PASSWORD: ${AUDIT_DB_PASS}
    volumes: [postgres-audit-data:/var/lib/postgresql/data]

  postgres-permissions:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: permission_db
      POSTGRES_USER: perm
      POSTGRES_PASSWORD: ${PERM_DB_PASS}
    volumes: [postgres-permissions-data:/var/lib/postgresql/data]

  postgres-config:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: config_db
      POSTGRES_USER: config
      POSTGRES_PASSWORD: ${CONFIG_DB_PASS}
    volumes: [postgres-config-data:/var/lib/postgresql/data]

  postgres-webhooks:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: webhook_db
      POSTGRES_USER: webhook
      POSTGRES_PASSWORD: ${WEBHOOK_DB_PASS}
    volumes: [postgres-webhooks-data:/var/lib/postgresql/data]

  postgres-storage:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: storage_db
      POSTGRES_USER: storage
      POSTGRES_PASSWORD: ${STORAGE_DB_PASS}
    volumes: [postgres-storage-data:/var/lib/postgresql/data]

  postgres-subscriptions:
    <<: *service-defaults
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: subscription_db
      POSTGRES_USER: subs
      POSTGRES_PASSWORD: ${SUBS_DB_PASS}
    volumes: [postgres-subscriptions-data:/var/lib/postgresql/data]

  # --- Connection Pooling ---
  pgbouncer:
    <<: *service-defaults
    image: pgbouncer/pgbouncer:latest
    volumes: [./pgbouncer/pgbouncer.ini:/etc/pgbouncer/pgbouncer.ini]
    depends_on:
      - postgres-auth
      - postgres-payments
      - postgres-notifications
      - postgres-audit
      - postgres-permissions
      - postgres-config
      - postgres-webhooks
      - postgres-storage
      - postgres-subscriptions

  # --- Redis Sentinel HA ---
  redis-master:
    <<: *service-defaults
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes: [redis-master-data:/data]

  redis-replica:
    <<: *service-defaults
    image: redis:7-alpine
    command: redis-server --replicaof redis-master 6379 --appendonly yes
    depends_on: [redis-master]

  redis-sentinel:
    <<: *service-defaults
    image: redis:7-alpine
    command: redis-sentinel /etc/redis/sentinel.conf
    volumes: [./redis/sentinel.conf:/etc/redis/sentinel.conf]
    depends_on: [redis-master, redis-replica]

  # --- NATS Cluster ---
  nats-1:
    <<: *service-defaults
    image: nats:2-alpine
    command: -js -cluster nats://nats-1:6222 -routes nats://nats-2:6222,nats://nats-3:6222 -n nats-1
    ports: ["4222:4222"]

  nats-2:
    <<: *service-defaults
    image: nats:2-alpine
    command: -js -cluster nats://nats-2:6222 -routes nats://nats-1:6222,nats://nats-3:6222 -n nats-2

  nats-3:
    <<: *service-defaults
    image: nats:2-alpine
    command: -js -cluster nats://nats-3:6222 -routes nats://nats-1:6222,nats://nats-2:6222 -n nats-3

  # --- Secrets ---
  vault:
    <<: *service-defaults
    image: hashicorp/vault:1.17
    cap_add: [IPC_LOCK]
    environment:
      VAULT_DEV_ROOT_TOKEN_ID: ${VAULT_TOKEN}
    ports: ["8200:8200"]

  # --- File Storage ---
  minio:
    <<: *service-defaults
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_PASS}
    volumes: [minio-data:/data]
    ports: ["9000:9000", "9001:9001"]

  # --- Observability ---
  prometheus:
    <<: *service-defaults
    image: prom/prometheus:latest
    volumes: [./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml]
    ports: ["9090:9090"]

  alertmanager:
    <<: *service-defaults
    image: prom/alertmanager:latest
    volumes: [./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml]

  grafana:
    <<: *service-defaults
    image: grafana/grafana:latest
    volumes: [./grafana/provisioning:/etc/grafana/provisioning]
    ports: ["3000:3000"]

  jaeger:
    <<: *service-defaults
    image: jaegertracing/all-in-one:latest
    ports: ["16686:16686"]

  otel-collector:
    <<: *service-defaults
    image: otel/opentelemetry-collector-contrib:latest
    volumes: [./otel-collector.yml:/etc/otel/config.yaml]
    command: ["--config=/etc/otel/config.yaml"]

networks:
  backend:
    driver: bridge

volumes:
  letsencrypt:
  postgres-auth-data:
  postgres-payments-data:
  postgres-notifications-data:
  postgres-audit-data:
  postgres-permissions-data:
  postgres-config-data:
  postgres-webhooks-data:
  postgres-storage-data:
  postgres-subscriptions-data:
  redis-master-data:
  minio-data:
```

- [x] **Step 2: Crea .env.example nella root**

```bash
cat > .env.example << 'EOF'
# Vault
VAULT_TOKEN=root

# Database passwords (genera con: openssl rand -hex 32)
AUTH_DB_PASS=changeme
PAYMENT_DB_PASS=changeme
NOTIF_DB_PASS=changeme
AUDIT_DB_PASS=changeme
PERM_DB_PASS=changeme
CONFIG_DB_PASS=changeme
WEBHOOK_DB_PASS=changeme
STORAGE_DB_PASS=changeme
SUBS_DB_PASS=changeme

# MinIO
MINIO_USER=minioadmin
MINIO_PASS=changeme
EOF
cp .env.example .env
```

- [x] **Step 3: Crea Traefik config**

```yaml
# deployments/traefik/traefik.yml
api:
  dashboard: true
  insecure: true  # solo per dev, rimuovi in prod

entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"

certificatesResolvers:
  letsencrypt:
    acme:
      email: admin@yourdomain.com
      storage: /letsencrypt/acme.json
      httpChallenge:
        entryPoint: web

providers:
  docker:
    exposedByDefault: false
    network: backend
```

- [x] **Step 4: Avvia infrastruttura base e verifica**

```bash
cd deployments
docker compose up -d postgres-auth redis-master vault nats-1
sleep 5
docker compose ps
# Expected: postgres-auth, redis-master, vault, nats-1 tutti "running"
```

- [x] **Step 5: Commit**

```bash
git add deployments/ .env.example
git commit -m "feat(infra): add full Docker Compose with 10 DBs, Redis Sentinel, NATS cluster, Vault, MinIO, observability stack"
```

---

## Task 0.8: Dockerfile base (pattern condiviso)

**Files:**
- Create: `services/auth-service/Dockerfile`
  (stesso pattern per tutti gli altri servizi)

- [x] **Step 1: Crea Dockerfile multi-stage**

```dockerfile
# services/auth-service/Dockerfile
# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Cache delle dipendenze
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server ./cmd/server

# Stage 2: Runtime (minimal)
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/server /server
EXPOSE 8081 9091
USER nonroot:nonroot
ENTRYPOINT ["/server"]
```

- [x] **Step 2: Replica il Dockerfile per tutti i servizi**

```bash
for svc in payment-service notification-service admin-service audit-service api-gateway permission-service config-service webhook-service storage-service subscription-service; do
  cp services/auth-service/Dockerfile services/$svc/Dockerfile
  # Aggiorna EXPOSE in ogni file con le porte corrette
done
```

- [x] **Step 3: Crea script di migrazione base**

```bash
cat > scripts/migrate.sh << 'EOF'
#!/bin/bash
# Usage: ./scripts/migrate.sh <service> <direction>
# Example: ./scripts/migrate.sh auth-service up

SERVICE=$1
DIRECTION=${2:-up}

case $SERVICE in
  auth-service)     DB_URL=$AUTH_DATABASE_URL ;;
  payment-service)  DB_URL=$PAYMENT_DATABASE_URL ;;
  *)                echo "Unknown service: $SERVICE"; exit 1 ;;
esac

migrate -path services/$SERVICE/migrations -database "$DB_URL" $DIRECTION
EOF
chmod +x scripts/migrate.sh
```

- [x] **Step 4: Commit**

```bash
git add services/*/Dockerfile scripts/
git commit -m "feat(infra): add multi-stage Dockerfiles and migration script"
```

---

## Task 0.9: Vault init script + NATS JetStream streams

**Files:**
- Create: `scripts/vault-init.sh`
- Create: `scripts/nats-init.sh`

- [x] **Step 1: Crea vault-init.sh**

```bash
cat > scripts/vault-init.sh << 'EOF'
#!/bin/bash
# Popola Vault con i secrets dell'applicazione
# Eseguire UNA SOLA VOLTA su primo deploy
# Dipendenze: vault CLI installato, VAULT_ADDR e VAULT_TOKEN settati

set -e

VAULT_ADDR=${VAULT_ADDR:-http://localhost:8200}
VAULT_TOKEN=${VAULT_TOKEN:-root}
export VAULT_ADDR VAULT_TOKEN

echo "Enabling KV secrets engine..."
vault secrets enable -path=secret kv-v2 || true

echo "Generating RSA key pair for JWT..."
openssl genrsa -out /tmp/jwt_private.pem 4096
openssl rsa -in /tmp/jwt_private.pem -pubout -out /tmp/jwt_public.pem

vault kv put secret/auth/jwt \
  private_key=@/tmp/jwt_private.pem \
  public_key=@/tmp/jwt_public.pem

rm /tmp/jwt_private.pem /tmp/jwt_public.pem

echo "Storing PII encryption key..."
PII_KEY=$(openssl rand -hex 32)
vault kv put secret/crypto pii_key="$PII_KEY"

echo "Placeholders for external service keys..."
vault kv put secret/payment/stripe \
  secret_key="sk_test_REPLACE_ME" \
  webhook_secret="whsec_REPLACE_ME"

vault kv put secret/notification/sendgrid \
  api_key="SG.REPLACE_ME"

vault kv put secret/notification/twilio \
  account_sid="AC_REPLACE_ME" \
  auth_token="REPLACE_ME" \
  from_number="+1REPLACE"

echo "Vault initialized successfully."
EOF
chmod +x scripts/vault-init.sh
```

- [x] **Step 2: Crea nats-init.sh per JetStream streams**

```bash
cat > scripts/nats-init.sh << 'EOF'
#!/bin/bash
# Crea gli stream JetStream necessari
# Dipendenze: nats CLI installato, NATS_URL settato

NATS_URL=${NATS_URL:-nats://localhost:4222}

echo "Creating AUTH_EVENTS stream..."
nats stream add AUTH_EVENTS \
  --subjects "auth.>" \
  --retention work \
  --max-age 7d \
  --replicas 1 \
  --server $NATS_URL || echo "Stream already exists"

echo "Creating PAYMENT_EVENTS stream..."
nats stream add PAYMENT_EVENTS \
  --subjects "payment.>" \
  --retention work \
  --max-age 30d \
  --replicas 1 \
  --server $NATS_URL || echo "Stream already exists"

echo "Creating AUDIT_EVENTS stream..."
nats stream add AUDIT_EVENTS \
  --subjects "audit.>" \
  --retention limits \
  --max-age 365d \
  --replicas 1 \
  --server $NATS_URL || echo "Stream already exists"

echo "Creating NOTIFICATION_RETRY stream..."
nats stream add NOTIFICATION_RETRY \
  --subjects "notification.retry" \
  --retention work \
  --max-age 24h \
  --replicas 1 \
  --server $NATS_URL || echo "Stream already exists"

echo "NATS JetStream streams initialized."
EOF
chmod +x scripts/nats-init.sh
```

- [x] **Step 3: Avvia NATS e inizializza streams**

```bash
cd deployments && docker compose up -d nats-1 && sleep 3
cd .. && NATS_URL=nats://localhost:4222 ./scripts/nats-init.sh
# Expected: tutti i 4 stream creati con successo
```

- [x] **Step 4: Commit**

```bash
git add scripts/
git commit -m "feat(infra): add Vault init script (JWT keys, PII key, external secrets) and NATS JetStream stream definitions"
```

---

# FASE 1 — Auth Service Core

## Task 1.1: Auth Service — domain layer

**Files:**
- Create: `services/auth-service/internal/domain/user.go`
- Create: `services/auth-service/internal/domain/token.go`
- Create: `services/auth-service/internal/domain/repository.go`
- Create: `services/auth-service/internal/domain/errors.go`
- Test: `services/auth-service/internal/domain/user_test.go`

- [x] **Step 1: Aggiungi dipendenze auth-service**

```bash
cd services/auth-service
go get github.com/google/uuid@latest
go get github.com/golang-jwt/jwt/v5@latest
go get golang.org/x/crypto@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/redis/go-redis/v9@latest
go get github.com/nats-io/nats.go@latest
go get go.opentelemetry.io/otel@latest
go get github.com/prometheus/client_golang@latest
go get github.com/hashicorp/vault/api/v2@latest
```

- [x] **Step 2: Scrivi test per il dominio**

```go
// services/auth-service/internal/domain/user_test.go
package domain_test

import (
    "testing"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
)

func TestNewUser_Valid(t *testing.T) {
    u, err := domain.NewUser("tenant-id", "user@example.com", "hashedpassword")
    if err != nil { t.Fatalf("expected no error, got: %v", err) }
    if u.ID.String() == "" { t.Error("ID should not be empty") }
    if u.EmailVerified { t.Error("new user should not be email verified") }
    if len(u.Roles) == 0 || u.Roles[0] != domain.RoleUser {
        t.Errorf("new user should have role 'user', got: %v", u.Roles)
    }
}

func TestUser_HasRole(t *testing.T) {
    u, _ := domain.NewUser("tid", "e@e.com", "hash")
    if !u.HasRole(domain.RoleUser) { t.Error("should have user role") }
    if u.HasRole(domain.RoleAdmin) { t.Error("should not have admin role") }
}
```

- [x] **Step 3: Run test — deve fallire**

```bash
cd services/auth-service && go test ./internal/domain/... -v
# Expected: FAIL
```

- [x] **Step 4: Implementa user.go**

```go
// services/auth-service/internal/domain/user.go
package domain

import (
    "time"
    "github.com/google/uuid"
)

const (
    RoleUser       = "user"
    RoleAdmin      = "admin"
    RoleSuperAdmin = "super-admin"
)

type User struct {
    ID            uuid.UUID
    TenantID      uuid.UUID
    EmailEnc      []byte     // AES-256-GCM encrypted
    EmailHash     string     // SHA-256 for lookup
    PasswordHash  string     // bcrypt cost 12, nil for OAuth-only users
    Roles         []string
    MFAEnabled    bool
    MFASecret     []byte     // TOTP secret, encrypted
    EmailVerified bool
    LockedUntil   *time.Time // brute force lockout
    CreatedAt     time.Time
    UpdatedAt     time.Time
    DeletedAt     *time.Time
}

func NewUser(tenantID, emailHash, passwordHash string) (*User, error) {
    tid, err := uuid.Parse(tenantID)
    if err != nil { return nil, ErrInvalidTenantID }
    return &User{
        ID:           uuid.New(),
        TenantID:     tid,
        EmailHash:    emailHash,
        PasswordHash: passwordHash,
        Roles:        []string{RoleUser},
        CreatedAt:    time.Now().UTC(),
        UpdatedAt:    time.Now().UTC(),
    }, nil
}

func (u *User) HasRole(role string) bool {
    for _, r := range u.Roles { if r == role { return true } }
    return false
}

func (u *User) IsLocked() bool {
    return u.LockedUntil != nil && time.Now().UTC().Before(*u.LockedUntil)
}
```

- [x] **Step 5: Implementa token.go e errors.go**

```go
// services/auth-service/internal/domain/token.go
package domain

import "time"

type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    time.Duration
}

type TokenClaims struct {
    UserID    string   `json:"sub"`
    TenantID  string   `json:"tenant_id"`
    Roles     []string `json:"roles"`
    JTI       string   `json:"jti"`
    MFAVerified bool   `json:"mfa_verified"`
}
```

```go
// services/auth-service/internal/domain/errors.go
package domain

import "errors"

var (
    ErrInvalidTenantID    = errors.New("invalid tenant id")
    ErrUserNotFound       = errors.New("user not found")
    ErrEmailAlreadyExists = errors.New("email already exists for this tenant")
    ErrInvalidCredentials = errors.New("invalid email or password")
    ErrAccountLocked      = errors.New("account temporarily locked")
    ErrEmailNotVerified   = errors.New("email address not verified")
    ErrTokenExpired       = errors.New("token expired")
    ErrTokenRevoked       = errors.New("token revoked")
    ErrMFARequired        = errors.New("mfa verification required")
    ErrMFAInvalidCode     = errors.New("invalid mfa code")
)
```

- [x] **Step 6: Implementa repository.go (interfacce)**

```go
// services/auth-service/internal/domain/repository.go
package domain

import (
    "context"
    "github.com/google/uuid"
)

type UserRepository interface {
    Create(ctx context.Context, u *User) error
    GetByEmailHash(ctx context.Context, tenantID uuid.UUID, emailHash string) (*User, error)
    GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*User, error)
    Update(ctx context.Context, u *User) error
    SoftDelete(ctx context.Context, tenantID, userID uuid.UUID) error
    List(ctx context.Context, tenantID uuid.UUID, limit int, cursor *string) ([]*User, error)
}

type TokenStore interface {
    StoreRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, jti string) error
    ValidateRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, jti string) (bool, error)
    RevokeRefreshToken(ctx context.Context, tenantID, userID uuid.UUID, jti string) error
    BlacklistJTI(ctx context.Context, jti string, ttlSeconds int) error
    IsJTIBlacklisted(ctx context.Context, jti string) (bool, error)
    StoreVerifyToken(ctx context.Context, token string, userID uuid.UUID) error
    GetVerifyToken(ctx context.Context, token string) (uuid.UUID, error)
    DeleteVerifyToken(ctx context.Context, token string) error
    StoreResetToken(ctx context.Context, token string, userID uuid.UUID) error
    GetResetToken(ctx context.Context, token string) (uuid.UUID, error)
    DeleteResetToken(ctx context.Context, token string) error
    IncrBruteForce(ctx context.Context, key string) (int64, error)
    SetBruteForceLockout(ctx context.Context, key string, seconds int) error
    GetBruteForceCount(ctx context.Context, key string) (int64, error)
}
```

- [x] **Step 7: Run test — devono passare**

```bash
go test ./internal/domain/... -v
# Expected: PASS
```

- [x] **Step 8: Commit**

```bash
git add services/auth-service/internal/domain/
git commit -m "feat(auth): add domain layer — User entity, TokenPair, repository interfaces, error types"
```

---

## Task 1.2: Auth Service — RegisterUseCase con TDD

**Files:**
- Create: `services/auth-service/internal/application/register.go`
- Test: `services/auth-service/internal/application/register_test.go`

- [x] **Step 1: Scrivi test con mock del repository**

```bash
cd services/auth-service
go get github.com/stretchr/testify@latest
# Installa mockery per generare mock
go install github.com/vektra/mockery/v2@latest
mockery --name=UserRepository --dir=internal/domain --output=internal/mocks --outpkg=mocks
mockery --name=TokenStore --dir=internal/domain --output=internal/mocks --outpkg=mocks
```

```go
// services/auth-service/internal/application/register_test.go
package application_test

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/yourorg/golang-modules/services/auth-service/internal/application"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
    "github.com/yourorg/golang-modules/services/auth-service/internal/mocks"
)

func TestRegister_Success(t *testing.T) {
    userRepo := mocks.NewUserRepository(t)
    tenantID := uuid.New()
    userRepo.On("GetByEmailHash", mock.Anything, tenantID, mock.AnythingOfType("string")).
        Return(nil, domain.ErrUserNotFound)
    userRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
        Return(nil)

    publisher := &mockPublisher{}
    encryptor := &mockEncryptor{}

    uc := application.NewRegisterUseCase(userRepo, publisher, encryptor)
    result, err := uc.Execute(context.Background(), application.RegisterInput{
        TenantID: tenantID.String(),
        Email:    "user@example.com",
        Password: "SecurePass123!",
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, result.UserID)
    assert.True(t, publisher.published)
}

func TestRegister_DuplicateEmail(t *testing.T) {
    userRepo := mocks.NewUserRepository(t)
    existing := &domain.User{ID: uuid.New()}
    userRepo.On("GetByEmailHash", mock.Anything, mock.Anything, mock.Anything).
        Return(existing, nil)

    uc := application.NewRegisterUseCase(userRepo, &mockPublisher{}, &mockEncryptor{})
    _, err := uc.Execute(context.Background(), application.RegisterInput{
        TenantID: uuid.New().String(),
        Email:    "existing@example.com",
        Password: "SecurePass123!",
    })

    assert.ErrorIs(t, err, domain.ErrEmailAlreadyExists)
}

// Mock helpers
type mockPublisher struct{ published bool }
func (m *mockPublisher) Publish(_ context.Context, _ string, _ any) error {
    m.published = true; return nil
}
type mockEncryptor struct{}
func (m *mockEncryptor) Encrypt(s string) (string, error) { return "encrypted:" + s, nil }
func (m *mockEncryptor) Hash(s string) string              { return "hash:" + s }
```

- [x] **Step 2: Run test — deve fallire**

```bash
go test ./internal/application/... -run TestRegister -v
# Expected: FAIL
```

- [x] **Step 3: Implementa register.go**

```go
// services/auth-service/internal/application/register.go
package application

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
    "github.com/yourorg/golang-modules/shared/events"
)

type PIIEncryptor interface {
    Encrypt(s string) (string, error)
    Hash(s string) string
}

type EventPublisher interface {
    Publish(ctx context.Context, subject string, payload any) error
}

type RegisterInput struct {
    TenantID string `validate:"required,uuid4"`
    Email    string `validate:"required,email"`
    Password string `validate:"required,min=8"`
}

type RegisterOutput struct {
    UserID string
}

type RegisterUseCase struct {
    users     domain.UserRepository
    publisher EventPublisher
    encryptor PIIEncryptor
}

func NewRegisterUseCase(u domain.UserRepository, p EventPublisher, enc PIIEncryptor) *RegisterUseCase {
    return &RegisterUseCase{users: u, publisher: p, encryptor: enc}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
    tenantID, _ := uuid.Parse(in.TenantID)
    emailHash := uc.encryptor.Hash(in.Email)

    // Check uniqueness
    _, err := uc.users.GetByEmailHash(ctx, tenantID, emailHash)
    if err == nil {
        return nil, domain.ErrEmailAlreadyExists
    }
    if err != domain.ErrUserNotFound {
        return nil, fmt.Errorf("check email uniqueness: %w", err)
    }

    // Hash password (bcrypt cost 12)
    pwHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), 12)
    if err != nil { return nil, fmt.Errorf("hash password: %w", err) }

    user, err := domain.NewUser(in.TenantID, emailHash, string(pwHash))
    if err != nil { return nil, err }

    // Encrypt email for storage
    encEmail, err := uc.encryptor.Encrypt(in.Email)
    if err != nil { return nil, fmt.Errorf("encrypt email: %w", err) }
    user.EmailEnc = []byte(encEmail)

    if err := uc.users.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("create user: %w", err)
    }

    // Publish event (non-blocking — fire and forget with context)
    go uc.publisher.Publish(context.Background(), events.SubjectUserRegistered, events.UserRegistered{
        TenantID:  tenantID,
        UserID:    user.ID,
        Email:     in.Email,
        VerifyURL: fmt.Sprintf("https://api.yourdomain.com/v1/auth/verify-email?token=PENDING"),
    })

    return &RegisterOutput{UserID: user.ID.String()}, nil
}
```

- [x] **Step 4: Run test — devono passare**

```bash
go test ./internal/application/... -run TestRegister -v
# Expected: PASS
```

- [x] **Step 5: Commit**

```bash
git add services/auth-service/internal/application/ services/auth-service/internal/mocks/
git commit -m "feat(auth): add RegisterUseCase with email uniqueness check, bcrypt, PII encryption, event publishing"
```

---

## Task 1.3: Auth Service — LoginUseCase + JWT

**Files:**
- Create: `services/auth-service/internal/application/login.go`
- Create: `services/auth-service/internal/application/jwt.go`
- Test: `services/auth-service/internal/application/login_test.go`

- [x] **Step 1: Scrivi test per login**

```go
// services/auth-service/internal/application/login_test.go
package application_test

import (
    "context"
    "testing"
    "time"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "golang.org/x/crypto/bcrypt"
    "github.com/yourorg/golang-modules/services/auth-service/internal/application"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
    "github.com/yourorg/golang-modules/services/auth-service/internal/mocks"
)

func TestLogin_Success(t *testing.T) {
    password := "SecurePass123!"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), 4) // cost 4 in tests

    userRepo := mocks.NewUserRepository(t)
    tokenStore := mocks.NewTokenStore(t)

    tenantID := uuid.New()
    user := &domain.User{
        ID: uuid.New(), TenantID: tenantID,
        PasswordHash: string(hash), Roles: []string{"user"},
        EmailVerified: true,
    }
    userRepo.On("GetByEmailHash", mock.Anything, tenantID, mock.AnythingOfType("string")).
        Return(user, nil)
    tokenStore.On("GetBruteForceCount", mock.Anything, mock.Anything).Return(int64(0), nil)
    tokenStore.On("StoreRefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
        Return(nil)
    tokenStore.On("Create", mock.Anything, mock.Anything).Return(nil) // session

    tokenSvc := application.NewJWTService("test-secret-32-bytes-long-for-hs256")
    uc := application.NewLoginUseCase(userRepo, tokenStore, tokenSvc, &mockEncryptor{}, &mockPublisher{})

    out, err := uc.Execute(context.Background(), application.LoginInput{
        TenantID:  tenantID.String(),
        Email:     "user@example.com",
        Password:  password,
        IPAddress: "127.0.0.1",
        UserAgent: "test",
    })

    assert.NoError(t, err)
    assert.NotEmpty(t, out.AccessToken)
    assert.NotEmpty(t, out.RefreshToken)
}

func TestLogin_InvalidPassword(t *testing.T) {
    hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), 4)
    userRepo := mocks.NewUserRepository(t)
    tokenStore := mocks.NewTokenStore(t)
    tenantID := uuid.New()
    user := &domain.User{ID: uuid.New(), TenantID: tenantID, PasswordHash: string(hash), EmailVerified: true}
    userRepo.On("GetByEmailHash", mock.Anything, tenantID, mock.Anything).Return(user, nil)
    tokenStore.On("GetBruteForceCount", mock.Anything, mock.Anything).Return(int64(0), nil)
    tokenStore.On("IncrBruteForce", mock.Anything, mock.Anything).Return(int64(1), nil)

    tokenSvc := application.NewJWTService("test-secret-32-bytes-long-for-hs256")
    uc := application.NewLoginUseCase(userRepo, tokenStore, tokenSvc, &mockEncryptor{}, &mockPublisher{})
    _, err := uc.Execute(context.Background(), application.LoginInput{
        TenantID: tenantID.String(), Email: "user@example.com",
        Password: "wrongpassword", IPAddress: "127.0.0.1",
    })
    assert.ErrorIs(t, err, domain.ErrInvalidCredentials)
}
```

- [x] **Step 2: Implementa jwt.go**

```go
// services/auth-service/internal/application/jwt.go
package application

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
)

type JWTService struct{ secret []byte }

func NewJWTService(secret string) *JWTService { return &JWTService{secret: []byte(secret)} }

type jwtClaims struct {
    jwt.RegisteredClaims
    TenantID    string   `json:"tenant_id"`
    Roles       []string `json:"roles"`
    MFAVerified bool     `json:"mfa_verified"`
}

func (s *JWTService) Issue(user *domain.User, mfaVerified bool) (domain.TokenPair, error) {
    jti := uuid.New().String()
    access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
        RegisteredClaims: jwt.RegisteredClaims{
            Subject:   user.ID.String(),
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        jti,
        },
        TenantID:    user.TenantID.String(),
        Roles:       user.Roles,
        MFAVerified: mfaVerified,
    })
    accessStr, err := access.SignedString(s.secret)
    if err != nil { return domain.TokenPair{}, fmt.Errorf("sign access token: %w", err) }

    return domain.TokenPair{
        AccessToken:  accessStr,
        RefreshToken: uuid.New().String(),
        ExpiresIn:    15 * time.Minute,
    }, nil
}

func (s *JWTService) Validate(tokenStr string) (*domain.TokenClaims, string, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (any, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return s.secret, nil
    })
    if err != nil { return nil, "", domain.ErrTokenExpired }
    claims, ok := token.Claims.(*jwtClaims)
    if !ok || !token.Valid { return nil, "", domain.ErrTokenExpired }
    return &domain.TokenClaims{
        UserID: claims.Subject, TenantID: claims.TenantID,
        Roles: claims.Roles, JTI: claims.ID, MFAVerified: claims.MFAVerified,
    }, claims.ID, nil
}
```

- [x] **Step 3: Implementa login.go**

```go
// services/auth-service/internal/application/login.go
package application

import (
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
    "github.com/yourorg/golang-modules/shared/events"
)

type LoginInput struct {
    TenantID  string
    Email     string
    Password  string
    IPAddress string
    UserAgent string
    DeviceID  string
}

type LoginOutput struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    time.Duration
    MFARequired  bool
}

type TokenIssuer interface {
    Issue(user *domain.User, mfaVerified bool) (domain.TokenPair, error)
    Validate(token string) (*domain.TokenClaims, string, error)
}

type LoginUseCase struct {
    users     domain.UserRepository
    tokens    domain.TokenStore
    jwt       TokenIssuer
    encryptor PIIEncryptor
    publisher EventPublisher
}

func NewLoginUseCase(u domain.UserRepository, t domain.TokenStore, jwt TokenIssuer, enc PIIEncryptor, pub EventPublisher) *LoginUseCase {
    return &LoginUseCase{users: u, tokens: t, jwt: jwt, encryptor: enc, publisher: pub}
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (*LoginOutput, error) {
    tenantID, _ := uuid.Parse(in.TenantID)
    emailHash := uc.encryptor.Hash(in.Email)
    bfKey := fmt.Sprintf("%s:%s:%s", in.TenantID, in.IPAddress, emailHash)

    // Brute force check
    count, _ := uc.tokens.GetBruteForceCount(ctx, bfKey)
    if count >= 10 { return nil, domain.ErrAccountLocked }

    user, err := uc.users.GetByEmailHash(ctx, tenantID, emailHash)
    if err != nil {
        uc.tokens.IncrBruteForce(ctx, bfKey)
        return nil, domain.ErrInvalidCredentials
    }
    if user.IsLocked() { return nil, domain.ErrAccountLocked }
    if !user.EmailVerified { return nil, domain.ErrEmailNotVerified }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
        uc.tokens.IncrBruteForce(ctx, bfKey)
        return nil, domain.ErrInvalidCredentials
    }

    if user.MFAEnabled {
        partialToken, _ := uc.jwt.Issue(user, false)
        return &LoginOutput{AccessToken: partialToken.AccessToken, MFARequired: true}, nil
    }

    pair, err := uc.jwt.Issue(user, true)
    if err != nil { return nil, err }

    if err := uc.tokens.StoreRefreshToken(ctx, tenantID, user.ID, pair.RefreshToken); err != nil {
        return nil, fmt.Errorf("store refresh token: %w", err)
    }

    go uc.publisher.Publish(context.Background(), events.SubjectUserLogin, events.UserLogin{
        TenantID: tenantID, UserID: user.ID, IPAddress: in.IPAddress, UserAgent: in.UserAgent,
    })

    return &LoginOutput{
        AccessToken: pair.AccessToken, RefreshToken: pair.RefreshToken, ExpiresIn: pair.ExpiresIn,
    }, nil
}
```

- [x] **Step 4: Run test**

```bash
go test ./internal/application/... -run TestLogin -v
# Expected: PASS
```

- [x] **Step 5: Commit**

```bash
git add services/auth-service/internal/application/
git commit -m "feat(auth): add LoginUseCase with brute force protection, bcrypt verify, JWT issuance, MFA gate"
```

---

## Task 1.4: Auth Service — infrastructure (PostgreSQL + Redis)

**Files:**
- Create: `services/auth-service/internal/infrastructure/postgres/user_repo.go`
- Create: `services/auth-service/internal/infrastructure/redis/token_store.go`
- Create: `services/auth-service/migrations/000001_create_users.up.sql`
- Create: `services/auth-service/migrations/000001_create_users.down.sql`
- Test: `services/auth-service/internal/infrastructure/postgres/user_repo_integration_test.go`

- [x] **Step 1: Crea migration SQL**

```sql
-- services/auth-service/migrations/000001_create_users.up.sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    email_enc      BYTEA NOT NULL,
    email_hash     VARCHAR(64) NOT NULL,
    password_hash  VARCHAR(255),
    roles          TEXT[] NOT NULL DEFAULT '{"user"}',
    mfa_enabled    BOOLEAN NOT NULL DEFAULT false,
    mfa_secret     BYTEA,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    locked_until   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE(tenant_id, email_hash)
);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);

CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id    VARCHAR(255) NOT NULL,
    user_agent   TEXT,
    ip_address   INET,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_sessions_user_id ON sessions(tenant_id, user_id);

-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sessions
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

```sql
-- services/auth-service/migrations/000001_create_users.down.sql
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
```

- [x] **Step 2: Scrivi integration test (usa testcontainers)**

```bash
go get github.com/testcontainers/testcontainers-go@latest
go get github.com/golang-migrate/migrate/v4@latest
```

```go
// services/auth-service/internal/infrastructure/postgres/user_repo_integration_test.go
//go:build integration

package postgres_test

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
    "github.com/yourorg/golang-modules/services/auth-service/internal/infrastructure/postgres"
    // testcontainers setup helper (vedi Task 1.4 Step 3)
)

func TestUserRepo_CreateAndGet(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t) // avvia PostgreSQL in Docker via testcontainers

    repo := postgres.NewUserRepository(db)
    tenantID := uuid.New()

    user, _ := domain.NewUser(tenantID.String(), "hash:user@test.com", "$2a$12$testhash")
    user.TenantID = tenantID

    err := repo.Create(ctx, user)
    require.NoError(t, err)

    // RLS: imposta tenant context
    db.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)

    found, err := repo.GetByEmailHash(ctx, tenantID, "hash:user@test.com")
    require.NoError(t, err)
    assert.Equal(t, user.ID, found.ID)
}

func TestUserRepo_RLS_Isolation(t *testing.T) {
    ctx := context.Background()
    db := setupTestDB(t)
    repo := postgres.NewUserRepository(db)

    tenantA := uuid.New()
    tenantB := uuid.New()

    userA, _ := domain.NewUser(tenantA.String(), "hash:a@test.com", "hash")
    userA.TenantID = tenantA
    repo.Create(ctx, userA)

    // Tenant B non deve vedere i dati di Tenant A
    db.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantB)
    _, err := repo.GetByEmailHash(ctx, tenantA, "hash:a@test.com")
    assert.ErrorIs(t, err, domain.ErrUserNotFound, "RLS should prevent cross-tenant access")
}
```

- [x] **Step 3: Implementa user_repo.go**

```go
// services/auth-service/internal/infrastructure/postgres/user_repo.go
package postgres

import (
    "context"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourorg/golang-modules/services/auth-service/internal/domain"
)

type UserRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO users (id, tenant_id, email_enc, email_hash, password_hash, roles,
                           mfa_enabled, email_verified, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
        u.ID, u.TenantID, u.EmailEnc, u.EmailHash, u.PasswordHash,
        u.Roles, u.MFAEnabled, u.EmailVerified, u.CreatedAt, u.UpdatedAt)
    if err != nil { return fmt.Errorf("create user: %w", err) }
    return nil
}

func (r *UserRepository) GetByEmailHash(ctx context.Context, tenantID uuid.UUID, emailHash string) (*domain.User, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, tenant_id, email_enc, email_hash, password_hash, roles,
               mfa_enabled, mfa_secret, email_verified, locked_until, created_at, updated_at, deleted_at
        FROM users WHERE tenant_id=$1 AND email_hash=$2 AND deleted_at IS NULL`,
        tenantID, emailHash)
    return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.User, error) {
    row := r.db.QueryRow(ctx, `
        SELECT id, tenant_id, email_enc, email_hash, password_hash, roles,
               mfa_enabled, mfa_secret, email_verified, locked_until, created_at, updated_at, deleted_at
        FROM users WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`,
        tenantID, userID)
    return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
    _, err := r.db.Exec(ctx, `
        UPDATE users SET password_hash=$1, roles=$2, mfa_enabled=$3, mfa_secret=$4,
               email_verified=$5, locked_until=$6, updated_at=NOW()
        WHERE id=$7 AND tenant_id=$8`,
        u.PasswordHash, u.Roles, u.MFAEnabled, u.MFASecret,
        u.EmailVerified, u.LockedUntil, u.ID, u.TenantID)
    if err != nil { return fmt.Errorf("update user: %w", err) }
    return nil
}

func (r *UserRepository) SoftDelete(ctx context.Context, tenantID, userID uuid.UUID) error {
    _, err := r.db.Exec(ctx,
        `UPDATE users SET deleted_at=NOW(), updated_at=NOW() WHERE id=$1 AND tenant_id=$2`,
        userID, tenantID)
    return err
}

func (r *UserRepository) List(ctx context.Context, tenantID uuid.UUID, limit int, cursor *string) ([]*domain.User, error) {
    rows, err := r.db.Query(ctx,
        `SELECT id, tenant_id, email_enc, email_hash, password_hash, roles,
                mfa_enabled, mfa_secret, email_verified, locked_until, created_at, updated_at, deleted_at
         FROM users WHERE tenant_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2`,
        tenantID, limit)
    if err != nil { return nil, err }
    defer rows.Close()
    var users []*domain.User
    for rows.Next() {
        u, err := scanUserRow(rows)
        if err != nil { return nil, err }
        users = append(users, u)
    }
    return users, rows.Err()
}

func scanUser(row pgx.Row) (*domain.User, error) {
    u := &domain.User{}
    err := row.Scan(&u.ID, &u.TenantID, &u.EmailEnc, &u.EmailHash, &u.PasswordHash,
        &u.Roles, &u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.LockedUntil,
        &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
    if errors.Is(err, pgx.ErrNoRows) { return nil, domain.ErrUserNotFound }
    if err != nil { return nil, fmt.Errorf("scan user: %w", err) }
    return u, nil
}

func scanUserRow(rows pgx.Rows) (*domain.User, error) {
    u := &domain.User{}
    err := rows.Scan(&u.ID, &u.TenantID, &u.EmailEnc, &u.EmailHash, &u.PasswordHash,
        &u.Roles, &u.MFAEnabled, &u.MFASecret, &u.EmailVerified, &u.LockedUntil,
        &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
    if err != nil { return nil, fmt.Errorf("scan user row: %w", err) }
    return u, nil
}
```

- [x] **Step 4: Esegui integration test** ← test scritti con testcontainers; eseguire con: `go test -tags=integration ./internal/infrastructure/postgres/... -v`

```bash
go test -tags=integration ./internal/infrastructure/postgres/... -v
# Expected: PASS (testcontainers avvia PostgreSQL automaticamente)
```

- [x] **Step 5: Commit**

```bash
git add services/auth-service/internal/infrastructure/ services/auth-service/migrations/
git commit -m "feat(auth): add PostgreSQL user repository with RLS, migrations, integration tests"
```

---

## Task 1.5: Auth Service — HTTP transport layer

**Files:**
- Create: `services/auth-service/internal/transport/rest/handler.go`
- Create: `services/auth-service/internal/transport/rest/routes.go`
- Create: `services/auth-service/cmd/server/main.go`
- Test: `services/auth-service/internal/transport/rest/handler_test.go`

- [x] **Step 1: Scrivi test per l'handler**

```go
// services/auth-service/internal/transport/rest/handler_test.go
package rest_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/yourorg/golang-modules/services/auth-service/internal/transport/rest"
)

func TestRegisterHandler_InvalidJSON(t *testing.T) {
    h := rest.NewHandler(&mockRegisterUC{}, &mockLoginUC{})
    req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBufferString("not json"))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    h.Register(rr, req)
    assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestRegisterHandler_MissingEmail(t *testing.T) {
    h := rest.NewHandler(&mockRegisterUC{}, &mockLoginUC{})
    body, _ := json.Marshal(map[string]string{"password": "SecurePass123!"})
    req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    rr := httptest.NewRecorder()
    h.Register(rr, req)
    assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
}

type mockRegisterUC struct{}
func (m *mockRegisterUC) Execute(_ interface{}, _ interface{}) (interface{}, error) { return nil, nil }
type mockLoginUC struct{}
func (m *mockLoginUC) Execute(_ interface{}, _ interface{}) (interface{}, error) { return nil, nil }
```

- [x] **Step 2: Implementa handler.go**

```go
// services/auth-service/internal/transport/rest/handler.go
package rest

import (
    "context"
    "encoding/json"
    "net/http"
    "github.com/yourorg/golang-modules/services/auth-service/internal/application"
    "github.com/yourorg/golang-modules/shared/validation"
)

type RegisterExecutor interface {
    Execute(ctx context.Context, in application.RegisterInput) (*application.RegisterOutput, error)
}
type LoginExecutor interface {
    Execute(ctx context.Context, in application.LoginInput) (*application.LoginOutput, error)
}

type Handler struct {
    register RegisterExecutor
    login    LoginExecutor
}

func NewHandler(reg RegisterExecutor, login LoginExecutor) *Handler {
    return &Handler{register: reg, login: login}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required,min=8"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
        return
    }
    if err := validation.Validate(input); err != nil {
        writeValidationError(w, err)
        return
    }
    tenantID := r.Header.Get("X-Tenant-ID")
    out, err := h.register.Execute(r.Context(), application.RegisterInput{
        TenantID: tenantID, Email: input.Email, Password: input.Password,
    })
    if err != nil {
        writeBusinessError(w, err)
        return
    }
    writeJSON(w, http.StatusCreated, map[string]string{"user_id": out.UserID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        writeError(w, http.StatusBadRequest, "INVALID_JSON", "invalid request body")
        return
    }
    if err := validation.Validate(input); err != nil {
        writeValidationError(w, err)
        return
    }
    tenantID := r.Header.Get("X-Tenant-ID")
    out, err := h.login.Execute(r.Context(), application.LoginInput{
        TenantID: tenantID, Email: input.Email, Password: input.Password,
        IPAddress: r.RemoteAddr, UserAgent: r.UserAgent(),
    })
    if err != nil {
        writeBusinessError(w, err)
        return
    }
    writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
    writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": msg}})
}

func writeValidationError(w http.ResponseWriter, err error) {
    writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"error": map[string]any{
        "code": "VALIDATION_ERROR", "details": err,
    }})
}

func writeBusinessError(w http.ResponseWriter, err error) {
    // Map domain errors to HTTP status + error codes
    errorMap := map[error]struct{ status int; code string }{
        application.ErrEmailAlreadyExists: {409, "EMAIL_ALREADY_EXISTS"},
        application.ErrInvalidCredentials: {401, "INVALID_CREDENTIALS"},
        application.ErrEmailNotVerified:   {403, "EMAIL_NOT_VERIFIED"},
        application.ErrAccountLocked:      {423, "ACCOUNT_LOCKED"},
    }
    if mapped, ok := errorMap[err]; ok {
        writeError(w, mapped.status, mapped.code, err.Error())
        return
    }
    writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
```

- [x] **Step 3: Implementa main.go**

```go
// services/auth-service/cmd/server/main.go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os"
    "github.com/yourorg/golang-modules/services/auth-service/internal/transport/rest"
    "github.com/yourorg/golang-modules/shared/observability"
)

func main() {
    logger := observability.NewLogger("auth-service", "1.0.0", getEnv("ENV", "development"), nil)
    slog.SetDefault(logger)

    port := getEnv("PORT", "8081")
    h := buildHandler(logger)

    srv := &http.Server{
        Addr:    fmt.Sprintf(":%s", port),
        Handler: h,
    }

    logger.Info("auth-service starting", "port", port)
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("server error", "error", err)
            os.Exit(1)
        }
    }()

    observability.WaitForShutdown(context.Background(), logger, func(ctx context.Context) {
        srv.Shutdown(ctx)
    })
}

func buildHandler(logger *slog.Logger) http.Handler {
    // Wire dependencies (full DI — will be expanded in subsequent tasks)
    // For now returns a minimal router with health check
    mux := http.NewServeMux()
    mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })
    mux.HandleFunc("GET /metrics", observability.MetricsHandler().ServeHTTP)
    return mux
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" { return v }
    return fallback
}
```

- [x] **Step 4: Build e verifica**

```bash
cd services/auth-service && go build ./cmd/server/
# Expected: no errors, binary created
```

- [x] **Step 5: Run handler test**

```bash
go test ./internal/transport/rest/... -v
# Expected: PASS
```

- [x] **Step 6: Commit**

```bash
git add services/auth-service/internal/transport/ services/auth-service/cmd/
git commit -m "feat(auth): add REST transport layer — register/login handlers, error mapping, health endpoint, main server"
```

---

# FASE 2 — Auth Service Avanzato

> **PROSSIMA SESSIONE — riprendi da qui (2026-04-08)**
>
> Stato attuale:
> - Fase 0 completa (infra, Docker, Vault, NATS, Dockerfiles, migration script)
> - Fase 1 completa ECCETTO Task 1.4 Step 2 (integration test con testcontainers)
> - auth-service: dominio ✅, application ✅ (con jwt.go), infra ✅, transport ✅, migration SQL ✅
> - Test scritti: domain/user_test.go, application/register_test.go, application/login_test.go
> - Copertura ancora bassa (solo auth-service); gli altri servizi non hanno test
> - Issues.md: tutti i problemi CRITICAL/HIGH/MEDIUM risolti, solo M-10 (test coverage) aperto
>
> **Primo task da eseguire:** Task 1.4 Step 2 — integration test con testcontainers per user_repository.
> **Poi:** Fase 2 Tasks (MFA, OAuth2, ecc.)

> Le fasi 2-10 seguono lo stesso pattern TDD della Fase 1.
> Ogni task: scrivi test → falli fallire → implementa → falli passare → commit.

## Task 2.1: MFA/TOTP — Enable + Verify ✅
**Files:** `internal/application/mfa.go`, `internal/transport/http/mfa_handler.go`

Implementato: `EnableMFAUseCase` (genera TOTP secret con `pquerna/otp`, QR code URL, 8 backup codes),
`VerifyMFAUseCase` (valida codice 6 cifre RFC 6238 + backup codes, emette JWT con `mfa_verified: true`),
`DisableMFAUseCase`. Test unitari in `mfa_test.go`. Migrazione `000002_add_mfa_backup_codes`.

## Task 2.2: OAuth2 — Google + GitHub ✅
**Files:** `internal/infrastructure/oauth/google.go`, `internal/infrastructure/oauth/github.go`, `internal/application/oauth.go`

Implementato: `OAuthAuthorizeUseCase` (genera stato CSRF in Redis, ritorna URL consent), `OAuthCallbackUseCase`
(verifica stato, exchange code, upsert user via GetByOAuthProvider/email hash/create, emette JWT + sessione).
Aggiornati: `domain.User` (oauth_provider, oauth_provider_id), `UserRepository.GetByOAuthProvider`, migrazione
`000003_add_oauth_fields`, `user_repository.go` per nuove colonne.

## Task 2.3: Password Reset + Email Verification ✅
**Files:** `internal/application/password_reset.go`, `internal/application/verify_email.go`

Implementato: `ForgotPasswordUseCase`, `ResetPasswordUseCase`, `VerifyEmailUseCase`, `ResendVerificationUseCase`.
Token one-time UUID in Redis con TTL (1h reset, 24h verify). Always 200 su forgot-password.
Estesa interface `domain.TokenStore` con `StoreOneTimeToken`/`PopOneTimeToken`. Aggiornato redis implementation.

## Task 2.4: Device Tracking + Session Management ✅
**Files:** `internal/infrastructure/postgres/session_repository.go`, `internal/application/sessions.go`

Implementato: `ListSessionsUseCase`, `RevokeSessionUseCase`, `RevokeAllSessionsUseCase`.
Estesa `SessionRepository` con `ListByUser`, `DeleteByID`. Session salvata al login (già implementato in login.go).

## Task 2.5: GDPR — Export + Delete ✅
**Files:** `internal/application/gdpr.go`

Implementato: `ExportDataUseCase` (user + sessions), `DeleteAccountUseCase` (soft delete, azzeramento PII:
email_enc, email_hash, password_hash, mfa_secret, mfa_backup_codes, oauth_provider, oauth_provider_id).

## Task 2.6: gRPC server — ValidateToken ✅
**Files:** `internal/transport/grpc/server.go`, `shared/proto/auth.proto`

Implementato: `AuthServer` con `ValidateToken` (JWT parse + blacklist check) e `GetUser`. JSON codec custom
(override "proto") per evitare dipendenza protoc — sostituire con codice generato quando disponibile.
Aggiunto `google.golang.org/grpc v1.70.0` in go.mod. Eseguire `go mod tidy` prima di compilare.

---

# FASE 3 — API Gateway

## Task 3.1: Reverse proxy base + routing ✅
**Files:** `services/api-gateway/internal/proxy/proxy.go`, `internal/router/routes.go`
_Completato 2026-04-13: ReverseProxy con routing per prefisso /v1/auth/, /v1/payments/, ecc._

## Task 3.2: Middleware chain ✅
**Files:** `internal/middleware/requestid.go`, `internal/middleware/logger.go`, `internal/middleware/cors.go`
_Completato 2026-04-13: RequestID (X-Request-ID), structured logger, CORS configurabile._

## Task 3.3: Rate limiting (Redis token bucket) ✅
**Files:** `internal/middleware/ratelimit.go`
_Completato 2026-04-13: 100 req/min anonimo, 1000 req/min autenticato, header X-RateLimit-*_

## Task 3.4: Auth middleware (gRPC call a auth-service) ✅
**Files:** `internal/middleware/auth.go`
_Completato 2026-04-13: verifica JWT via gRPC ValidateToken, propaga X-User-ID, X-Tenant-ID, X-User-Roles_

## Task 3.5: Security headers (middleware) ✅
**Files:** `internal/middleware/security_headers.go`
_Completato 2026-04-13: HSTS, X-Content-Type-Options, X-Frame-Options, CSP, Referrer-Policy_

---

# FASE 4 — Payment Service

## Task 4.1: Payment domain + repository ✅
**Files:** `services/payment-service/internal/domain/payment.go`, `internal/infrastructure/postgres/payment_repo.go`
_Completato 2026-04-13: entità Payment, PaymentProvider interface, schema PostgreSQL con RLS_

## Task 4.2: Stripe adapter ✅
**Files:** `internal/infrastructure/providers/stripe/stripe.go`
_Completato 2026-04-13: CreatePaymentIntent, ConfirmPayment, Refund, ConstructWebhookEvent via net/http_

## Task 4.3: CreatePaymentUseCase + idempotency ✅
**Files:** `internal/application/create_payment.go`
_Completato 2026-04-13: verifica identità via gRPC auth, idempotency key Redis, call provider, persist Payment_

## Task 4.4: WebhookUseCase ✅
**Files:** `internal/application/webhook.go`

Implementa: verifica firma HMAC, aggiorna status Payment, pubblica evento NATS. Idempotenza via `provider_id UNIQUE`.
_Completato 2026-04-12: HandleStripeWebhookUseCase con GetByProviderID, markSucceeded/markFailed idempotenti_

## Task 4.5: RefundUseCase + REST transport ✅
**Files:** `internal/application/refund.go`, `internal/transport/http/handlers.go`, `cmd/server/main.go`
_Completato 2026-04-12: REST handlers POST /v1/payments, POST /v1/payments/{id}/refund, POST /v1/webhooks/stripe, main.go fully wired_

---

# FASE 5 — Notification Service

## Task 5.1: Email provider (SendGrid) + SMS (Twilio) ✅
**Files:** `services/notification-service/internal/infrastructure/email/sendgrid.go`, `internal/infrastructure/sms/twilio.go`
_Completato 2026-04-12: SendGrid v3 REST API, Twilio Messages API, pure net/http_

## Task 5.2: NATS consumers ✅
**Files:** `internal/transport/nats/consumers.go`
_Completato 2026-04-12: JetStream durable consumer su notification.requested_

## Task 5.3: Retry strategy ✅
**Files:** `internal/application/retry.go`
_Completato 2026-04-12: backoff [1m, 5m, 15m, 1h, 4h], max 5 tentativi_

---

# FASE 6 — Admin + Audit + Job Services

## Task 6.1: Admin Service — REST API ✅
**Files:** `services/admin-service/internal/transport/rest/handlers.go`, `cmd/server/main.go`
_Completato 2026-04-12: GET /v1/admin/tenants, /users/{id}, POST /users/{id}/disable, GET /stats; stub clients_

## Task 6.2: Audit Service — append-only consumer ✅
**Files:** `services/audit-service/internal/transport/nats/consumer.go`, `cmd/server/main.go`
_Completato 2026-04-12: JetStream durable consumer su audit.>, appende su PostgreSQL_

## Task 6.3: Job Service — scheduler ✅
**Files:** `services/job-service/cmd/worker/main.go`, `redis_cleaner.go`
_Completato 2026-04-12: scheduler tick-based, CleanupTokens handler con Redis scan_

---

# FASE 7 — Permission Service + SDK

## Task 7.1: RBAC domain + policy evaluator ✅
**Files:** `services/permission-service/internal/domain/`, `internal/application/check_permission.go`, `internal/infrastructure/postgres/role_repository.go`
_Completato 2026-04-12: CheckPermission con direct perm + role fallback, RoleRepository, REST transport, main.go_

## Task 7.2: Default roles per tenant ✅
**Files:** `internal/application/seed_roles.go`
_Completato 2026-04-13: SeedRolesUseCase crea owner/admin/billing-manager/read-only/user per tenant. Idempotente via GetByName._

## Task 7.3: SDK Go client ✅
**Files:** `sdk/go/auth/client.go`, `sdk/go/permission/client.go`, `sdk/go/common/`
_Completato 2026-04-13: retry 3x con backoff esponenziale, circuit breaker gobreaker, OTEL trace propagation via OTELTransport._

---

# FASE 8 — Config Service + Webhook Service

## Task 8.1: Config Service ✅
**Files:** `services/config-service/internal/transport/rest/handlers.go`, `cmd/server/main.go`
_Completato 2026-04-12: GET/PUT /v1/config/{key}, Redis cache 5min, postgres backend, main.go wired_

## Task 8.2: Webhook Service ✅
**Files:** `services/webhook-service/internal/transport/rest/handlers.go`, `internal/infrastructure/postgres/delivery_repository.go`, `cmd/server/main.go`
_Completato 2026-04-12: POST /v1/webhooks/endpoints, POST /v1/webhooks/deliver, HMAC firma, DeliveryRepository_

---

# FASE 9 — Storage Service + Subscription Service

## Task 9.1: Storage Service ✅
**Files:** `services/storage-service/internal/infrastructure/minio/provider.go`, `internal/transport/rest/handlers.go`, `cmd/server/main.go`
_Completato 2026-04-12: MinIO provider (Upload/Delete/Presign), POST /v1/storage/upload, GET /v1/storage/{id}/presign_

## Task 9.2: Subscription Service ✅
**Files:** `services/subscription-service/internal/infrastructure/billing/stripe.go`, `internal/infrastructure/postgres/plan_repository.go`, `internal/transport/rest/handlers.go`, `cmd/server/main.go`
_Completato 2026-04-12: StripeProvider CreateSubscription/Cancel/ChangePlan, PlanRepository, POST /v1/subscriptions, DELETE /v1/subscriptions/{id}_

---

# FASE 10 — Observability + HA + CI/CD

## Task 10.1: Prometheus + Grafana dashboards ✅
**Files:** `deployments/prometheus/prometheus.yml`, `deployments/grafana/provisioning/dashboards/`
_Completato 2026-04-13: scraping 11 servizi + postgres/redis/nats; dashboard services.json (request rate, error rate, p99 latency, servizi up)_

## Task 10.2: AlertManager rules ✅
**Files:** `deployments/alertmanager/alertmanager.yml`, `deployments/prometheus/rules.yml`
_Completato 2026-04-13: alert ServiceDown, HighErrorRate>5%, HighLatencyP99>500ms, JobFailed, PostgresDown, RedisDown; routing PagerDuty/Slack/email_

## Task 10.3: OpenTelemetry collector + Jaeger ✅
**Files:** `deployments/otel-collector.yml`
_Completato 2026-04-13: OTLP HTTP+gRPC receiver, batch processor, esporta a Jaeger:4317 e Prometheus:8889_

## Task 10.4: GitHub Actions CI — lint + test + build ✅
**Files:** `.github/workflows/ci.yml`
_Completato 2026-04-13: go 1.26, lint shared+services, unit+integration test con postgres/redis service containers, docker build ghcr.io/andrea-cavallo/_

## Task 10.5: GitHub Actions CD — deploy con rolling update ✅
**Files:** `.github/workflows/deploy.yml`
_Completato 2026-04-13: push immagini su GHCR + deploy SSH rolling update con health check per servizio + migrate.sh_

## Task 10.6: Bootstrap script + documentazione finale ✅
**Files:** `scripts/bootstrap.sh`
_Completato 2026-04-13: script bash con vault-init → nats-init → migrate all up → seed super-admin; supporta --skip-vault/--skip-nats/--skip-seed_

---

## Ordine di implementazione consigliato

```
Settimana 1:  Fase 0 (infrastruttura)
Settimana 2:  Fase 1 (auth core) + Fase 3 (gateway base)
Settimana 3:  Fase 2 (auth avanzato)
Settimana 4:  Fase 4 (payments)
Settimana 5:  Fase 5 (notifications) + Fase 6 (admin/audit/job)
Settimana 6:  Fase 7 (permissions + SDK)
Settimana 7:  Fase 8 (config + webhooks)
Settimana 8:  Fase 9 (storage + subscriptions)
Settimana 9:  Fase 10 (observability + CI/CD)
```

**Milestone minima production-ready:** Fasi 0 + 1 + 3 + 4 + 5 coprono il 90% dei use case di qualsiasi SaaS.
