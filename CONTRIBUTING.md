# Contributing To ForgeCore

Thank you for helping improve ForgeCore. This project is meant to be a serious backend SDK and service foundation, so contributions should keep the platform consistent, boring in the best way, and easy to reuse.

## Core Principles

- Prefer reusable SDK primitives over duplicated service-local code.
- Keep package names specific and technical.
- Avoid generic packages such as `utils`, `common`, `helpers`, or `misc`.
- Keep service boundaries clear.
- Do not import infrastructure code from `domain/` or `application/`.
- Keep tenant isolation explicit and verified.
- Keep README and docs aligned with the real project state.

## Required Checks

Run the project checks before submitting changes:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-boundaries.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-proto-contracts.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-sdk-clients.ps1
powershell -ExecutionPolicy Bypass -File .\scripts\check-tenant-migrations.ps1
```

Build all modules:

```powershell
$mods = Get-ChildItem -Recurse -File -Filter go.mod
foreach ($mod in $mods) {
    Push-Location (Split-Path $mod.FullName -Parent)
    go build ./...
    Pop-Location
}
```

Run tests where available:

```powershell
cd shared
go test ./...

cd ..\sdk\go
go test ./...

cd ..\..\services\forgecore-auth
go test ./...
```

## Code Style

- Go target is `1.26`.
- Keep functions small and direct.
- Use guard clauses instead of deeply nested branches.
- Always handle errors explicitly.
- Do not ignore errors with `_`.
- Use named constants instead of magic strings or numbers.
- Avoid placeholder comments such as `TODO implement`.
- Logs inside services should remain in Italian, matching the existing service convention.

## Service Architecture

Every service follows:

```text
transport -> application -> domain
infrastructure -> domain
```

Layer rules:

- `domain/`: entities, value objects, repository interfaces, no external infrastructure dependencies.
- `application/`: use cases, input validation, orchestration, domain/application ports only.
- `infrastructure/`: PostgreSQL, Redis, NATS, Stripe, MinIO, SendGrid, provider implementations.
- `transport/`: REST, gRPC, NATS consumers, HTTP handlers.

## Configuration

All service configuration must go through the shared config SDK:

- `shared/configloader`
- `shared/configschema`
- `shared/configsource`

Priority order:

```text
default SDK < YAML < ENV
```

Do not add new local `envOr` helpers to services.

## Multi-Tenancy

Every tenant-aware table must include:

- `tenant_id UUID NOT NULL`
- tenant index
- RLS enabled
- `tenant_isolation` policy

Use `shared/postgres.WithTenantTx` for tenant-aware PostgreSQL transactions.

## Events And Proto

- Proto contracts live in `shared/proto`.
- Event contracts live in `shared/events`.
- Proto packages must use `.v1`.
- Event names must be versioned with `.v1`.
- Do not reuse proto field numbers.
- Do not remove existing event fields without creating a new event version.

## Documentation

Update `README.md` when a contribution changes:

- public SDK usage
- service naming
- repository structure
- commands
- compatibility rules
- operational workflow

Update `docs/forgecore/compatibility-matrix.md` when changing proto, events, schema, or SDK compatibility.
