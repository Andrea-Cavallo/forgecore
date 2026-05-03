# OWASP Endpoint Map

This map is the starting point for endpoint-by-endpoint security verification.

Legend:

- A01: Broken Access Control
- A02: Cryptographic Failures
- A03: Injection
- A04: Insecure Design
- A05: Security Misconfiguration
- A06: Vulnerable and Outdated Components
- A07: Identification and Authentication Failures
- A08: Software and Data Integrity Failures
- A09: Security Logging and Monitoring Failures
- A10: Server-Side Request Forgery

| Surface | Endpoint | Auth | Main OWASP Risks | Required Controls |
| --- | --- | --- | --- | --- |
| gateway | `GET /healthz` | public | A05 | no secrets, minimal JSON |
| gateway | `GET /readyz` | public | A05 | readiness only, no sensitive dependency details |
| auth | `POST /v1/auth/register` | public | A03, A07, A09 | validation, password hashing, tenant binding, audit |
| auth | `POST /v1/auth/login` | public | A07, A09 | rate limit, MFA flow, session tracking, audit |
| auth | `POST /v1/auth/refresh` | public/token | A07 | refresh rotation, revocation check |
| auth | `POST /v1/auth/forgot-password` | public | A07 | non-enumerating response, rate limit |
| auth | `POST /v1/auth/reset-password` | public/token | A07 | one-time token, password policy |
| auth | `POST /v1/auth/verify-email` | public/token | A07 | one-time token |
| payments | `POST /v1/payments` | protected | A01, A04, A09 | RBAC, tenant isolation, idempotency, audit |
| payments | `POST /v1/payments/{id}/refund` | protected | A01, A04, A09 | RBAC, idempotency, provider audit |
| payments | `POST /v1/webhooks/stripe` | signature | A04, A08 | Stripe signature, replay/idempotency |
| permissions | `/v1/permissions/*` | protected | A01 | RBAC, tenant scope |
| config | `/v1/config/*` | protected | A01, A05, A09 | RBAC, schema validation, secret redaction, audit |
| webhooks | `/v1/webhooks/*` | protected | A01, A10 | RBAC, SSRF prevention, HMAC signing |
| storage | `/v1/storage/*` | protected | A01, A02 | tenant metadata, presigned expiry |
| subscriptions | `/v1/subscriptions/*` | protected | A01, A04, A09 | RBAC, idempotency, provider reconciliation |
| audit | `/v1/audit/*` | protected | A01, A09 | append-only reads, admin role |
| admin | `/v1/admin/*` | protected | A01, A09 | admin role, mandatory audit |

Every protected endpoint must have a test proving unauthenticated access fails and unauthorized role access fails.

RBAC is enforced in `forgecore-gateway/internal/middleware/rbac.go` and verified against `docs/forgecore/rbac-endpoint-matrix.md` by `scripts/check-rbac-security.ps1`.

Security-sensitive mutating endpoints are logged by `forgecore-gateway/internal/middleware/audit.go` with service, method, path, tenant id, user id and request id.
