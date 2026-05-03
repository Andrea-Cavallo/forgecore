# Jobs Runbook

## Symptoms

- jobs repeat side effects
- cleanup does not run
- scheduler exits early

## Checks

1. Verify Redis connectivity.
2. Check job type registration in `forgecore-jobs`.
3. Confirm handler is idempotent.
4. Inspect retry count and payload tenant id.

## Recovery

- Re-dispatch with the same idempotency key.
- Fix handler registration before increasing retry count.
- Keep job payloads small and versioned.
