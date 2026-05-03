# Secrets Lifecycle

ForgeCore secrets must be managed as short-lived or rotatable runtime configuration.

## Secret Classes

| Class | Examples | Rotation |
| --- | --- | --- |
| JWT signing keys | RS256 private/public keypair | staged key rotation with `kid` |
| PII crypto keys | AES-256 key, HMAC pepper | dual-read migration window |
| Provider secrets | Stripe, SendGrid, Twilio, MinIO | provider-side rotate and deploy |
| Webhook secrets | endpoint HMAC secret | per-endpoint rotate |

## Required Controls

- Secrets use `configschema.Field{Secret: true}`.
- Secrets must not be logged.
- Runtime string output must redact secret values.
- Production secrets must come from a secret manager or mounted secret, not committed files.
- Key rotation must support an overlap period where old and new keys are both accepted where needed.

## JWT Key Rotation Target

- Tokens include a `kid`.
- New tokens are signed with the current key.
- Old keys remain valid until all old access tokens expire.
- `NewRotatingJWTService(currentKID, currentSecret, previous)` signs with the current key and validates previous `kid` values during the overlap window.
