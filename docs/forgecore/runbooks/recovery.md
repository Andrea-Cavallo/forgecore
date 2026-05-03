# Recovery Runbook

Use this runbook when ForgeCore runtime hardening checks detect degraded services.

## Health And Readiness

Check every public service through:

```powershell
Invoke-RestMethod http://localhost:<port>/healthz
Invoke-RestMethod http://localhost:<port>/readyz
```

Expected response:

```json
{
  "service": "forgecore-payments",
  "status": "ok",
  "checks": {
    "postgres": "ok",
    "nats": "ok"
  }
}
```

## PostgreSQL Recovery

1. Verify container/process status.
2. Run `pg_isready`.
3. Check connection strings in `.env`.
4. Verify migrations with `scripts/check-tenant-migrations.ps1`.
5. Re-run service readiness.

## Redis Recovery

1. Run `redis-cli ping`.
2. Check `REDIS_ADDR`.
3. Restart dependent services: `forgecore-auth`, `forgecore-config`, `forgecore-jobs`.
4. Re-run readiness.

## NATS Recovery

1. Check `http://localhost:8222/healthz`.
2. Verify JetStream streams with `scripts/nats-init.sh`.
3. Restart event consumers: audit and notifications.
4. Inspect outbox backlog before replaying events.

## Outbox Recovery

1. Query pending messages by tenant and subject.
2. Confirm downstream service health.
3. Re-dispatch in bounded batches.
4. Mark permanently failed messages only after operator review.

## Idempotency Recovery

1. Search by tenant, operation and idempotency key.
2. If status is `completed`, return the stored response.
3. If status is `started` and stale, mark failed before retry.
4. Never execute the same provider mutation twice for a conflicting fingerprint.

## Security Checks

Run:

```powershell
make security-check
```

If `govulncheck` is missing locally, install it:

```powershell
go install golang.org/x/vuln/cmd/govulncheck@latest
```
