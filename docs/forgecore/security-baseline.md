# ForgeCore Security Baseline

Every ForgeCore service must satisfy this baseline.

Current status: ForgeCore has security-aware foundations, but the OWASP posture is not yet fully security-verified. Phase 11 tracks the work required to make controls executable through tests, CI checks and runtime verification.

Detailed OWASP hardening plan:

```text
docs/forgecore/owasp-security-hardening.md
```

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
| `forgecore-gateway` | token bypass, header spoofing, rate abuse | auth middleware, request id, CORS, rate limit |
| `forgecore-auth` | credential theft, token misuse, account takeover | PII encryption, password hashing, MFA, session revocation |
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
