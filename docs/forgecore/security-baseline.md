# ForgeCore Security Baseline

Every ForgeCore service must satisfy this baseline.

Current status: ForgeCore has executable OWASP Level 1 baseline controls for the available gateway and service surface. Remaining security work should deepen coverage, but the Phase 11 baseline is now verified by tests, static checks and CI-oriented scripts.

Detailed OWASP hardening plan:

```text
docs/forgecore/owasp-security-hardening.md
```

Current verified controls:

- gateway CORS and forbidden-origin test
- gateway missing-token JSON envelope test
- webhook SSRF rejection test
- Stripe webhook invalid-signature rejection test
- local security script and GitHub security workflow skeleton
- `govulncheck` integration when available
- container image scanning through Trivy in CI
- gateway RBAC middleware with endpoint-by-endpoint role matrix
- mandatory structured audit logging for sensitive gateway mutations
- auth application E2E test for register, login, refresh, logout, me and protected token validation
- JWT key rotation with `kid` and previous-key validation window

## Common

- Do not log secrets, tokens, PII plaintext, or payment provider secrets.
- Use `shared/configschema.Field{Secret: true}` for secret configuration.
- Use tenant-aware database access for tenant-owned data.
- Validate tenant id and user id at transport boundaries.
- Use structured errors for external responses.
- Keep logs in Italian inside service runtime code.

## Service Threat Model

| Service | Primary Risks | Required Controls |
| --- | --- | --- |
| `forgecore-gateway` | token bypass, header spoofing, rate abuse | auth middleware, request id, CORS, rate limit, RBAC matrix, audit middleware |
| `forgecore-auth` | credential theft, token misuse, account takeover | PII encryption, password hashing, MFA, session revocation, refresh token storage, JWT key rotation |
| `forgecore-payments` | duplicate charge, webhook spoofing | idempotency key, signature verification, provider event tracking |
| `forgecore-notifications` | spam, PII leakage | template allowlist, recipient validation, secret redaction |
| `forgecore-admin` | privilege escalation | permission checks, audit logging |
| `forgecore-audit` | tampering, missing evidence | append-only repository, immutable events |
| `forgecore-jobs` | repeated side effects | idempotent job handlers, bounded retries |
| `forgecore-permissions` | authorization drift | tenant-scoped roles and boundary tests |
| `forgecore-config` | unsafe runtime changes | validation schema, secret redaction, audit events |
| `forgecore-webhooks` | SSRF, duplicate delivery | URL validation, private IP block, HMAC signatures |
| `forgecore-storage` | unauthorized object access | tenant metadata, presigned URL expiry |
| `forgecore-subscriptions` | duplicate subscription, billing drift | provider idempotency, status reconciliation |
