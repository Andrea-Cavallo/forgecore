# Storage Runbook

## Symptoms

- presigned URL fails
- metadata exists but object missing
- object exists but metadata missing

## Checks

1. Verify MinIO health.
2. Check `files` metadata row for tenant.
3. Confirm bucket/key naming.
4. Confirm presigned URL expiry.
5. Check object provider credentials are redacted in logs.

## Recovery

- Recreate presigned URL.
- Reconcile metadata and object store.
- Do not expose raw MinIO secret in logs.
