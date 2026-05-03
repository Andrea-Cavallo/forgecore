# Payments Runbook

## Symptoms

- duplicate charge suspected
- Stripe webhook rejected
- payment stuck in pending

## Checks

1. Check `idempotency_key` for the payment request.
2. Verify provider id and status.
3. Verify Stripe webhook signature.
4. Confirm NATS event publication or outbox state.
5. Inspect `payments_provider_id_idx` and `payments_idempotency_key_idx`.

## Recovery

- Never retry provider charge without the original idempotency key.
- Reconcile payment status from provider before manual state changes.
- Publish missing events through outbox replay when available.
