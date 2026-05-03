# Frontend API Readiness

ForgeCore frontends should integrate through `forgecore-gateway`.

Direct calls to individual services are reserved for internal service-to-service traffic, operational tooling, or tests.

## Public Gateway Surface

Public unauthenticated routes:

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

Authenticated frontend routes must send:

- `Authorization: Bearer <access-token>`
- `X-Tenant-ID: <tenant-id>` when tenant context is not derived from the token
- `X-Request-ID: <request-id>` when the frontend already has a correlation id
- `Idempotency-Key: <key>` for retried mutations such as payments, subscriptions and webhook-facing commands

## Error Envelope

Gateway errors should use this shape:

```json
{
  "code": "invalid_token",
  "message": "token non valido",
  "request_id": "req_123"
}
```

Service transports should converge on the same shape through `shared/apperrors`.

## CORS

The gateway supports a comma-separated CORS allowlist through `CORS_ORIGIN`.

Example:

```powershell
$env:CORS_ORIGIN = "http://localhost:3000,https://app.example.com"
```

Preflight requests are handled before authentication.

## OpenAPI

The initial frontend-facing contract lives in:

```text
docs/forgecore/openapi/forgecore-gateway.v1.yaml
```

This contract is the source for future TypeScript client generation.

## E2E

Run the gateway frontend E2E smoke test:

```powershell
make test-e2e
```

The current test starts a real HTTP upstream with `httptest`, builds the gateway middleware/proxy chain, verifies CORS preflight, security headers, health/readiness and public auth proxying.
