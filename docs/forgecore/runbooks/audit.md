# Audit Runbook

## Symptoms

- expected audit entry missing
- audit consumer stopped
- audit query returns empty for tenant

## Checks

1. Confirm NATS stream and consumer are healthy.
2. Confirm `audit_entries` has RLS and tenant policy.
3. Verify event subject starts with `audit.`.
4. Check append-only repository errors.

## Recovery

- Restart audit consumer.
- Replay events from NATS if retention allows.
- Do not update or delete audit rows.
