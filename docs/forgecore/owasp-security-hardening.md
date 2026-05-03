# OWASP Security Hardening

ForgeCore is currently security-aware, not yet security-verified.

The goal of Phase 11 is to make OWASP controls executable: mapped to endpoints, covered by tests, enforced in CI and observable at runtime.

Reference targets:

- OWASP Top 10:2021
- OWASP ASVS Level 1 minimum baseline

## Current Position

| OWASP Area | Current ForgeCore Position | Target |
| --- | --- | --- |
| A01 Broken Access Control | Gateway auth, tenant/RLS patterns and permissions service exist | Endpoint-by-endpoint RBAC matrix and tests |
| A02 Cryptographic Failures | PII encryption, HMAC lookup, secret redaction and JWT patterns exist | Key rotation, secret lifecycle and verified crypto tests |
| A03 Injection | Validation wrapper and parameterized database style exist | Repository audit and injection test cases |
| A04 Insecure Design | Threat model, DDD boundaries, idempotency and outbox docs exist | Implemented controls and abuse-case tests |
| A05 Security Misconfiguration | Security headers, CORS and config SDK exist | Production config profile and misconfiguration checks |
| A06 Vulnerable Components | Builds are verified | `govulncheck`, dependency scanning and image scanning |
| A07 Identification and Authentication Failures | Auth, MFA and sessions exist | Full auth E2E: register, login, refresh, logout, me and protected call |
| A08 Software and Data Integrity Failures | Versioned events and compatibility matrix exist | CI supply-chain checks and signed release path |
| A09 Security Logging and Monitoring Failures | Audit service and observability exist | Mandatory audit logging for sensitive actions |
| A10 Server-Side Request Forgery | Webhook URL validation and private IP block exist | SSRF tests and coverage for all outbound fetch points |

## Phase 11 Checklist

- [ ] Map every gateway and service endpoint against OWASP Top 10.
- [ ] Add `govulncheck` and dependency scanning.
- [ ] Add container image scanning.
- [ ] Add security tests for CORS, auth, tenant isolation, SSRF and webhook signatures.
- [ ] Verify RBAC/permissions on every protected endpoint.
- [ ] Complete auth E2E: register, login, refresh, logout, me and protected call.
- [ ] Define frontend token storage policy.
- [ ] Define CSRF policy if cookies are used.
- [ ] Implement key rotation and secrets lifecycle.
- [ ] Add mandatory audit logging for sensitive actions.
- [ ] Raise the baseline toward OWASP ASVS Level 1.
- [ ] Turn controls into automated tests, CI security checks and runtime checks.

## Required Security Tests

| Test Area | Minimum Cases |
| --- | --- |
| CORS | allowed origin succeeds, forbidden origin fails, preflight bypasses auth |
| Auth | missing token, invalid token, expired token, revoked token, valid token |
| Tenant isolation | tenant A cannot read tenant B data, missing tenant fails |
| RBAC | user without role receives forbidden on protected action |
| SSRF | loopback/private/link-local URLs are rejected |
| Webhook signature | valid signature accepted, invalid signature rejected, replay considered |
| Secrets | secret fields are redacted in logs/config string output |
| Audit | sensitive action writes audit event with tenant/user/request metadata |

## Frontend Token Policy To Decide

ForgeCore must choose one official frontend auth mode before production:

- bearer access token in memory plus refresh token in secure HttpOnly cookie
- bearer access token and refresh token both held by the frontend
- full cookie session with CSRF token

Recommended production direction: short-lived access token in memory, refresh token in `HttpOnly`, `Secure`, `SameSite=Lax` cookie, and explicit CSRF protection for state-changing cookie-authenticated requests.

## CI Security Gates

Phase 11 should add a CI stage that runs:

```powershell
govulncheck ./...
make test-e2e
make smoke
make runtime-check
make security-check
```

Container scanning should run against each service image before publishing.

Supporting documents:

- `docs/forgecore/owasp-endpoint-map.md`
- `docs/forgecore/frontend-token-csrf-policy.md`
- `docs/forgecore/secrets-lifecycle.md`

## Definition Of Done

Phase 11 is done only when:

- every task checkbox is verified
- OWASP endpoint map exists
- CI security checks are active
- security tests cover gateway and critical services
- auth E2E covers a protected call
- README and security baseline reflect verified status, not planned status
