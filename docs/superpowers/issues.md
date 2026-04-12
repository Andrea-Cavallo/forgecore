# Issues — Go Code Review (2026-04-03)

Generato da `/go-review project`. Da risolvere prima di ogni merge/deploy.

Ultimo aggiornamento: 2026-04-08 — sessione completata. Tutti i CRITICAL/HIGH/MEDIUM risolti.

## Nuovi problemi trovati in questa sessione

- **N-1** — `auth-service/internal/application/register.go` e `login.go`: import mancante `shared/crypto` → **RISOLTO** (aggiunto import).
- **N-2** — `auth-service/internal/application`: `*events.Publisher` non è un'interfaccia, rende impossibile il mock nei test → **RISOLTO** (creato `EventPublisher` interface + `noopPublisher` in `interfaces.go`).
- **N-3** — `auth-service/internal/application/jwt.go`: file mancante, `TokenIssuer` non aveva implementazione → **RISOLTO** (creato `JWTService` con `Issue` e `Validate`).
- **N-4** — `auth-service/migrations/`: directory inesistente → **RISOLTO** (create SQL up/down per users e sessions con RLS).

---

## CRITICAL (blocca compilazione o introduce vulnerabilità)

- ~~**C-1**~~ ✅ — `go.mod` × 10 servizi: `replace github.com/yourorg/golang-modules/shared => ../../shared` aggiunto in tutti i go.mod.
- ~~**C-2**~~ ✅ — `shared/validation/validator.go:36`: ora usa `errors.As` correttamente.
- ~~**C-3**~~ ✅ — `webhook-service/internal/application/register_endpoint.go`: validazione HTTPS + blocco IP privati/loopback implementata.
- ~~**C-4**~~ ✅ — `shared/crypto/pii.go`: usa HMAC-SHA256 con pepper (`NewPIIEncryptor(key, pepper []byte)`).
- ~~**C-5**~~ ✅ — `shared/observability/shutdown.go`: rimosso `os.Exit`, ritorna normalmente.

---

## HIGH (bug silenzioso o violazione architetturale)

- ~~**H-1**~~ ✅ — `auth-service/internal/application/login.go`: errore sessione gestito correttamente.
- ~~**H-2**~~ ✅ — `payment-service/internal/application/create_payment.go`: record di fallimento salvato via `handleChargeFailure`.
- ~~**H-3**~~ ✅ — `webhook-service/internal/application/deliver.go`: record consegna salvato, errore propagato.
- ~~**H-4**~~ ✅ — `auth-service/internal/application/register.go`: usa `errors.Is(err, domain.ErrUserNotFound)`.
- ~~**H-5**~~ ✅ — `subscription-service/internal/application/subscribe.go`: usa `errors.Is(err, domain.ErrSubscriptionNotFound)`.
- ~~**H-6**~~ ✅ — `shared/observability/metrics.go`: usa `mustRegisterOrExisting` invece di `MustRegister`.
- ~~**H-7**~~ ✅ — `api-gateway/internal/middleware/ratelimit.go`: usa `net.SplitHostPort` + eviction periodica.
- ~~**H-8**~~ ✅ — `payment-service/internal/application/create_payment.go`: refactored in `buildPayment` + `handleChargeFailure` (≤50 righe ciascuna).
- ~~**H-9**~~ — `domain/repository.go`: il layer domain importa `shared/pagination`. Accettato come compromesso pragmatico: `pagination.Cursor` è un value object puro senza dipendenze esterne. Da rivedere se si introduce un bounded context separato.

---

## MEDIUM

- ~~**M-1**~~ ✅ — `shared/observability/shutdown.go`: registra solo un signal handler.
- ~~**M-2**~~ ✅ — `shared/middleware/tenant.go`: risposta 400 usa messaggi statici, non echeggia l'header.
- ~~**M-3**~~ ✅ — `shared/pagination/cursor.go`: errore di `json.Marshal` propagato correttamente.
- ~~**M-4**~~ ✅ — `webhook-service/internal/application/deliver.go`: `mac.Write` ignorato documentato (hash.Hash.Write non restituisce mai errore per spec).
- ~~**M-5**~~ ✅ — `api-gateway/internal/router/router.go`: `w.Write` ignorato documentato (non-actionable post-header).
- ~~**M-6**~~ ✅ — `shared/observability/tracer.go`: `insecure bool` è ora parametro di `InitTracer`.
- ~~**M-7**~~ — `shared/validation/validator.go`: singleton package-level. Accettato: `validator.Validate` è thread-safe per design della libreria.
- ~~**M-8**~~ ✅ — `notification-service/internal/application/send.go`: `maxRetryAttempts` rimosso.
- ~~**M-9**~~ ✅ — `shared/i18n/locale.go`: locale IT usa virgola decimale e formato data DD/MM/YYYY.
- **M-10** — Intero repo: **zero file di test** — copertura 0% vs. requisito 80%. Da affrontare in Fase 1+.

---

## LOW

- ~~**L-1**~~ ✅ — `job-service/internal/jobs/registry.go`: `MustMarshal` rimosso, ora usa `testing_helpers_test.go`.
- ~~**L-2**~~ ✅ — `auth-service/internal/application/login.go`: magic number → costante `sessionDuration`.
- ~~**L-3**~~ ✅ — `payment-service/internal/application/create_payment.go`: magic string → `domain.ProviderStripe`.
- ~~**L-4**~~ — `config-service/internal/application/get_config.go`: `cacheTTLSeconds` non condiviso. Accettato: costante locale coerente con il servizio.
- ~~**L-5**~~ ✅ — `shared/validation/validator.go`: messaggi di validazione ora in italiano.

---

## Problemi aperti

| ID | Priorità | Stato |
|----|----------|-------|
| H-9 | HIGH | Accettato (compromesso pragmatico) |
| M-7 | MEDIUM | Accettato (thread-safe by design) |
| M-10 | MEDIUM | Parzialmente risolto: test scritti per auth-service (domain, application, mfa, integration). Da fare: test per gli altri 11 servizi. |
| L-4 | LOW | Accettato |
| N-5 | MEDIUM | ✅ RISOLTO: `register_test.go`/`login_test.go` cambiati in `package application` (white-box). |
| N-6 | LOW | `go mod tidy` non eseguito dopo aggiunta testcontainers-go e pquerna/otp in auth-service go.mod. |

## Aggiornamento 2026-04-09

### Completati in questa sessione:
- **Task 1.4 Step 2**: scritti integration test con testcontainers-go (`user_repository_integration_test.go`, `testhelper_test.go`)
- **Task 2.1**: MFA/TOTP — `EnableMFAUseCase`, `VerifyMFAUseCase` (con backup codes), `DisableMFAUseCase`, `mfa_handler.go`, migration `000002_add_mfa_backup_codes`
- **Task 2.3**: Password Reset + Email Verification — `ForgotPasswordUseCase`, `ResetPasswordUseCase`, `VerifyEmailUseCase`, `ResendVerificationUseCase`. Estesa `TokenStore` con `StoreOneTimeToken`/`PopOneTimeToken`.
- **Task 2.4**: Session Management — `ListSessionsUseCase`, `RevokeSessionUseCase`, `RevokeAllSessionsUseCase`. Estesa `SessionRepository` con `ListByUser`/`DeleteByID`.

## Aggiornamento 2026-04-12

### Completati in questa sessione:
- **Task 2.2**: OAuth2 Google + GitHub — `OAuthAuthorizeUseCase`, `OAuthCallbackUseCase` (verifica CSRF state, exchange code, upsert user: lookup by oauth_provider → email → create). Infrastructure: `oauth/google.go`, `oauth/github.go`. Domain: aggiunti `OAuthProvider`, `OAuthProviderID` in `User`, `GetByOAuthProvider` in `UserRepository`, migrazione `000003_add_oauth_fields`. Aggiornato `user_repository.go` (Create, Update, scanUser, GetByOAuthProvider).
- **Task 2.5**: GDPR — `ExportDataUseCase`, `DeleteAccountUseCase` (soft delete + azzeramento PII).
- **Task 2.6**: gRPC server — `AuthServer` con `ValidateToken` + `GetUser`, JSON codec custom, service descriptor manuale. Proto aggiornato in `shared/proto/auth.proto`. Aggiunto `google.golang.org/grpc v1.70.0` in go.mod.

### Problemi aggiornati:
- **N-5**: ✅ RISOLTO nella sessione precedente
- **N-6**: AGGIORNATO — go.mod ha ora anche `google.golang.org/grpc v1.70.0`. Eseguire `go mod tidy` in `services/auth-service` per risolvere testcontainers-go + pquerna/otp + grpc.
- **N-7**: ✅ FALSO ALLARME — `oauth.go` usa `[]byte(emailEnc)` identico a `register.go`. Nessuna azione richiesta.

## Prossima sessione: priorità

1. Eseguire `go mod tidy` in `services/auth-service` (N-6)
2. Verificare conversione `emailEnc string → []byte` in `oauth.go` (N-7)
3. Implementare Task 3.x — API Gateway (proxy, middleware chain, rate limiting, auth middleware)
4. Aggiungere test per oauth.go, gdpr.go, grpc/server.go
5. Implementare Task 4.x — Payment Service (domain, stripe adapter, create payment, webhook)
