# ForgeCore

![ForgeCore](./coreforge.png)

ForgeCore is a backend SDK and microservices foundation for building production-grade, multi-tenant Go systems faster.

It exists to remove the repetitive backend work every serious service platform needs: configuration loading, tenant isolation, validation, encrypted PII handling, pagination, typed events, observability, service clients, PostgreSQL RLS conventions, and operational guardrails.

Instead of rebuilding these pieces inside every service, ForgeCore gives you a coherent set of shared packages and reference services that can be reused as a backend platform layer.

## Why ForgeCore Exists

Modern backend services often repeat the same infrastructure code:

- environment and YAML configuration
- tenant context propagation
- PostgreSQL row-level security setup
- request IDs and auth claims
- validation and structured errors
- NATS event publishing
- cursor pagination
- PII encryption and hashing
- Prometheus metrics and JSON logs
- retry-aware service clients
- migration and boundary checks

ForgeCore turns those concerns into reusable, tested building blocks. The intent is to help teams start new services with the boring but important foundations already solved.

## What You Can Build With It

Use ForgeCore when you need backend services that are:

- multi-tenant by default
- observable from day one
- consistent across service boundaries
- ready for PostgreSQL RLS
- event-driven with typed NATS payloads
- configurable through YAML and environment variables
- organized with DDD-style boundaries
- fast to scaffold and hard to accidentally wire incorrectly

## Repository Layout

```text
shared/                 Core SDK packages reused by services
sdk/go/                 Go client SDK for internal service calls
services/               ForgeCore reference microservices
deployments/            Prometheus, Grafana, Traefik, OTEL, alerting
scripts/                Verification and operational scripts
docs/forgecore/         Compatibility and client process docs
docker-compose.yml      Local development stack
```

## Run ForgeCore For A Frontend

A frontend should integrate with ForgeCore through `forgecore-gateway`.

The gateway is the public HTTP entrypoint. It forwards requests to the internal services, applies request IDs, CORS, security headers, rate limiting and authentication checks.

Default local frontend API base URL:

```text
http://localhost:8080
```

Recommended first local setup:

```powershell
Copy-Item .env.example .env
make smoke
docker compose up -d postgres redis nats minio prometheus grafana
```

Run database migrations from Git Bash, WSL or another Bash shell with the `migrate` CLI installed:

```bash
./scripts/migrate.sh all up
```

Bootstrap local dependencies and seed a first admin user:

```bash
SUPERADMIN_EMAIL=admin@forgecore.local SUPERADMIN_PASSWORD=ChangeMe123! ./scripts/bootstrap.sh
```

Then start the backend services:

```powershell
docker compose up -d forgecore-auth forgecore-gateway
```

For the full local platform:

```powershell
docker compose up -d
```

Check the gateway:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/readyz
```

For frontend development, set `CORS_ORIGIN` in `.env` to the origin used by your app:

```env
CORS_ORIGIN=http://localhost:3000,http://localhost:5173
```

Frontend requests should use:

- base URL: `http://localhost:8080`
- `Authorization: Bearer <access-token>` for protected routes
- `X-Tenant-ID: <tenant-id>` when the tenant is not fully derived from the token
- `X-Request-ID: <uuid-or-request-id>` for correlation
- `Idempotency-Key: <stable-key>` for retried mutations such as payments and subscriptions

Public routes currently exposed through the gateway:

- `GET /healthz`
- `GET /readyz`
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/forgot-password`
- `POST /v1/auth/reset-password`
- `POST /v1/auth/verify-email`
- `POST /v1/auth/resend-verification`
- `GET /v1/auth/oauth/{provider}`
- `GET /v1/auth/oauth/{provider}/callback`

The frontend-facing contract starts here:

- `docs/forgecore/frontend-api-readiness.md`
- `docs/forgecore/openapi/forgecore-gateway.v1.yaml`

Run the gateway E2E smoke test:

```powershell
make test-e2e
```

## Services

ForgeCore service names use the pattern:

```text
forgecore-<bounded-context>
```

Current services:

| Service | Public through gateway | Local REST | Local gRPC | Main dependencies | Purpose |
| --- | --- | ---: | ---: | --- | --- |
| `forgecore-gateway` | yes | 8080 | - | auth gRPC | Public API entrypoint for frontends |
| `forgecore-auth` | yes, under `/v1/auth/*` | 8081 | 9091 | PostgreSQL, Redis | Users, sessions, JWT, MFA, OAuth |
| `forgecore-payments` | protected, under `/v1/payments/*` | 8082 | 9092 | PostgreSQL, NATS, Stripe | Payments and Stripe webhooks |
| `forgecore-notifications` | protected, under `/v1/notifications/*` | 8083 | - | PostgreSQL, NATS, SendGrid | Email/notification delivery |
| `forgecore-admin` | protected, under `/v1/admin/*` | 8084 | - | PostgreSQL | Backoffice API |
| `forgecore-audit` | protected, under `/v1/audit/*` | 8085 | 9095 | PostgreSQL, NATS | Append-only audit trail |
| `forgecore-jobs` | no direct frontend route | - | - | Redis, NATS | Async jobs and schedulers |
| `forgecore-permissions` | protected, under `/v1/permissions/*` | 8087 | 9097 | PostgreSQL | Roles, permissions and policies |
| `forgecore-config` | protected, under `/v1/config/*` | 8088 | 9098 | PostgreSQL, Redis | Runtime and tenant configuration |
| `forgecore-webhooks` | protected/admin, under `/v1/webhooks/*` | 8089 | - | PostgreSQL, NATS | Outbound webhook endpoints and deliveries |
| `forgecore-storage` | protected, under `/v1/storage/*` | 8090 | - | PostgreSQL, MinIO | File metadata and object storage |
| `forgecore-subscriptions` | protected, under `/v1/subscriptions/*` | 8091 | 9099 | PostgreSQL, Stripe | Plans and subscriptions |

The frontend should not call service ports directly in normal use. Use `forgecore-gateway` and let the gateway route to the correct bounded context.

## Shared SDK Packages

The `shared/` module contains the reusable backend SDK:

| Package | Purpose |
| --- | --- |
| `apperrors` | Application error codes, wrapping, and HTTP mapping |
| `configloader` | Loads default, YAML, and ENV configuration |
| `configschema` | Defines config keys, types, defaults, required fields, and secrets |
| `configsource` | Config sources for maps, YAML files, and ENV |
| `crypto` | AES-256-GCM PII encryption and HMAC lookup hashing |
| `events` | Typed, versioned NATS event payloads and publisher |
| `i18n` | Amount/date formatting helpers |
| `middleware` | Tenant, auth claims, request ID, and public header constants |
| `observability` | JSON logger, Prometheus metrics, OTLP tracing, shutdown helpers |
| `pagination` | Cursor encode/decode and limit normalization |
| `postgres` | Tenant-aware PostgreSQL transaction helper |
| `proto` | gRPC contracts |
| `validation` | Validator wrapper and field-level validation errors |

## Configuration

ForgeCore services read configuration through the shared SDK:

```text
default values < YAML file < environment variables
```

Set an optional YAML file with:

```powershell
$env:FORGECORE_CONFIG_FILE = "C:\path\to\forgecore.yaml"
```

Example YAML:

```yaml
port: ":8088"
database_url: "postgres://postgres:postgres@localhost:5432/config?sslmode=disable"
redis_addr: "localhost:6379"
```

Environment variables always win over YAML values.

For local Docker Compose, copy `.env.example` to `.env`. The most important variables are:

| Variable | Used by | Meaning |
| --- | --- | --- |
| `DATABASE_URL` | runtime services | PostgreSQL URL inside the Compose network |
| `*_DATABASE_URL` | migration script | PostgreSQL URLs from the host machine |
| `REDIS_ADDR` | auth, config, jobs | Redis address inside the Compose network |
| `NATS_URL` | event-driven services | NATS URL inside the Compose network |
| `CORS_ORIGIN` | gateway | Comma-separated frontend origins |
| `AUTH_GRPC_ADDR` | gateway | Internal gRPC address of `forgecore-auth` |
| `*_SERVICE_URL` | gateway | Internal HTTP upstream URLs |
| `STRIPE_SECRET_KEY` | payments, subscriptions | Stripe API secret |
| `STRIPE_WEBHOOK_SECRET` | payments | Stripe webhook signing secret |
| `SENDGRID_API_KEY` | notifications | SendGrid API key |
| `MINIO_ENDPOINT` | storage | MinIO endpoint |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | storage | MinIO credentials |
| `SUPERADMIN_EMAIL` / `SUPERADMIN_PASSWORD` | bootstrap | Local seed admin |

`DATABASE_URL` is intentionally shared by the current local stack: every service uses the same PostgreSQL database with tenant-aware tables and RLS policies. The migration script also accepts service-specific URLs, so the same project can later run with one database per service:

```bash
AUTH_DATABASE_URL=postgres://... ./scripts/migrate.sh forgecore-auth up
PAYMENT_DATABASE_URL=postgres://... ./scripts/migrate.sh forgecore-payments up
```

Secrets should be declared in schemas with `Secret: true` and read through:

```go
values.Secret("STRIPE_SECRET_KEY").Value()
```

This keeps the string representation redacted and reduces accidental secret logging.

## Using ForgeCore In Your Service

Add the shared module to your service and use the SDK packages directly.

Example configuration schema:

```go
package main

import (
    "context"

    "github.com/Andrea-Cavallo/golang-modules/shared/configloader"
    "github.com/Andrea-Cavallo/golang-modules/shared/configschema"
)

const (
    keyPort        = "PORT"
    keyDatabaseURL = "DATABASE_URL"
)

var schema = configschema.Schema{
    {Key: keyPort, Default: ":8080", Kind: configschema.String},
    {Key: keyDatabaseURL, Required: true, Kind: configschema.String},
}

func load(ctx context.Context) error {
    values, err := configloader.NewDefault(schema).Load(ctx)
    if err != nil {
        return err
    }
    _ = values.String(keyPort)
    return nil
}
```

Example tenant-aware PostgreSQL transaction:

```go
err := postgres.WithTenantTx(ctx, pool, tenantID, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO records (tenant_id, name) VALUES ($1, $2)", tenantID, name)
    return err
})
```

Example versioned event:

```go
event := events.PaymentSucceeded{
    Version:       events.EventVersionV1,
    EventName:     events.EventPaymentSucceeded,
    CorrelationID: requestID,
    TenantID:      tenantID,
    UserID:        userID,
    PaymentID:     paymentID,
    Amount:        amount,
    Currency:      "EUR",
    Provider:      "stripe",
    OccurredAt:    time.Now().UTC(),
}
```

## Architecture Rules

Each service follows this dependency direction:

```text
transport -> application -> domain
infrastructure -> domain
```

Rules:

- `domain/` contains entities, value objects, and repository interfaces.
- `application/` contains use cases and depends only on domain and application ports.
- `infrastructure/` implements repositories and external providers.
- `transport/` contains REST, gRPC, NATS consumers, and handlers.
- No infrastructure imports are allowed in `domain/` or `application/`.

Check boundaries with:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-boundaries.ps1
```

## Multi-Tenancy

Every tenant-aware table must include:

- `tenant_id UUID NOT NULL`
- an index on `tenant_id`
- PostgreSQL row-level security
- a `tenant_isolation` policy

Verify migrations with:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-tenant-migrations.ps1
```

## Compatibility

Compatibility rules live in:

```text
docs/forgecore/compatibility-matrix.md
```

Highlights:

- proto packages use `*.v1`
- NATS event names use `.v1`
- existing proto field numbers must not be reused
- existing event fields must not be removed without a new event version
- database migrations are append-only

## Verification

Run these before considering a change complete:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-boundaries.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-proto-contracts.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-sdk-clients.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-tenant-migrations.ps1
```

Build every Go module:

```powershell
$mods = Get-ChildItem -Recurse -File -Filter go.mod
foreach ($mod in $mods) {
    Push-Location (Split-Path $mod.FullName -Parent)
    go build ./...
    Pop-Location
}
```

Run shared SDK tests:

```powershell
cd shared
go test ./...
```

## Local Operations

Create a local environment file first:

```powershell
Copy-Item .env.example .env
```

Validate local Compose configuration and static checks:

```powershell
make smoke
```

Start the local stack:

```powershell
docker compose up -d postgres redis nats minio prometheus grafana
```

Run all migrations:

```bash
./scripts/migrate.sh all up
```

Bootstrap local dependencies:

```bash
./scripts/bootstrap.sh
```

Operational docs:

- `docs/forgecore/reliability-patterns.md`
- `docs/forgecore/security-baseline.md`
- `docs/forgecore/owasp-security-hardening.md`
- `docs/forgecore/runbooks/tenant.md`
- `docs/forgecore/runbooks/webhooks.md`
- `docs/forgecore/runbooks/jobs.md`
- `docs/forgecore/runbooks/payments.md`
- `docs/forgecore/runbooks/audit.md`
- `docs/forgecore/runbooks/storage.md`

## Official Commands

ForgeCore includes a Makefile for common workflows:

```powershell
make verify
make build
make test-shared
make test-sdk
make test-auth
make test-e2e
make smoke
make scaffold-dryrun name=forgecore-example
make scaffold name=forgecore-example
```

PowerShell scripts are available directly as well:

- `scripts/build-all.ps1`
- `scripts/check-boundaries.ps1`
- `scripts/check-proto-contracts.ps1`
- `scripts/check-sdk-clients.ps1`
- `scripts/check-tenant-migrations.ps1`
- `scripts/e2e-gateway.ps1`
- `scripts/scaffold-service.ps1` with `-DryRun` support
- `scripts/smoke-local.ps1`

## Frontend Integration

Frontend applications should call `forgecore-gateway` as the public API entrypoint. The practical setup flow is documented in `Run ForgeCore For A Frontend` above.

The frontend-facing contract is documented in:

- `docs/forgecore/frontend-api-readiness.md`
- `docs/forgecore/openapi/forgecore-gateway.v1.yaml`

Gateway E2E coverage verifies CORS preflight, security headers, health/readiness and public auth proxying:

```powershell
make test-e2e
```

## Releases

Release policy:

- changelog: `CHANGELOG.md`
- release strategy: `docs/forgecore/release.md`
- architecture decisions: `docs/forgecore/adr/`

Breaking changes must update the compatibility matrix and changelog.

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md) before opening a PR.
