# Tenant Runbook

## Symptoms

- records missing for a valid tenant
- cross-tenant data not visible when expected
- PostgreSQL RLS errors

## Checks

1. Confirm `X-Tenant-ID` is present and valid UUID.
2. Confirm request context contains tenant id.
3. Confirm transaction uses `SET LOCAL app.tenant_id`.
4. Run `scripts/check-tenant-migrations.ps1`.
5. Inspect table RLS and `tenant_isolation` policy.

## Recovery

- Reject requests without tenant id.
- Re-run missing migrations.
- Use `shared/postgres.WithTenantTx` for new tenant-aware writes.
