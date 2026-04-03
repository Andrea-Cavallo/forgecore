# CLAUDE.md — Superpowers Microservices Orchestrator

## Project Goal

Build a production-ready Go 1.24 microservices monorepo called `superpowers`.
12 services, DDD architecture, Docker Compose deploy, multi-tenant PostgreSQL RLS.

Root: `C:\Users\Andrea\Desktop\golang-modules\`

---

## RULES — Read before doing anything

1. NEVER summarize. NEVER explain what you are about to do. Write code immediately.
2. Read the plans file first. Find the first `[ ]` task. Execute ONLY that task.
3. After completing a task, mark it `[x]` in the plans file. Then stop.
4. Each file must compile. No placeholder comments like `// TODO implement`.
5. All functions must be ≤50 lines. Use guard clauses, not nested if-else.
6. Always handle errors explicitly. No `_` for errors.
7. Use named constants, never magic strings or numbers.
8. Log messages in Italian (matching existing PDLD codebase convention).

Plans file: `C:\Users\Andrea\Desktop\golang-modules\docs\superpowers\plans\2026-03-30-microservices-implementation.md`
Specs file: `C:\Users\Andrea\Desktop\golang-modules\docs\superpowers\specs\2026-03-30-microservices-design.md`
Issues file: `C:\Users\Andrea\Desktop\golang-modules\docs\superpowers\issues.md`

> **IMPORTANTE**: Prima di scrivere o modificare qualsiasi file, leggi `issues.md`.
> Ogni task deve risolvere o non introdurre nuovi problemi elencati lì.
> Aggiorna `issues.md` marcando come risolti i problemi che correggi.

---

## Architecture

```
transport/ → application/ → domain/
infrastructure/ → domain/ (implements interfaces)
```

- `domain/`         — entities, value objects, repository interfaces. Zero external deps.
- `application/`    — use cases. Depends only on domain.
- `infrastructure/` — postgres, redis, external providers.
- `transport/`      — REST handlers, gRPC, NATS consumers.

---

## Multi-Tenancy Pattern (apply to every table)

```sql
CREATE TABLE <name> (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    ...
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON <name>(tenant_id);
ALTER TABLE <name> ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON <name>
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

Go middleware sets tenant on each DB connection:
```go
conn.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
```

---

## Standard File Templates

### go.mod (per service)
```
module github.com/yourorg/golang-modules/services/<service-name>

go 1.24

require (
    github.com/google/uuid v1.6.0
    github.com/jackc/pgx/v5 v5.7.0
    github.com/redis/go-redis/v9 v9.7.0
    go.uber.org/zap v1.27.0
)
```

### main.go skeleton
```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    if err := run(ctx); err != nil {
        slog.Error("avvio servizio fallito", "errore", err)
        os.Exit(1)
    }
}
```

### Repository interface (domain layer)
```go
type <Entity>Repository interface {
    Create(ctx context.Context, e *<Entity>) error
    GetByID(ctx context.Context, id uuid.UUID) (*<Entity>, error)
    Update(ctx context.Context, e *<Entity>) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*<Entity>, error)
}
```

### Use case pattern
```go
type <Action>UseCase struct {
    repo   domain.<Entity>Repository
    events events.Publisher
}

func (uc *<Action>UseCase) Execute(ctx context.Context, input <Action>Input) (*<Action>Output, error) {
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("input non valido: %w", err)
    }
    // business logic here
}
```

---

## Services Summary

| Service             | Port REST | Port gRPC | Key dependencies         |
|---------------------|-----------|-----------|--------------------------|
| api-gateway         | 8080      | -         | Traefik, auth gRPC       |
| auth-service        | 8081      | 9091      | postgres, redis, JWT RS256 |
| payment-service     | 8082      | 9092      | postgres, stripe         |
| notification-service| 8083      | -         | postgres, NATS, sendgrid |
| admin-service       | 8084      | -         | auth+payment gRPC clients|
| audit-service       | 8085      | 9095      | postgres (append-only), NATS |
| job-service         | -         | -         | redis (asynq), NATS      |
| permission-service  | 8087      | 9097      | postgres                 |
| config-service      | 8088      | 9098      | postgres, redis          |
| webhook-service     | 8089      | -         | postgres, NATS           |
| storage-service     | 8090      | -         | minio, postgres          |
| subscription-service| 8091      | 9099      | postgres, stripe         |

---

## Sub-Agent Instructions

When the orchestrator (primary Build agent) assigns a task, spawn one sub-agent per service using the Task tool. Each sub-agent must:

1. Read this CLAUDE.md file first.
2. Read only the section of the specs relevant to its assigned service.
3. Build ONLY the files for its assigned service — never touch other services.
4. Follow the DDD 4-layer structure exactly.
5. Report back: list of files created, nothing else.

### How to spawn parallel agents (primary agent prompt)

```
Spawn one sub-agent per service in parallel using the Task tool.
Assign each agent exactly one service from the list below.
Each agent must read CLAUDE.md, then build the complete DDD skeleton for its service.
Wait for all agents to complete, then update the plans file marking the task as done.

Services to assign:
- @general → auth-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → payment-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → notification-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → permission-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → config-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → audit-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → webhook-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → storage-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → subscription-service: build domain/, application/ skeleton, go.mod, Dockerfile
- @general → job-service: build jobs/, scheduler/ skeleton, go.mod, Dockerfile
- @general → admin-service: build application/, transport/ skeleton, go.mod, Dockerfile
- @general → api-gateway: build proxy/, middleware/, router/ skeleton, go.mod, Dockerfile
```

---

## Shared Packages to build first (Phase 0)

Before spawning service agents, build these shared packages:

```
shared/
├── proto/          — .proto files for all 10 services
├── events/         — NATS typed event structs
├── middleware/      — tenant context, auth check, request-id
├── validation/     — go-playground/validator wrappers
├── crypto/         — AES-256 PII encryption helpers
├── pagination/     — cursor-based pagination (Encode/Decode)
├── i18n/           — locale helpers, format amount/date
└── observability/  — slog JSON, Prometheus, graceful shutdown
```

Build shared packages sequentially (they are dependencies of all services).
Build services in parallel after shared/ is complete.

---

## Definition of Done (per task)

A task is `[x]` only when:
- [ ] All files compile (`go build ./...` passes)
- [ ] No `// TODO` placeholders remain
- [ ] Repository interface exists in domain/
- [ ] Use cases are wired to the interface
- [ ] go.mod is present and valid
- [ ] Dockerfile is present
