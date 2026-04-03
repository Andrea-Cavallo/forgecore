# Microservizi Go 1.24 — Enterprise Production-Ready Design

**Data:** 2026-03-30
**Stato:** Approvato

---

## Contesto e Obiettivo

Sistema a microservizi in Go 1.24 scalabile, sicuro e production-ready composto da **10 servizi** riusabili come fondamenta per qualsiasi SaaS/backend:

| Servizio | Responsabilita' |
|---|---|
| `api-gateway` | Routing, rate limit, auth middleware, API versioning, TLS |
| `auth-service` | JWT RS256, OAuth2, MFA/TOTP, brute force, device tracking, GDPR |
| `payment-service` | Pagamenti con provider pluggable (Stripe-first), multi-tenant |
| `notification-service` | Email (SendGrid) + SMS (Twilio) pluggable, NATS consumer, retry |
| `admin-service` | Admin REST API, stats aggregate, gestione tenant/utenti |
| `audit-service` | Audit log immutabile append-only, GDPR compliant |
| `job-service` | Background jobs: retry notifiche, cleanup token, scheduled tasks |
| `permission-service` | RBAC/ABAC granulare, policy resource-based, valutazione permessi |
| `config-service` | Feature flags e configurazione per-tenant (MFA policy, OAuth providers, password rules) |
| `webhook-service` | Outbound webhook delivery verso sistemi tenant, retry, firma HMAC, delivery log |
| `storage-service` | File storage S3-compatible (MinIO), ricevute PDF, avatar, export CSV |
| `subscription-service` | Pagamenti ricorrenti, lifecycle subscription, billing cycle, proration |

**Deploy target:** Docker Compose + VPS
**Comunicazione:** REST/HTTP verso l'esterno, gRPC interno, NATS per eventi asincroni
**Database:** PostgreSQL per-service + Redis (token store, brute force, job queue)
**Multi-tenancy:** Row-Level Security PostgreSQL con `tenant_id` su ogni tabella
**Secrets:** HashiCorp Vault
**TLS:** Traefik + Let's Encrypt

---

## Struttura Repository (Monorepo)

```
golang-modules/
├── services/
│   ├── api-gateway/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── proxy/           # Reverse proxy handlers
│   │   │   ├── middleware/      # Rate limit, CORS, request-id, auth check
│   │   │   └── router/          # Route definitions + versioning /v1/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── auth-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # User, Token, Session, OAuthProvider, MFADevice
│   │   │   ├── application/     # UseCases: Register, Login, Refresh, Logout,
│   │   │   │                    #           OAuth2, MFA, GDPR, DeviceTracking
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # User, Session, OAuthAccount repositories
│   │   │   │   ├── redis/       # Token store, blacklist, brute force counters
│   │   │   │   └── oauth/       # Google, GitHub adapters
│   │   │   └── transport/
│   │   │       ├── rest/        # HTTP handlers v1
│   │   │       └── grpc/        # ValidateToken, GetUser
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── payment-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # Payment, Transaction, PaymentProvider interface
│   │   │   ├── application/     # UseCases: CreatePayment, Webhook, Refund, List
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # Payment repository
│   │   │   │   └── providers/
│   │   │   │       └── stripe/  # Stripe adapter
│   │   │   └── transport/
│   │   │       ├── rest/
│   │   │       └── grpc/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── notification-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # Notification, EmailProvider, SMSProvider
│   │   │   ├── application/     # UseCases: Send, Retry, LogNotification
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # Notification audit log
│   │   │   │   ├── email/       # SendGrid adapter
│   │   │   │   └── sms/         # Twilio adapter
│   │   │   └── transport/
│   │   │       └── nats/        # Event consumers
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── admin-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── application/     # UseCases: ManageTenants, ManageUsers, Stats
│   │   │   ├── infrastructure/
│   │   │   │   └── clients/     # gRPC clients per auth + payment
│   │   │   └── transport/
│   │   │       └── rest/        # /admin/* handlers
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── audit-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # AuditEvent (immutable)
│   │   │   ├── application/     # UseCase: RecordEvent, QueryEvents
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # append-only, no UPDATE/DELETE grant
│   │   │   │   └── nats/        # Consumer su audit.*
│   │   │   └── transport/
│   │   │       ├── rest/        # Query audit log (admin only)
│   │   │       └── grpc/
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── job-service/
│   │   ├── cmd/worker/main.go
│   │   ├── internal/
│   │   │   ├── jobs/            # RetryNotification, CleanupTokens, DailyReport
│   │   │   └── scheduler/       # Cron scheduler + asynq Redis queue
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── permission-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # Role, Permission, Policy, Resource
│   │   │   ├── application/     # UseCases: Check, Grant, Revoke, ListPermissions
│   │   │   ├── infrastructure/
│   │   │   │   └── postgres/    # Roles, permissions, role_bindings
│   │   │   └── transport/
│   │   │       ├── rest/        # /v1/permissions/*
│   │   │       └── grpc/        # CheckPermission (chiamato da ogni servizio)
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── config-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # TenantConfig, FeatureFlag, ConfigKey
│   │   │   ├── application/     # UseCases: GetConfig, SetConfig, ResetToDefault
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # tenant_configs table
│   │   │   │   └── redis/       # Cache config con invalidation
│   │   │   └── transport/
│   │   │       ├── rest/        # /v1/config/*
│   │   │       └── grpc/        # GetTenantConfig (chiamato da auth, payment, etc.)
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── webhook-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # WebhookEndpoint, WebhookDelivery, WebhookEvent
│   │   │   ├── application/     # UseCases: Register, Deliver, Retry, ListDeliveries
│   │   │   ├── infrastructure/
│   │   │   │   ├── postgres/    # endpoints, deliveries tables
│   │   │   │   └── nats/        # Consumer su tutti gli eventi
│   │   │   └── transport/
│   │   │       └── rest/        # /v1/webhooks/*
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── storage-service/
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── domain/          # File, Bucket, PresignedURL
│   │   │   ├── application/     # UseCases: Upload, Download, Delete, GeneratePresigned
│   │   │   ├── infrastructure/
│   │   │   │   ├── minio/       # MinIO S3-compatible adapter
│   │   │   │   └── postgres/    # File metadata
│   │   │   └── transport/
│   │   │       └── rest/        # /v1/files/*
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── subscription-service/
│       ├── cmd/server/main.go
│       ├── internal/
│       │   ├── domain/          # Subscription, Plan, BillingCycle, Invoice
│       │   ├── application/     # UseCases: Subscribe, Cancel, Upgrade, ProcessRenewal
│       │   ├── infrastructure/
│       │   │   ├── postgres/    # subscriptions, plans, invoices tables
│       │   │   └── providers/   # Stripe Billing adapter
│       │   └── transport/
│       │       ├── rest/        # /v1/subscriptions/*
│       │       └── grpc/
│       ├── go.mod
│       └── Dockerfile
│
├── shared/
│   ├── proto/                   # .proto per tutti i 10 servizi
│   ├── events/                  # NATS event struct definitions (typed Go)
│   ├── middleware/              # Tenant context, auth, permission check, request-id
│   ├── validation/             # go-playground/validator, errori strutturati
│   ├── crypto/                  # PII encryption helpers (pgcrypto AES-256)
│   ├── pagination/              # Cursor-based pagination helpers (standard condiviso)
│   ├── i18n/                    # Template traduzioni, locale helpers
│   └── observability/           # slog JSON, OTEL tracer, Prometheus, graceful shutdown
│
├── sdk/
│   └── go/                      # SDK Go client per i servizi (auth, payment, permission)
│       ├── auth/                # Client auth-service con retry + circuit breaker
│       ├── payment/             # Client payment-service
│       ├── permission/          # Client permission-service
│       └── go.mod
│
├── deployments/
│   ├── docker-compose.yml
│   ├── docker-compose.dev.yml
│   ├── traefik/
│   │   └── traefik.yml          # TLS, Let's Encrypt, middlewares
│   ├── vault/
│   │   └── config.hcl           # Vault dev/prod config
│   ├── pgbouncer/
│   │   └── pgbouncer.ini
│   ├── prometheus/
│   │   └── prometheus.yml
│   ├── alertmanager/
│   │   └── alertmanager.yml     # Slack/email alert rules
│   └── grafana/
│       └── provisioning/
│
├── scripts/
│   ├── gen-proto.sh
│   ├── migrate.sh
│   └── vault-init.sh            # Inizializza Vault secrets
│
└── docs/
    ├── CLAUDE.md
    └── superpowers/specs/
```

---

## Pattern Architetturale (tutti i servizi)

**DDD a 4 layer con Dependency Rule:**
```
transport/ → application/ → domain/
infrastructure/ → domain/ (implementa interfacce)
```

- `domain/` — entita', value objects, interfacce repository. Zero dipendenze esterne.
- `application/` — use cases. Dipende solo da `domain/`.
- `infrastructure/` — implementazioni concrete (postgres, redis, provider HTTP).
- `transport/` — handler REST/gRPC/NATS. Dipende da `application/`.

---

## Multi-Tenancy (RLS PostgreSQL)

Ogni tabella in ogni servizio segue questo schema:

```sql
CREATE TABLE <table> (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    ...
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON <table>(tenant_id);
ALTER TABLE <table> ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON <table>
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

Il middleware Go imposta il contesto tenant su ogni DB connection:
```go
conn.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
```

Il JWT include il claim `tenant_id`. Il gateway lo estrae e lo propaga via header `X-Tenant-ID`.

---

## Auth Service

### Dominio

```go
type User struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    Email        string       // cifrato AES-256 in DB, indice su hash(email)
    PasswordHash string       // bcrypt cost 12
    Roles        []string     // ["user"], ["admin"], ["super-admin"]
    MFAEnabled   bool
    MFASecret    string       // TOTP secret, cifrato in DB
    CreatedAt    time.Time
    DeletedAt    *time.Time   // soft delete per GDPR
}

type Session struct {
    ID         uuid.UUID
    TenantID   uuid.UUID
    UserID     uuid.UUID
    DeviceID   string
    UserAgent  string
    IPAddress  string
    LastSeenAt time.Time
    ExpiresAt  time.Time
}

type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCode(ctx context.Context, code string) (*OAuthUser, error)
}
```

### Use Cases

- `RegisterUseCase` — valida email unica per tenant, hash bcrypt cost 12, pubblica `auth.user.registered`
- `LoginUseCase` — verifica brute force → credenziali → MFA se abilitato → emette TokenPair → crea Session
- `RefreshUseCase` — valida refresh token Redis → ruota token (sliding window)
- `LogoutUseCase` — revoca refresh token Redis, blacklista access token jti, elimina Session
- `OAuthCallbackUseCase` — exchange code → upsert utente → TokenPair
- `EnableMFAUseCase` — genera TOTP secret, restituisce QR code + backup codes
- `VerifyMFAUseCase` — verifica codice TOTP 6 cifre (RFC 6238)
- `ValidateTokenUseCase` — verifica JWT + blacklist Redis (gRPC)
- `VerifyEmailUseCase` — valida token Redis → imposta `email_verified = true`
- `ForgotPasswordUseCase` — genera token one-time Redis (TTL 1h) → pubblica `auth.user.password_reset`
- `ResetPasswordUseCase` — valida token Redis → bcrypt nuova password → invalida tutti i refresh token utente
- `ChangePasswordUseCase` — verifica password corrente → bcrypt nuova → invalida sessioni precedenti
- `ExportDataUseCase` — GDPR export JSON dati utente
- `DeleteAccountUseCase` — GDPR soft delete + anonimizzazione PII

### Token Strategy

- **Access Token:** JWT RS256, TTL 15 minuti, claims: `sub`, `tenant_id`, `roles`, `jti`, `mfa_verified`
- **Refresh Token:** UUID v4 opaco, TTL 7 giorni, Redis: `refresh:{tenantID}:{userID}:{jti}`
- **Revocation:** logout → blacklist jti in Redis (TTL residuo) + elimina refresh token
- **MFA:** se abilitato, access token iniziale ha `mfa_verified: false`, valido solo per `/auth/mfa/verify`

### Brute Force Protection

```
Redis key: bruteforce:{tenantID}:{ip}:{email}
Tentativi: 1-5   → nessun blocco
Tentativo 5      → lockout 15 minuti
Tentativo 8      → lockout 1 ora
Tentativo 10+    → lockout 24 ore
Risposta:        401 + header Retry-After
```

### PostgreSQL Schema

```sql
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL,
    email_enc     BYTEA NOT NULL,          -- AES-256 encrypted
    email_hash    VARCHAR(64) NOT NULL,    -- SHA-256 per lookup
    password_hash VARCHAR(255),
    roles         TEXT[] NOT NULL DEFAULT '{"user"}',
    mfa_enabled      BOOLEAN NOT NULL DEFAULT false,
    mfa_secret       BYTEA,                  -- TOTP secret encrypted
    email_verified   BOOLEAN NOT NULL DEFAULT false,
    locked_until     TIMESTAMPTZ,           -- brute force lockout recovery
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    UNIQUE(tenant_id, email_hash)
);

CREATE TABLE oauth_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider    VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    UNIQUE(tenant_id, provider, provider_id)
);

CREATE TABLE sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id   VARCHAR(255) NOT NULL,
    user_agent  TEXT,
    ip_address  INET,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE mfa_backup_codes (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(64) NOT NULL,
    used_at   TIMESTAMPTZ
);

-- RLS su tutte le tabelle (pattern ripetuto)
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### Redis Namespace

```
refresh:{tenantID}:{userID}:{jti}      TTL: 7d
blacklist:{jti}                        TTL: remaining access token TTL
bruteforce:{tenantID}:{ip}:{email}     TTL: lockout duration
oauth:state:{state}                    TTL: 5min
email_verify:{token}                   TTL: 24h  → userID
password_reset:{token}                 TTL: 1h   → userID
```

### Endpoints REST (/v1)

| Method | Path | Auth | Descrizione |
|--------|------|------|-------------|
| POST | `/v1/auth/register` | No | Registrazione (invia email verifica) |
| POST | `/v1/auth/login` | No | Login (+ brute force check) |
| POST | `/v1/auth/refresh` | No | Refresh token |
| POST | `/v1/auth/logout` | Bearer | Revoca token |
| GET | `/v1/auth/oauth/:provider` | No | Init OAuth2 |
| GET | `/v1/auth/oauth/:provider/callback` | No | OAuth2 callback |
| GET | `/v1/auth/me` | Bearer | Profilo utente |
| PUT | `/v1/auth/me/password` | Bearer | Cambia password |
| POST | `/v1/auth/verify-email` | No | Verifica email (token da email) |
| POST | `/v1/auth/resend-verification` | No | Reinvia email di verifica |
| POST | `/v1/auth/forgot-password` | No | Richiede reset, invia email con token |
| POST | `/v1/auth/reset-password` | No | Usa token, imposta nuova password |
| POST | `/v1/auth/mfa/enable` | Bearer | Abilita MFA, ritorna QR + backup codes |
| POST | `/v1/auth/mfa/verify` | Bearer (partial) | Verifica TOTP |
| DELETE | `/v1/auth/mfa` | Bearer+MFA | Disabilita MFA |
| GET | `/v1/auth/sessions` | Bearer | Lista sessioni attive |
| DELETE | `/v1/auth/sessions/:id` | Bearer | Revoca sessione device |
| GET | `/v1/auth/me/export` | Bearer | GDPR data export |
| DELETE | `/v1/auth/me` | Bearer+MFA | GDPR account deletion |
| GET | `/health` | No | Health check |
| GET | `/health/ready` | No | Readiness probe |
| GET | `/health/live` | No | Liveness probe |
| GET | `/metrics` | No | Prometheus |

### gRPC (porta 9091)

```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}
```

---

## Payment Service

### Dominio

```go
type PaymentProvider interface {
    CreatePaymentIntent(ctx context.Context, req CreateIntentRequest) (*PaymentIntent, error)
    ConfirmPayment(ctx context.Context, intentID string) (*Payment, error)
    Refund(ctx context.Context, paymentID string, amount int64) (*Refund, error)
    ConstructWebhookEvent(payload []byte, signature string) (*WebhookEvent, error)
}

type Payment struct {
    ID         uuid.UUID
    TenantID   uuid.UUID
    UserID     uuid.UUID
    Amount     int64           // centesimi
    Currency   string          // "eur", "usd"
    Status     PaymentStatus   // pending, succeeded, failed, refunded
    Provider   string          // "stripe"
    ProviderID string          // ID esterno (idempotenza)
    Metadata   map[string]any
    CreatedAt  time.Time
    UpdatedAt  time.Time
}
```

### PostgreSQL Schema

```sql
CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    amount      BIGINT NOT NULL CHECK (amount > 0),
    currency    VARCHAR(3) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    provider    VARCHAR(20) NOT NULL,
    provider_id VARCHAR(255),
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider, provider_id)  -- idempotenza webhook
);
CREATE INDEX ON payments(tenant_id);
CREATE INDEX ON payments(tenant_id, user_id);
CREATE INDEX ON payments(tenant_id, status);
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON payments
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### Endpoints REST (/v1)

| Method | Path | Auth | Descrizione |
|--------|------|------|-------------|
| POST | `/v1/payments` | Bearer | Crea payment intent (supporta `Idempotency-Key` header) |
| GET | `/v1/payments` | Bearer | Lista pagamenti (paginata) |
| GET | `/v1/payments/:id` | Bearer | Dettaglio |
| POST | `/v1/payments/:id/refund` | Bearer | Rimborso |
| POST | `/v1/payments/webhook/:provider` | Signature | Webhook provider |
| GET | `/health` | No | Health |
| GET | `/metrics` | No | Prometheus |

### NATS Events pubblicati

```
payment.succeeded  → {paymentID, tenantID, userID, amount, currency}
payment.failed     → {paymentID, tenantID, userID, reason}
payment.refunded   → {paymentID, tenantID, userID, amount}
```

---

## Notification Service

### Dominio

```go
type EmailProvider interface {
    Send(ctx context.Context, msg *EmailMessage) error
}
type SMSProvider interface {
    Send(ctx context.Context, msg *SMSMessage) error
}

// Implementazioni: SendGrid (email), Twilio (SMS)
// Aggiungibili senza toccare il dominio: SES, Vonage, etc.
```

### Consumer NATS

```go
nc.Subscribe("auth.user.registered",     h.WelcomeEmail)
nc.Subscribe("auth.user.password_reset", h.PasswordResetEmail)
nc.Subscribe("auth.mfa.enabled",         h.MFAEnabledEmail)
nc.Subscribe("payment.succeeded",        h.PaymentReceiptEmailAndSMS)
nc.Subscribe("payment.failed",           h.PaymentFailedEmail)
nc.Subscribe("payment.refunded",         h.RefundConfirmEmail)
nc.Subscribe("admin.user.disabled",      h.AccountDisabledEmail)
nc.Subscribe("notification.retry",       h.RetryFailed)  // da job-service
```

### Retry Strategy

Notifica fallita → pubblica `notification.retry` con backoff esponenziale (1min, 5min, 15min, 1h). Dopo 5 tentativi → status `failed`, alert admin.

---

## Admin Service

Aggregatore read-model: chiama auth e payment via gRPC. Nessun DB proprio — legge dai servizi downstream.

### Endpoints REST (/v1/admin)

| Method | Path | Ruolo | Descrizione |
|--------|------|-------|-------------|
| GET | `/v1/admin/tenants` | super-admin | Lista tenant |
| POST | `/v1/admin/tenants` | super-admin | Crea tenant |
| GET | `/v1/admin/tenants/:id/users` | admin | Utenti tenant |
| GET | `/v1/admin/tenants/:id/payments` | admin | Pagamenti tenant |
| GET | `/v1/admin/users/:id` | admin | Dettaglio utente |
| PUT | `/v1/admin/users/:id/roles` | admin | Modifica ruoli |
| DELETE | `/v1/admin/users/:id` | admin | Disabilita account |
| POST | `/v1/admin/users/:id/unlock` | admin | Sblocca account (brute force recovery) |
| GET | `/v1/admin/payments` | admin | Tutti i pagamenti |
| GET | `/v1/admin/stats` | admin | Dashboard metrics |
| GET | `/v1/admin/audit` | admin | Query audit log |

---

## Audit Service

Append-only: il DB user del servizio ha solo INSERT + SELECT, mai UPDATE o DELETE.

### Schema

```sql
CREATE TABLE audit_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL,
    actor_id    UUID,          -- userID o NULL per system events
    actor_type  VARCHAR(20),   -- "user", "system", "admin"
    action      VARCHAR(100) NOT NULL,  -- "user.login", "payment.succeeded"
    resource_id UUID,
    resource_type VARCHAR(50),
    metadata    JSONB,
    ip_address  INET,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- NO updated_at, NO deleted_at: immutable
);
CREATE INDEX ON audit_events(tenant_id, occurred_at DESC);
CREATE INDEX ON audit_events(tenant_id, actor_id);
CREATE INDEX ON audit_events(tenant_id, action);
ALTER TABLE audit_events ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON audit_events
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

### Consumer NATS

Subscribes a `audit.>` (wildcard) — ogni servizio pubblica su `audit.*` per eventi sensibili.

---

## Job Service

Worker puro (no porte HTTP pubbliche). Usa `asynq` con Redis backend.

### Jobs schedulati

| Job | Frequenza | Descrizione |
|-----|-----------|-------------|
| `CleanupExpiredTokens` | Ogni ora | Rimuove refresh token scaduti da Redis |
| `RetryFailedNotifications` | Ogni 5 minuti | Ripubblica notifiche fallite su NATS |
| `DailyRevenueReport` | 08:00 ogni giorno | Calcola stats e notifica admin via email |
| `SessionCleanup` | Ogni giorno | Elimina sessioni scadute da PostgreSQL |
| `AuditRetention` | Ogni settimana | Archivia eventi audit > retention policy |

---

## API Gateway

### Middleware Chain (ordine di esecuzione)

```
Request
  → RequestID (genera X-Request-ID)
  → Logger (log ingresso con request-id, method, path)
  → RateLimit (token bucket Redis: 100 req/min IP, 1000 req/min autenticato)
  → CORS
  → Auth (se route protetta: valida JWT via gRPC auth-service, imposta X-User-ID, X-Tenant-ID, X-User-Roles)
  → TenantContext (imposta tenant_id nel context)
  → Proxy (forwarda al servizio corretto)
  → Logger (log uscita con status, latenza)
```

### Routing

```
/v1/auth/*        → auth-service:8081
/v1/payments/*    → payment-service:8082
/v1/admin/*       → admin-service:8084
/v1/audit/*       → audit-service:8085
```

---

## Infrastruttura Docker Compose

```yaml
services:
  traefik:
    image: traefik:v3
    ports: ["80:80", "443:443"]
    volumes:
      - ./deployments/traefik/traefik.yml:/etc/traefik/traefik.yml
      - letsencrypt:/letsencrypt

  api-gateway:
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.gateway.rule=Host(`api.yourdomain.com`)"
      - "traefik.http.routers.gateway.tls.certresolver=letsencrypt"

  auth-service:
    ports: ["8081:8081", "9091:9091"]
    depends_on: [postgres-auth, redis, vault, nats]

  payment-service:
    ports: ["8082:8082", "9092:9092"]
    depends_on: [postgres-payments, auth-service, vault, nats]

  notification-service:
    depends_on: [postgres-notifications, nats, vault]

  admin-service:
    ports: ["8084:8084", "9094:9094"]
    depends_on: [auth-service, payment-service]

  audit-service:
    ports: ["8085:8085", "9095:9095"]
    depends_on: [postgres-audit, nats]

  job-service:
    depends_on: [redis, nats]

  postgres-auth:         image: postgres:16-alpine
  postgres-payments:     image: postgres:16-alpine
  postgres-notifications: image: postgres:16-alpine
  postgres-audit:        image: postgres:16-alpine

  pgbouncer:             image: pgbouncer/pgbouncer
  redis:                 image: redis:7-alpine
  nats:                  image: nats:2-alpine
  vault:                 image: hashicorp/vault:1.17

  prometheus:            image: prom/prometheus
  alertmanager:          image: prom/alertmanager
  grafana:               image: grafana/grafana
  jaeger:                image: jaegertracing/all-in-one
  otel-collector:        image: otel/opentelemetry-collector-contrib
```

---

## Sicurezza

### TLS
- Traefik gestisce TLS termination + Let's Encrypt auto-renew
- HTTP rediretto automaticamente a HTTPS
- TLS 1.2 minimo, cipher suites moderne

### mTLS gRPC Interno
- Vault PKI emette certificati per ogni servizio
- Ogni gRPC server verifica il certificato client
- Rotazione automatica certificati ogni 30 giorni

### Secrets (HashiCorp Vault)
- JWT private key (RS256): `secret/auth/jwt_private_key`
- DB passwords: `secret/db/<service>/password`
- Stripe secret key: `secret/payment/stripe_secret`
- SendGrid API key: `secret/notification/sendgrid_key`
- Twilio credentials: `secret/notification/twilio`
- PII encryption key: `secret/crypto/pii_key`
- I servizi leggono secrets da Vault all'avvio via Vault Agent Sidecar

### PII Encryption
- Email e numero telefono cifrati AES-256-GCM con chiave da Vault
- Lookup tramite indice su `SHA-256(email)` — non reversibile senza chiave
- Chiave ruotabile senza downtime (re-encryption job schedulato)

### Input Validation
- `shared/validation/` — wrapper su `go-playground/validator v10`
- Ogni use case valida l'input prima di qualsiasi operazione
- Errori strutturati: `{"errors": [{"field": "email", "message": "invalid format"}]}`
- Sanitizzazione: strip HTML da tutti i campi stringa

---

## Observability

### Logging (slog JSON)
```json
{
  "time": "2026-03-30T10:00:00Z",
  "level": "INFO",
  "msg": "payment created",
  "service": "payment-service",
  "version": "1.0.0",
  "env": "production",
  "trace_id": "abc123",
  "span_id": "def456",
  "request_id": "uuid",
  "tenant_id": "uuid",
  "user_id": "uuid",
  "payment_id": "uuid",
  "amount": 1000,
  "currency": "eur"
}
```

### Metrics Prometheus (per ogni servizio)
- `http_requests_total{method,path,status_code,service}`
- `http_request_duration_seconds{method,path,service}` (histogram, p50/p95/p99)
- `grpc_requests_total{method,status,service}`
- `payments_created_total{provider,currency,tenant_id}`
- `auth_tokens_issued_total{type,tenant_id}`
- `notifications_sent_total{channel,template,status}`
- `audit_events_total{action}`

### Alerting (AlertManager)
- Servizio down > 1 minuto → alert critico Slack
- Error rate > 5% su 5 minuti → alert warning
- Latenza p99 > 500ms su 5 minuti → alert warning
- Job fallito > 3 volte → alert critico

### Distributed Tracing (OpenTelemetry → Jaeger)
- Ogni request HTTP riceve `trace_id` (W3C `traceparent`)
- Gateway propaga trace context a tutti i servizi via header
- gRPC propaga trace context via metadata
- Span inclusi: HTTP handler → use case → repository → DB query

### Health Checks
```
GET /health        {"status":"ok","checks":{"db":"ok","redis":"ok","nats":"ok","vault":"ok"}}
GET /health/ready  200 = pronto, 503 = non pronto
GET /health/live   200 = vivo
```

---

## Graceful Shutdown (tutti i servizi)

```go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
defer stop()

// Avvia server
go server.ListenAndServe()

// Attendi segnale
<-ctx.Done()

// Shutdown con timeout 30s
shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
server.Shutdown(shutdownCtx)
db.Close()
redis.Close()
nats.Drain() // aspetta consumer NATS di finire
```

---

## API Versioning

- Tutti gli endpoint pubblici hanno prefisso `/v1/`
- Breaking changes → nuovo prefisso `/v2/` senza rimuovere `/v1/`
- Gateway gestisce routing per versione
- Deprecation header: `Sunset: <date>` per endpoint obsoleti

---

## Dipendenze Go principali

| Libreria | Versione | Scopo |
|----------|----------|-------|
| `google.golang.org/grpc` | latest | gRPC client/server |
| `github.com/golang-jwt/jwt/v5` | v5 | JWT RS256 |
| `golang.org/x/oauth2` | latest | OAuth2 |
| `github.com/redis/go-redis/v9` | v9 | Redis client |
| `github.com/jackc/pgx/v5` | v5 | PostgreSQL driver |
| `github.com/golang-migrate/migrate/v4` | v4 | DB migrations |
| `github.com/stripe/stripe-go/v76` | v76 | Stripe |
| `go.opentelemetry.io/otel` | latest | Tracing |
| `github.com/prometheus/client_golang` | latest | Metrics |
| `github.com/nats-io/nats.go` | latest | NATS client |
| `github.com/hibiken/asynq` | latest | Background jobs |
| `github.com/pquerna/otp` | latest | TOTP/MFA |
| `github.com/go-playground/validator/v10` | v10 | Input validation |
| `github.com/google/uuid` | latest | UUID v4 |
| `github.com/hashicorp/vault/api` | v2 | Vault client |
| `golang.org/x/crypto` | latest | bcrypt, AES |
| `github.com/nats-io/nats.go` | latest | NATS JetStream client |
| `github.com/swaggo/swag` | latest | OpenAPI codegen |

---

## NATS JetStream (Durabilita' Messaggi)

Il design usa **NATS JetStream** (non NATS core pub/sub) per garantire la consegna dei messaggi anche quando i consumer sono temporaneamente down.

### Streams configurati

```
Stream: AUTH_EVENTS
  Subjects: auth.>
  Retention: WorkQueue
  MaxAge: 7d
  Replicas: 3 (cluster NATS)

Stream: PAYMENT_EVENTS
  Subjects: payment.>
  Retention: WorkQueue
  MaxAge: 30d
  Replicas: 3

Stream: AUDIT_EVENTS
  Subjects: audit.>
  Retention: Limits
  MaxAge: 365d   -- audit trail annuale
  Replicas: 3

Stream: NOTIFICATION_RETRY
  Subjects: notification.retry
  Retention: WorkQueue
  MaxAge: 24h
  Replicas: 3
```

### Consumer durability

```go
// Consumer durevole: se notification-service si riavvia, riprende dal punto in cui si era fermato
js.Subscribe("payment.succeeded", handler,
    nats.Durable("notification-payment-succeeded"),
    nats.AckExplicit(),
    nats.MaxDeliver(5),         // max 5 tentativi
    nats.AckWait(30*time.Second),
)
```

### Dead Letter Queue

Dopo `MaxDeliver` tentativi falliti → messaggio finisce in `DLQ.{subject}`. Il `job-service` processa la DLQ e notifica admin via alert.

---

## Email Verification Flow

```
1. POST /v1/auth/register
   → crea utente con email_verified = false
   → genera token UUID (TTL 24h) in Redis: email_verify:{token} → userID
   → pubblica auth.user.registered con verifyURL
   → notification-service invia email con link

2. GET /v1/auth/verify-email?token={token}
   → legge Redis → userID
   → imposta email_verified = true
   → elimina token da Redis
   → risponde 200

3. Login con email non verificata
   → risponde 403 con code: "email_not_verified"
   → suggerisce POST /v1/auth/resend-verification

4. POST /v1/auth/resend-verification
   → rate limit: max 3 email ogni 10min per utente
   → genera nuovo token, invalida il precedente
   → reinvia email
```

---

## Password Reset Flow

```
1. POST /v1/auth/forgot-password {"email": "user@example.com"}
   → sempre risponde 200 (non rivela se email esiste — security best practice)
   → se utente esiste: genera token UUID (TTL 1h) in Redis: password_reset:{token} → userID
   → pubblica auth.user.password_reset con resetURL
   → notification-service invia email

2. POST /v1/auth/reset-password {"token": "...", "new_password": "..."}
   → valida token Redis → userID
   → valida policy password (min 8 char, 1 maiuscolo, 1 numero)
   → bcrypt nuova password (cost 12)
   → invalida TUTTI i refresh token dell'utente (Redis: DELETE refresh:{tenantID}:{userID}:*)
   → elimina token da Redis
   → pubblica auth.user.password_changed (audit)
   → risponde 200
```

---

## Idempotency Keys (Payment Service)

```
Header: Idempotency-Key: {UUID generato dal client}

Flusso:
1. Client invia POST /v1/payments con Idempotency-Key
2. Gateway/service controlla Redis: idempotency:{tenantID}:{key}
3. Se esiste → restituisce la risposta cached (HTTP 200, stesso body)
4. Se non esiste → processa, salva risposta in Redis (TTL 24h), restituisce risposta

Redis key: idempotency:{tenantID}:{idempotencyKey}  TTL: 24h
```

Questo protegge da retry del client che causerebbero pagamenti duplicati.

---

## System Bootstrap (Primo Avvio)

Problema: come si crea il primo super-admin e il primo tenant?

### Soluzione: comando CLI seed

```bash
# Eseguito UNA SOLA VOLTA su primo deploy
go run ./services/auth-service/cmd/seed \
  --tenant-name "System" \
  --admin-email "admin@yourdomain.com" \
  --admin-password "$(openssl rand -base64 32)"
```

Il comando:
1. Crea tenant `system` con ID fisso (`00000000-0000-0000-0000-000000000001`)
2. Crea utente con ruolo `super-admin`, `email_verified = true`
3. Stampa le credenziali a stdout (mai salvate su DB o log)
4. Fallisce se super-admin esiste gia' (idempotente)

Alternativa per ambienti senza accesso CLI: variabili d'ambiente `SEED_ADMIN_EMAIL` + `SEED_ADMIN_PASSWORD` processate solo al primo avvio (flag `bootstrapped` in Redis).

---

## Testing Strategy

### Struttura per ogni servizio

```
services/<service>/
├── internal/
│   ├── domain/
│   │   └── *_test.go          # Unit test puri, zero dipendenze esterne
│   ├── application/
│   │   └── *_test.go          # Unit test con mock dei repository (mockery)
│   └── infrastructure/
│       └── *_integration_test.go  # Integration test vs PostgreSQL/Redis reali
└── e2e/
    └── *_test.go              # E2E test vs servizio avviato (testcontainers-go)
```

### Tipi di test

**Unit tests** (`go test ./internal/domain/... ./internal/application/...`):
- Mock generati con `mockery` dalle interfacce del dominio
- Coprono tutti i casi d'uso e la business logic
- Target coverage: >85%
- Veloci: <100ms per package

**Integration tests** (`go test -tags=integration ./internal/infrastructure/...`):
- Usano `testcontainers-go` per avviare PostgreSQL e Redis reali in Docker
- Testano repository, migration, RLS policy
- Verifica esplicita dell'isolamento tenant (inserisci dati tenant A, verifica non visibili da tenant B)

**E2E tests** (`go test -tags=e2e ./e2e/...`):
- Stack completo con `docker-compose -f docker-compose.test.yml up`
- Testano i flussi completi: register → verify email → login → MFA → payment
- Eseguiti in CI su ogni PR

**Load tests** (`./scripts/load-test.sh`):
- Strumento: `k6`
- Scenari: login 1000 utenti concorrenti, creazione 500 pagamenti/sec
- Threshold: p99 < 200ms, error rate < 0.1%
- Eseguiti pre-release, non su ogni commit

### Coverage minima richiesta

| Layer | Target |
|-------|--------|
| domain/ | 90% |
| application/ | 85% |
| infrastructure/ | 70% (integration test) |
| transport/ | 60% |

---

## CI/CD Pipeline (GitHub Actions)

### Pipeline per ogni PR

```yaml
# .github/workflows/ci.yml
jobs:
  lint:
    - golangci-lint run ./...
    - go vet ./...
    - govulncheck ./...

  test:
    - go test -race -coverprofile=coverage.out ./internal/domain/... ./internal/application/...
    - go test -tags=integration -race ./internal/infrastructure/...
    - Upload coverage to Codecov (threshold: 80%)

  build:
    - docker build services/auth-service
    - docker build services/payment-service
    - docker build services/notification-service
    - docker build services/admin-service
    - docker build services/audit-service
    - docker build services/job-service
    - docker build services/api-gateway

  proto:
    - buf lint shared/proto
    - buf generate shared/proto
    - Verifica che il codice generato sia committato (no drift)
```

### Pipeline di deploy (main branch)

```yaml
# .github/workflows/deploy.yml
jobs:
  e2e:
    - docker-compose -f docker-compose.test.yml up -d
    - go test -tags=e2e ./e2e/...
    - docker-compose down

  push-images:
    - docker buildx build --platform linux/amd64,linux/arm64
    - docker push ghcr.io/yourorg/{service}:{git-sha}

  deploy:
    - ssh VPS
    - docker-compose pull
    - docker-compose up -d --no-deps --scale api-gateway=2  # rolling update
    - ./scripts/migrate.sh (run migrations)
    - Health check: curl /health/ready x 30s
    - Rollback automatico se health check fallisce
```

### Zero-downtime deployment

```bash
# Rolling update su Docker Compose (2 istanze per servizio critico)
docker-compose up -d --no-deps --scale auth-service=2
sleep 10  # attendi che nuova istanza sia healthy
docker-compose up -d --no-deps --scale auth-service=1
```

---

## Database Migrations (Zero-Downtime)

### Regole per migration sicure

1. **Mai DROP COLUMN o DROP TABLE** in una singola migration — fai in 3 step:
   - Step 1: aggiungi nuova colonna nullable
   - Step 2: migra dati (deploy applicazione che usa entrambe le colonne)
   - Step 3: rimuovi vecchia colonna (deploy successivo)

2. **Mai rinominare colonne** — aggiungi nuova, migra, rimuovi vecchia

3. **Indici creati con CONCURRENTLY** — non bloccano scritture:
   ```sql
   CREATE INDEX CONCURRENTLY idx_payments_user_id ON payments(user_id);
   ```

4. **Migration eseguite all'avvio** in un goroutine separato prima di accettare traffico:
   ```go
   if err := runMigrations(db); err != nil {
       log.Fatal("migration failed", "error", err)
   }
   ```

5. **Rollback**: ogni migration ha il corrispondente `down` migration

### Struttura migrations

```
services/<service>/
└── migrations/
    ├── 000001_create_users.up.sql
    ├── 000001_create_users.down.sql
    ├── 000002_add_mfa.up.sql
    └── 000002_add_mfa.down.sql
```

---

## Configurazione e Validazione all'Avvio

Ogni servizio ha una struct `Config` con tag `validate` — fail fast se mancano variabili critiche:

```go
type Config struct {
    Port        int    `env:"PORT" validate:"required,min=1024,max=65535"`
    DatabaseURL string `env:"DATABASE_URL" validate:"required,url"`
    RedisURL    string `env:"REDIS_URL" validate:"required"`
    VaultAddr   string `env:"VAULT_ADDR" validate:"required,url"`
    VaultToken  string `env:"VAULT_TOKEN" validate:"required"`
    NATSUrl     string `env:"NATS_URL" validate:"required"`
    Environment string `env:"ENV" validate:"required,oneof=development staging production"`
    LogLevel    string `env:"LOG_LEVEL" validate:"oneof=debug info warn error"`
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    if err := envconfig.Process("", cfg); err != nil {
        return nil, fmt.Errorf("loading env: %w", err)
    }
    if err := validator.New().Struct(cfg); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)  // fail fast
    }
    return cfg, nil
}
```

Un file `.env.example` con tutte le variabili (senza valori sensibili) e' committato per ogni servizio.

---

## Security Headers HTTP

Traefik aggiunge questi header a tutte le risposte:

```yaml
# traefik/traefik.yml — middleware secureHeaders
http:
  middlewares:
    secure-headers:
      headers:
        stsSeconds: 31536000              # HSTS 1 anno
        stsIncludeSubdomains: true
        stsPreload: true
        forceSTSHeader: true
        contentTypeNosniff: true          # X-Content-Type-Options: nosniff
        frameDeny: true                   # X-Frame-Options: DENY
        referrerPolicy: "strict-origin-when-cross-origin"
        contentSecurityPolicy: "default-src 'self'"
        permissionsPolicy: "camera=(), microphone=(), geolocation=()"
        customResponseHeaders:
          X-Powered-By: ""                # rimuove header che rivelano tecnologia
          Server: ""
```

---

## OpenAPI / Documentazione API

### Approccio: spec-first con `buf` per gRPC, `swag` per REST

```bash
# Genera spec OpenAPI 3.0 dai commenti Go
swag init -g cmd/server/main.go -o docs/api

# Genera client SDK da spec (opzionale)
openapi-generator generate -i docs/api/swagger.json -g typescript-axios
```

Ogni handler REST ha commenti `swag`:
```go
// CreatePayment godoc
// @Summary      Crea un payment intent
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        Idempotency-Key  header  string  false  "Idempotency key"
// @Param        body  body  CreatePaymentRequest  true  "Payment request"
// @Success      201  {object}  PaymentResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      401  {object}  ErrorResponse
// @Router       /v1/payments [post]
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
```

La spec OpenAPI e' servita da ogni servizio su `/docs` (ambiente non-production).

---

## Catalogo Errori Standard

Ogni servizio restituisce errori nel formato:

```json
{
  "error": {
    "code": "EMAIL_NOT_VERIFIED",
    "message": "Email address must be verified before login",
    "request_id": "uuid"
  }
}
```

### Codici errore auth-service

| Code | HTTP | Descrizione |
|------|------|-------------|
| `INVALID_CREDENTIALS` | 401 | Email o password errati |
| `EMAIL_NOT_VERIFIED` | 403 | Email non verificata |
| `ACCOUNT_LOCKED` | 423 | Brute force lockout (header Retry-After) |
| `TOKEN_EXPIRED` | 401 | JWT scaduto |
| `TOKEN_REVOKED` | 401 | Token nella blacklist |
| `MFA_REQUIRED` | 403 | MFA abilitato, verifica richiesta |
| `MFA_INVALID_CODE` | 422 | Codice TOTP non valido |
| `EMAIL_ALREADY_EXISTS` | 409 | Email gia' registrata nel tenant |
| `RESET_TOKEN_INVALID` | 422 | Token reset scaduto o non valido |

### Codici errore payment-service

| Code | HTTP | Descrizione |
|------|------|-------------|
| `PAYMENT_FAILED` | 402 | Provider ha rifiutato il pagamento |
| `PAYMENT_NOT_FOUND` | 404 | Pagamento non trovato |
| `REFUND_NOT_ALLOWED` | 422 | Pagamento non in stato succeeded |
| `IDEMPOTENCY_CONFLICT` | 422 | Idempotency key riusata con body diverso |
| `PROVIDER_UNAVAILABLE` | 503 | Provider di pagamento irraggiungibile |

---

## Alta Disponibilita' (HA) e Resilienza

### Redis Sentinel (minimo per VPS)

```yaml
# docker-compose.yml
redis-master:
  image: redis:7-alpine
  command: redis-server --appendonly yes

redis-replica:
  image: redis:7-alpine
  command: redis-server --replicaof redis-master 6379 --appendonly yes

redis-sentinel:
  image: redis:7-alpine
  command: redis-sentinel /etc/redis/sentinel.conf
```

I servizi Go usano il client con Sentinel: failover automatico in caso di down del master.

### PostgreSQL Read Replica

```yaml
postgres-auth-primary:
  image: postgres:16-alpine

postgres-auth-replica:
  image: postgres:16-alpine
  environment:
    POSTGRES_PRIMARY_HOST: postgres-auth-primary
```

Le query di lettura (GetUser, ListPayments) usano la replica. Le scritture usano il primary.

### NATS Cluster (3 nodi)

```yaml
nats-1:
  image: nats:2-alpine
  command: -cluster nats://nats-1:6222 -routes nats://nats-2:6222,nats://nats-3:6222

nats-2:
  image: nats:2-alpine
  command: -cluster nats://nats-2:6222 -routes nats://nats-1:6222,nats://nats-3:6222

nats-3:
  image: nats:2-alpine
  command: -cluster nats://nats-3:6222 -routes nats://nats-1:6222,nats://nats-2:6222
```

### Circuit Breaker

Chiamate gRPC tra servizi wrapped con circuit breaker (`sony/gobreaker`):
- 5 fallimenti in 10s → circuito aperto
- Retry con backoff esponenziale: 100ms, 200ms, 400ms, 800ms
- Timeout per chiamata gRPC: 5s

---

## Backup e Recovery

### PostgreSQL — backup giornaliero

```bash
# job-service: CronJob giornaliero alle 03:00
pg_dump $DATABASE_URL | gzip | \
  aws s3 cp - s3://your-bucket/backups/$(date +%Y%m%d)/postgres-auth.sql.gz

# Retention: 30 giorni giornalieri, 12 mesi mensili
```

### Redis — snapshot

```yaml
redis-master:
  command: >
    redis-server
    --appendonly yes
    --save 900 1      # snapshot ogni 900s se 1+ modifiche
    --save 300 10     # snapshot ogni 300s se 10+ modifiche
```

Snapshot copiato su S3 ogni 6 ore dal job-service.

### RTO/RPO

| Scenario | RPO (dati persi) | RTO (downtime) |
|----------|-----------------|----------------|
| Crash singolo servizio | 0 (stateless) | <30s (restart automatico) |
| Crash Redis | <1min (AOF) | <2min (Sentinel failover) |
| Crash PostgreSQL primary | <1s (replication lag) | <1min (replica promotion) |
| Perdita VPS completa | <6h (ultimo backup) | <2h (nuovo VPS + restore) |

---

## Permission Service (RBAC/ABAC)

### Dominio

```go
// Struttura permessi: Subject + Action + Resource
// "user:{userID} can payments:refund on payment:{paymentID}"

type Permission struct {
    Subject      string   // "user:{id}", "role:{name}", "tenant:{id}"
    Action       string   // "payments:create", "payments:refund", "users:read"
    Resource     string   // "payment:{id}", "payment:*", "user:{id}"
    Conditions   map[string]any  // es. {"owner": true} per owner-only access
}

type Role struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    Name        string        // "admin", "billing-manager", "read-only"
    Permissions []Permission
}

// Policy valutata a runtime
type PolicyDecision struct {
    Allowed bool
    Reason  string
}
```

### Valutazione permessi (OPA-style)

```
CheckPermission(subject, action, resource, context) → allow/deny

Ordine valutazione:
1. Super-admin → always allow
2. Explicit DENY → deny (priorita' su allow)
3. Role permissions → check match
4. Default → deny
```

### PostgreSQL Schema

```sql
CREATE TABLE roles (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    name       VARCHAR(100) NOT NULL,
    UNIQUE(tenant_id, name)
);

CREATE TABLE permissions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id  UUID NOT NULL,
    role_id    UUID REFERENCES roles(id) ON DELETE CASCADE,
    action     VARCHAR(100) NOT NULL,   -- "payments:refund"
    resource   VARCHAR(200) NOT NULL,   -- "payment:*" o "payment:{id}"
    effect     VARCHAR(10) NOT NULL DEFAULT 'allow',  -- allow | deny
    conditions JSONB
);

CREATE TABLE user_roles (
    user_id    UUID NOT NULL,
    role_id    UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    tenant_id  UUID NOT NULL,
    PRIMARY KEY(user_id, role_id)
);
```

### Ruoli predefiniti (per ogni nuovo tenant)

| Ruolo | Permessi |
|-------|---------|
| `owner` | `*:*` sul proprio tenant |
| `admin` | `users:*`, `payments:read`, `payments:refund` |
| `billing-manager` | `payments:*`, `subscriptions:*` |
| `read-only` | `*.read` su tutto |
| `user` | `payments:create`, `payments:read` (solo propri) |

### gRPC Interface

```protobuf
service PermissionService {
  rpc CheckPermission(CheckRequest) returns (CheckResponse);
  rpc GetUserPermissions(GetUserPermissionsRequest) returns (PermissionsResponse);
  rpc AssignRole(AssignRoleRequest) returns (google.protobuf.Empty);
  rpc RevokeRole(RevokeRoleRequest) returns (google.protobuf.Empty);
}
```

Ogni servizio chiama `CheckPermission` prima di eseguire azioni sensibili. Il gateway lo chiama per operazioni `/admin/*`.

### Caching permessi

Permessi cached in Redis per 5 minuti per (userID, tenantID). Invalidati quando un ruolo viene modificato. Evita una chiamata gRPC per ogni request.

---

## Config Service (Feature Flags per Tenant)

### Dominio

```go
type TenantConfig struct {
    TenantID uuid.UUID
    // Auth
    MFARequired          bool
    MFAMethods           []string  // ["totp", "sms"]
    EmailVerificationRequired bool
    PasswordMinLength    int
    PasswordRequireSymbols bool
    SessionMaxDevices    int       // max dispositivi simultanei
    // OAuth
    EnabledOAuthProviders []string // ["google", "github"]
    // Payments
    EnabledPaymentProviders []string
    MaxRefundDays           int
    // Notifications
    NotificationLocale   string   // "it", "en", "es"
    // Subscriptions
    TrialDays            int
    // Webhooks
    WebhookSigningSecret string
}
```

### Configurazioni con default globali

```
Priority: tenant_config → global_default

Se il tenant non ha impostato una chiave → usa il default globale.
Super-admin puo' cambiare i global defaults.
```

### Endpoints REST (/v1/config)

| Method | Path | Ruolo | Descrizione |
|--------|------|-------|-------------|
| GET | `/v1/config/tenant` | admin | Configurazione tenant corrente |
| PUT | `/v1/config/tenant` | admin | Aggiorna configurazione |
| GET | `/v1/config/tenant/defaults` | admin | Mostra i default globali |
| POST | `/v1/config/tenant/reset` | admin | Reset a default globali |
| GET | `/v1/config/features` | any | Feature flags pubblici (MFA required, etc.) |

### Cache Strategy

```
Redis key: config:{tenantID}   TTL: 5min
Invalidato su ogni PUT /v1/config/tenant
Config-service pubblica config.updated su NATS → tutti i servizi invalidano la loro cache locale
```

---

## Webhook Service (Outbound Delivery)

I tenant registrano endpoint HTTP che ricevono eventi quando accadono cose nel sistema.

### Dominio

```go
type WebhookEndpoint struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    URL          string
    Secret       string      // HMAC signing secret
    Events       []string    // ["payment.succeeded", "payment.failed"] o ["*"]
    Active       bool
    CreatedAt    time.Time
}

type WebhookDelivery struct {
    ID           uuid.UUID
    TenantID     uuid.UUID
    EndpointID   uuid.UUID
    Event        string
    Payload      json.RawMessage
    Status       string      // pending, success, failed
    AttemptCount int
    LastAttemptAt *time.Time
    ResponseCode  *int
    ResponseBody  *string
    NextRetryAt  *time.Time
}
```

### Delivery con firma HMAC

```go
// Firma ogni payload come Stripe
signature := hmac.New(sha256.New, []byte(endpoint.Secret))
signature.Write([]byte(fmt.Sprintf("%d.%s", timestamp, payload)))
header := fmt.Sprintf("t=%d,v1=%x", timestamp, signature.Sum(nil))
req.Header.Set("X-Webhook-Signature", header)
req.Header.Set("X-Webhook-Event", event)
req.Header.Set("X-Webhook-ID", delivery.ID.String())
```

### Retry Strategy

```
Tentativo 1: immediato
Tentativo 2: +5 minuti
Tentativo 3: +30 minuti
Tentativo 4: +2 ore
Tentativo 5: +8 ore
Dopo 5 tentativi → status "failed", alert admin, delivery loggata per debug
```

### Consumer NATS

```go
// Subscribes a tutti gli eventi, filtra per endpoint registrati del tenant
js.Subscribe(">", h.HandleEvent, nats.Durable("webhook-service"))
```

### Endpoints REST (/v1/webhooks)

| Method | Path | Auth | Descrizione |
|--------|------|------|-------------|
| POST | `/v1/webhooks/endpoints` | Bearer+admin | Registra endpoint |
| GET | `/v1/webhooks/endpoints` | Bearer+admin | Lista endpoints |
| PUT | `/v1/webhooks/endpoints/:id` | Bearer+admin | Aggiorna endpoint |
| DELETE | `/v1/webhooks/endpoints/:id` | Bearer+admin | Elimina endpoint |
| POST | `/v1/webhooks/endpoints/:id/test` | Bearer+admin | Invia evento di test |
| GET | `/v1/webhooks/deliveries` | Bearer+admin | Log consegne |
| POST | `/v1/webhooks/deliveries/:id/retry` | Bearer+admin | Forza retry manuale |

---

## Storage Service (File Storage S3-Compatible)

### Dominio

```go
type File struct {
    ID          uuid.UUID
    TenantID    uuid.UUID
    UserID      uuid.UUID
    Bucket      string      // "avatars", "invoices", "exports"
    Key         string      // path nel bucket
    FileName    string
    ContentType string
    Size        int64
    Public      bool
    CreatedAt   time.Time
}
```

### Bucket Strategy

| Bucket | Accesso | TTL presigned URL | Descrizione |
|--------|---------|-------------------|-------------|
| `avatars` | Public | N/A | Avatar utenti |
| `invoices` | Private | 1h | Ricevute PDF pagamenti |
| `exports` | Private | 15min | Export CSV audit/pagamenti |
| `uploads` | Private | 15min | Upload temporanei |

### Upload Flow (presigned URL)

```
1. Client: POST /v1/files/presigned {"filename": "invoice.pdf", "bucket": "invoices"}
2. Storage-service: genera presigned PUT URL MinIO (TTL 15min)
3. Client: PUT direttamente su MinIO (non passa per il servizio)
4. Client: POST /v1/files/confirm {"key": "..."}  → registra metadata su PostgreSQL
```

Nessun file passa attraverso il servizio — riduce latenza e banda.

### Generazione PDF Ricevute

```go
// job-service: triggered da payment.succeeded
// Genera PDF con template HTML → wkhtmltopdf o chromium headless
// Carica su MinIO bucket "invoices"
// Pubblica invoice.created con presigned URL
// notification-service invia email con link download
```

### Endpoints REST (/v1/files)

| Method | Path | Auth | Descrizione |
|--------|------|------|-------------|
| POST | `/v1/files/presigned` | Bearer | Genera presigned upload URL |
| POST | `/v1/files/confirm` | Bearer | Conferma upload completato |
| GET | `/v1/files/:id` | Bearer | Metadata file |
| GET | `/v1/files/:id/download` | Bearer | Redirect a presigned download URL |
| DELETE | `/v1/files/:id` | Bearer | Elimina file |

---

## Subscription Service (Pagamenti Ricorrenti)

### Dominio

```go
type Plan struct {
    ID          uuid.UUID
    TenantID    uuid.UUID    // NULL = piano globale
    Name        string       // "Starter", "Pro", "Enterprise"
    Amount      int64        // centesimi/periodo
    Currency    string
    Interval    string       // "month", "year"
    TrialDays   int
    Features    map[string]any
}

type Subscription struct {
    ID                 uuid.UUID
    TenantID           uuid.UUID
    UserID             uuid.UUID
    PlanID             uuid.UUID
    Status             SubscriptionStatus
    CurrentPeriodStart time.Time
    CurrentPeriodEnd   time.Time
    TrialEnd           *time.Time
    CancelAtPeriodEnd  bool
    ProviderID         string    // Stripe subscription ID
    CreatedAt          time.Time
}

type SubscriptionStatus string
const (
    StatusTrialing   SubscriptionStatus = "trialing"
    StatusActive     SubscriptionStatus = "active"
    StatusPastDue    SubscriptionStatus = "past_due"   // pagamento fallito
    StatusCanceled   SubscriptionStatus = "canceled"
    StatusUnpaid     SubscriptionStatus = "unpaid"     // dopo grace period
)
```

### Lifecycle

```
trialing → active          (trial scaduto, primo pagamento ok)
active   → past_due        (rinnovo fallito)
past_due → active          (pagamento riuscito nel grace period di 3gg)
past_due → unpaid          (grace period scaduto)
active   → canceled        (cancel_at_period_end = true, fine periodo)
any      → canceled        (cancellazione immediata)
```

### Webhook Stripe per rinnovi

```
invoice.paid              → rinnova subscription, aggiorna current_period_end
invoice.payment_failed    → imposta past_due, pubblica subscription.past_due
customer.subscription.deleted → imposta canceled
```

### Endpoints REST (/v1/subscriptions)

| Method | Path | Auth | Descrizione |
|--------|------|------|-------------|
| GET | `/v1/subscriptions/plans` | No | Lista piani disponibili |
| POST | `/v1/subscriptions` | Bearer | Sottoscrivi piano |
| GET | `/v1/subscriptions/current` | Bearer | Subscription corrente |
| POST | `/v1/subscriptions/cancel` | Bearer | Cancella (end of period) |
| POST | `/v1/subscriptions/reactivate` | Bearer | Riattiva prima di scadenza |
| POST | `/v1/subscriptions/upgrade` | Bearer | Cambia piano (con proration) |
| GET | `/v1/subscriptions/invoices` | Bearer | Lista fatture |
| GET | `/v1/subscriptions/invoices/:id` | Bearer | Dettaglio fattura + PDF link |

---

## Caching Strategy

### Livelli di cache

```
L1: In-process (sync.Map / ristretto)
    - Durata: 30s
    - Dati: config tenant, ruoli utente, feature flags
    - Invalidazione: NATS event config.updated, permission.changed

L2: Redis
    - Durata: 5min default, configurabile per tipo
    - Dati: permessi, config tenant, session info
    - Invalidazione: publish su Redis channel "cache:invalidate:{key}"

L3: HTTP (ETags)
    - GET /v1/payments/:id, /v1/subscriptions/current
    - ETag = hash(payload + updated_at)
    - 304 Not Modified se ETag corrisponde
```

### Cache keys standard

```
perm:{tenantID}:{userID}               TTL: 5min
config:{tenantID}                      TTL: 5min
session:{sessionID}                    TTL: 1h
plan:list:{tenantID}                   TTL: 1h  (cambiano raramente)
```

### Cache Aside Pattern

```go
func (r *CachedPaymentRepo) Get(ctx context.Context, id uuid.UUID) (*Payment, error) {
    key := fmt.Sprintf("payment:%s", id)
    if cached, err := r.redis.Get(ctx, key).Result(); err == nil {
        var p Payment
        json.Unmarshal([]byte(cached), &p)
        return &p, nil
    }
    p, err := r.db.Get(ctx, id)
    if err != nil { return nil, err }
    r.redis.Set(ctx, key, p, 5*time.Minute)
    return p, nil
}
```

---

## Internazionalizzazione (i18n)

### Locale per utente e tenant

```
Priority: user.locale → tenant.default_locale → "en"
```

Il JWT include il claim `locale`. Ogni servizio lo usa per formattare risposte e messaggi d'errore.

### Template notifiche multilingua

```
shared/i18n/
├── templates/
│   ├── welcome/
│   │   ├── en.html
│   │   ├── it.html
│   │   └── es.html
│   ├── payment_receipt/
│   │   ├── en.html
│   │   ├── it.html
│   │   └── es.html
│   └── password_reset/
│       ├── en.html
│       └── it.html
└── translations/
    ├── en.json    # {"email_not_verified": "Please verify your email"}
    ├── it.json    # {"email_not_verified": "Verifica la tua email"}
    └── es.json
```

### Errori tradotti

```go
// Gli error code sono language-neutral (es. "EMAIL_NOT_VERIFIED")
// Il client traduce usando il proprio locale
// Oppure: Accept-Language header → server restituisce messaggio tradotto
```

### Formattazione valute e date

```go
// shared/i18n/format.go
func FormatAmount(amount int64, currency, locale string) string
func FormatDate(t time.Time, locale string) string
// "1000 EUR" → "€10,00" (it) | "$10.00" (en-US) | "10,00 €" (de)
```

---

## Paginazione Cursor-Based (Standard Condiviso)

Tutti gli endpoint di lista usano cursor-based pagination — non offset. Scala su milioni di record senza degradazione delle performance.

### Request

```
GET /v1/payments?limit=20&cursor=eyJpZCI6InV1aWQiLCJjcmVhdGVkX2F0IjoiMjAyNi0wMS0wMSJ9
```

### Response

```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6Ii4uLiJ9",
    "prev_cursor": "eyJpZCI6Ii4uLiJ9",
    "has_next": true,
    "has_prev": false,
    "limit": 20
  }
}
```

### Implementazione

```go
// shared/pagination/cursor.go
type Cursor struct {
    ID        uuid.UUID  `json:"id"`
    CreatedAt time.Time  `json:"created_at"`
}

func Encode(c Cursor) string     // base64url(json(cursor))
func Decode(s string) (Cursor, error)

// SQL generato:
// WHERE (created_at, id) < ($cursor_created_at, $cursor_id)
// ORDER BY created_at DESC, id DESC
// LIMIT $limit + 1  -- +1 per sapere se esiste next page
```

---

## SDK Go Client (Riusabilita')

Package Go standalone riusabile in qualsiasi progetto. Wrappa i client gRPC con retry, circuit breaker, e tracing gia' integrati.

```go
// Uso dall'esterno in qualsiasi app Go:
import "github.com/yourorg/golang-modules/sdk/go/auth"

client, err := auth.NewClient(auth.Config{
    Addr:        "auth-service:9091",
    TLSCertPath: "/certs/client.crt",
    Timeout:     5 * time.Second,
})

// ValidateToken con retry automatico
result, err := client.ValidateToken(ctx, token)

// CheckPermission
allowed, err := client.CheckPermission(ctx, auth.PermissionCheck{
    UserID:   userID,
    Action:   "payments:refund",
    Resource: "payment:" + paymentID.String(),
})
```

### Funzionalita' del SDK

- **Retry automatico** con backoff esponenziale (3 tentativi, jitter)
- **Circuit breaker** integrato (`sony/gobreaker`)
- **Trace propagation** OTEL automatica
- **Connection pooling** configurabile
- **Health check** del servizio remoto
- Versione del SDK allineata alle versioni dei servizi via semantic versioning

### Packages SDK

```
sdk/go/
├── auth/        # ValidateToken, GetUser, CheckPermission
├── payment/     # GetPayment, ListPayments
├── permission/  # CheckPermission, GetUserRoles
├── config/      # GetTenantConfig, GetFeatureFlag
└── common/      # ErrorTypes, Pagination, Context helpers
```
