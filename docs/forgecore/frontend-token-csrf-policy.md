# Frontend Token And CSRF Policy

ForgeCore production direction:

- access token: short-lived JWT kept in frontend memory
- refresh token: `HttpOnly`, `Secure`, `SameSite=Lax` cookie
- state-changing cookie-authenticated requests: explicit CSRF token
- logout: revoke refresh token/session server-side
- refresh: rotate refresh token on every use

Rules:

- Never store refresh tokens in `localStorage`.
- Do not log access tokens, refresh tokens or authorization headers.
- Bearer-only local development is allowed while E2E auth is being completed.
- If a frontend uses cookies for authenticated requests, CSRF protection is mandatory.

CSRF token shape:

```text
X-CSRF-Token: <opaque-random-token>
```

The token must be bound to the user session and rotated after login/logout.
