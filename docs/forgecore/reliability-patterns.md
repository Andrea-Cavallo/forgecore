# ForgeCore Reliability Patterns

## Idempotency

Every operation that can be retried must accept or derive an idempotency key.

Required areas:

- webhooks: use provider event id or delivery id
- payments: use `idempotency_key`
- subscriptions: use provider subscription id plus tenant id
- notifications: use template, recipient, tenant, and triggering event id
- jobs: use job type plus tenant and payload hash

Pattern:

```text
tenant_id + operation + idempotency_key
```

Rules:

- Store the idempotency key before or in the same transaction as the side effect.
- Return the previously persisted result for duplicate keys.
- Never call external providers twice for the same committed idempotency key.
- Treat webhook delivery retries as expected behavior.

## Outbox

When a database write must publish a NATS event, prefer an outbox table in the same transaction.

Recommended columns:

- `id`
- `tenant_id`
- `subject`
- `event_name`
- `version`
- `payload`
- `status`
- `attempts`
- `next_retry_at`
- `created_at`
- `published_at`

Flow:

1. Write domain state and outbox row in one PostgreSQL transaction.
2. A worker publishes pending rows to NATS.
3. Mark rows as published only after NATS publish succeeds.
4. Retry with backoff and a max attempts policy.

This prevents "DB committed but event lost" failures.
