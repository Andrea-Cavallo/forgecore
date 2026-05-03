# Webhooks Runbook

## Symptoms

- provider retries webhook repeatedly
- delivery status remains failed
- endpoint receives duplicate payloads

## Checks

1. Verify webhook signature.
2. Check endpoint active flag and event subscription.
3. Inspect `webhook_deliveries` attempts and `next_retry_at`.
4. Confirm idempotency key is provider event id or delivery id.
5. Check private IP and HTTPS validation.

## Recovery

- Do not manually replay without preserving idempotency key.
- Fix endpoint URL and reactivate endpoint.
- Requeue failed deliveries with bounded attempts.
