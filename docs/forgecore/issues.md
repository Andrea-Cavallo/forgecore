# Issues — Go Code Review (2026-04-03)

Generato da `/go-review project`. Da risolvere prima di ogni merge/deploy.

Ultimo aggiornamento: 2026-04-08 — sessione completata. Tutti i CRITICAL/HIGH/MEDIUM risolti.

## Nuovi problemi trovati in questa sessione

- **N-1** — `forgecore-auth/internal/application/register.go` e `login.go`: import mancante `shared/crypto` → **RISOLTO** (aggiunto import).
- **N-2** — `forgecore-auth/internal/application`: `*events.Publisher` non è un'interfaccia, rende impossibile il mock nei test → **RISOLTO** (creato `EventPublisher` interface + `noopPublisher` in `interfaces.go`).
- **N-3** — `forgecore-auth/internal/application/jwt.go`: file mancante, `TokenIssuer` non aveva implementazione → **RISOLTO** (creato `JWTService` con `Issue` e `Validate`).
- **N-4** — `forgecore-auth/migrations/`: directory inesistente → **RISOLTO** (create SQL up/down per users e sessions con RLS).

---

## CRITICAL (blocca compilazione o introduce vulnerabilità)

- ~~**C-1**~~ ✅ — `go.mod` × 10 servizi: `replace github.com/yourorg/golang-modules/shared => ../../shared` aggiunto in tutti i go.mod.
- ~~**C-2**~~ ✅ — `shared/validation/validator.go:36`: ora usa `errors.As` correttamente.
- ~~**C-3**~~ ✅ — `forgecore-webhooks/internal/application/register_endpoint.go`: validazione HTTPS + blocco IP privati/loopback implementata.
- ~~**C-4**~~ ✅ — `shared/crypto/pii.go`: usa HMAC-SHA256 con pepper (`NewPIIEncryptor(key, pepper []byte)`).
- ~~**C-5**~~ ✅ — `shared/observability/shutdown.go`: rimosso `os.Exit`, ritorna normalmente.

---

## HIGH (bug silenzioso o violazione architetturale)

- ~~**H-1**~~ ✅ — `forgecore-auth/internal/application/login.go`: errore sessione gestito correttamente.
- ~~**H-2**~~ ✅ — `forgecore-payments/internal/application/create_payment.go`: record di fallimento salvato via `handleChargeFailure`.
- ~~**H-3**~~ ✅ — `forgecore-webhooks/internal/application/deliver.go`: record consegna salvato, errore propagato.
- ~~**H-4**~~ ✅ — `forgecore-auth/internal/application/register.go`: usa `errors.Is(err, domain.ErrUserNotFound)`.
- ~~**H-5**~~ ✅ — `forgecore-subscriptions/internal/application/subscribe.go`: usa `errors.Is(err, domain.ErrSubscriptionNotFound)`.
- ~~**H-6**~~ ✅ — `shared/observability/metrics.go`: usa `mustRegisterOrExisting` invece di `MustRegister`.
- ~~**H-7**~~ ✅ — `forgecore-gateway/internal/middleware/ratelimit.go`: usa `net.SplitHostPort` + eviction periodica.
- ~~**H-8**~~ ✅ — `forgecore-payments/internal/application/create_payment.go`: refactored in `buildPayment` + `handleChargeFailure` (≤50 righe ciascuna).
- ~~**H-9**~~ — `domain/repository.go`: il layer domain importa `shared/pagination`. Accettato come compromesso pragmatico: `pagination.Cursor` è un value object puro senza dipendenze esterne. Da rivedere se si introduce un bounded context separato.

---

## MEDIUM

- ~~**M-1**~~ ✅ — `shared/observability/shutdown.go`: registra solo un signal handler.
- ~~**M-2**~~ ✅ — `shared/middleware/tenant.go`: risposta 400 usa messaggi statici, non echeggia l'header.
- ~~**M-3**~~ ✅ — `shared/pagination/cursor.go`: errore di `json.Marshal` propagato correttamente.
- ~~**M-4**~~ ✅ — `forgecore-webhooks/internal/application/deliver.go`: `mac.Write` ignorato documentato (hash.Hash.Write non restituisce mai errore per spec).
- ~~**M-5**~~ ✅ — `forgecore-gateway/internal/router/router.go`: `w.Write` ignorato documentato (non-actionable post-header).
- ~~**M-6**~~ ✅ — `shared/observability/tracer.go`: `insecure bool` è ora parametro di `InitTracer`.
- ~~**M-7**~~ — `shared/validation/validator.go`: singleton package-level. Accettato: `validator.Validate` è thread-safe per design della libreria.
- ~~**M-8**~~ ✅ — `forgecore-notifications/internal/application/send.go`: `maxRetryAttempts` rimosso.
- ~~**M-9**~~ ✅ — `shared/i18n/locale.go`: locale IT usa virgola decimale e formato data DD/MM/YYYY.
- **M-10** — Intero repo: **zero file di test** — copertura 0% vs. requisito 80%. Da affrontare in Fase 1+.

---

## LOW

- ~~**L-1**~~ ✅ — `forgecore-jobs/internal/jobs/registry.go`: `MustMarshal` rimosso, ora usa `testing_helpers_test.go`.
- ~~**L-2**~~ ✅ — `forgecore-auth/internal/application/login.go`: magic number → costante `sessionDuration`.
- ~~**L-3**~~ ✅ — `forgecore-payments/internal/application/create_payment.go`: magic string → `domain.ProviderStripe`.
- ~~**L-4**~~ — `forgecore-config/internal/application/get_config.go`: `cacheTTLSeconds` non condiviso. Accettato: costante locale coerente con il servizio.
- ~~**L-5**~~ ✅ — `shared/validation/validator.go`: messaggi di validazione ora in italiano.

---

## Problemi aperti

| ID | Priorità | Stato |
|----|----------|-------|
| H-9 | HIGH | Accettato (compromesso pragmatico) |
| M-7 | MEDIUM | Accettato (thread-safe by design) |
| M-10 | MEDIUM | Parzialmente risolto: test scritti per forgecore-auth (domain, application, mfa, integration). Da fare: test per gli altri 11 servizi. |
| L-4 | LOW | Accettato |
| N-5 | MEDIUM | ✅ RISOLTO: `register_test.go`/`login_test.go` cambiati in `package application` (white-box). |
| N-6 | LOW | ✅ RISOLTO: `go mod tidy` eseguito; `GetByOAuthProvider` aggiunto agli stub `stubUserRepo` (register_test.go) e `mfaUserRepo` (mfa_test.go); tutti i test passano. |

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
- **N-6**: AGGIORNATO — go.mod ha ora anche `google.golang.org/grpc v1.70.0`. Eseguire `go mod tidy` in `services/forgecore-auth` per risolvere testcontainers-go + pquerna/otp + grpc.
- **N-7**: ✅ FALSO ALLARME — `oauth.go` usa `[]byte(emailEnc)` identico a `register.go`. Nessuna azione richiesta.

## Aggiornamento 2026-04-13

### Completati in questa sessione:
- **Refactor globale**: tutti i `go.mod` aggiornati a `go 1.26` e `github.com/yourorg` → `github.com/Andrea-Cavallo` (14 go.mod + 85 file .go/.proto/.yaml).
- **Task 3.1–3.5** (API Gateway): marcati ✅ — codice già presente da sessione precedente.
- **Task 4.1–4.3** (Payment Service): marcati ✅ — codice già presente da sessione precedente.
- **Task 7.2**: `SeedRolesUseCase` — crea owner/admin/billing-manager/read-only/user per tenant; idempotente.
- **Task 7.3**: SDK Go client — `sdk/go/auth/client.go`, `sdk/go/permission/client.go`, `sdk/go/clientretry/`, `sdk/go/clienttransport/` con retry 3x, gobreaker, OTEL transport.
- **Task 10.1**: `prometheus.yml` (scraping 11 servizi + postgres/redis/nats), Grafana datasource + dashboard JSON.
- **Task 10.2**: `rules.yml` (ServiceDown, HighErrorRate, HighLatencyP99, JobFailed), `alertmanager.yml` (PagerDuty/Slack/email, inhibit rules).
- **Task 10.3**: `otel-collector.yml` (OTLP HTTP+gRPC, batch, Jaeger exporter, Prometheus exporter).
- **Task 10.4**: `.github/workflows/ci.yml` — go 1.26, lint, unit/integration test con service containers, docker build.
- **Task 10.5**: `.github/workflows/deploy.yml` — push GHCR + SSH rolling deploy con health check.
- **Task 10.6**: `scripts/bootstrap.sh` — vault → nats → migrate → seed super-admin.

### Problemi aggiornati:
- **N-6**: APERTO — `go mod tidy` non ancora eseguito in `services/forgecore-auth`. Eseguire prima del prossimo build.

## Prossima sessione: priorità

1. ~~Eseguire `go mod tidy` in `services/forgecore-auth` (N-6)~~ ✅ RISOLTO
2. Aggiungere test per oauth.go, gdpr.go, grpc/server.go (M-10 parziale)
3. Verificare che `sdk/go` compili dopo `go mod tidy` (aggiungere go.sum)
4. Aggiornare CLAUDE.md: versione Go da 1.24 a 1.26
5. Eventuale Docker Compose update per forgecore-jobs (porta mancante)

## Aggiornamento 2026-05-03

### Completati in questa sessione:
- **Inventario ForgeCore Phase 0**: confermati 12 servizi esistenti con `go.mod` e `Dockerfile`, package `shared/` gia' presenti e `forgecore-config` gia' implementato come base da rifattorizzare.
- **Build baseline**: `go build ./...` passa in `sdk/go`, tutti i 12 servizi e `shared`.
- **Shared build fix**: `shared/go.sum` aggiornato con `go mod tidy`; `shared/validation/validator.go` corretto rimuovendo l'opzione non disponibile `validator.WithRequiredStructFields`.
- **Go module tidy**: eseguito in `forgecore-audit`, `forgecore-auth`, `forgecore-notifications`, `forgecore-payments`, `forgecore-storage`, `forgecore-subscriptions`, `forgecore-webhooks`.

### Problemi aperti trovati:
- **N-8** - RISOLTO: `graphify-out/` rigenerato con `_rebuild_code`; `GRAPH_REPORT.md` consultato dopo il rebuild.
- **N-9** - RISOLTO: cartelle servizio, moduli Go, import path e Docker Compose migrati a `forgecore-*`; build completa verificata.
- **N-10** - RISOLTO: config loading consolidato in `shared/configsource`, `shared/configschema`, `shared/configloader`; servizi collegati a YAML/ENV/default condivisi.
- **N-11** - RISOLTO: `sdk/go/common` diviso in `sdk/go/clientretry` e `sdk/go/clienttransport`.
- **N-12** - RISOLTO: boundary violation in `forgecore-payments/internal/application/webhook.go`; application non importa piu' provider Stripe infrastrutturale.
- **N-13** - RISOLTO: Phase 4 SDK shared stabilizzata con `shared/apperrors`, API pubbliche documentate e test mirati su config, crypto, middleware, pagination e validation.
- **N-14** - RISOLTO: eventi NATS versionati con metadati comuni e compatibility matrix ForgeCore.
- **N-15** - RISOLTO: aggiunti controlli `check-proto-contracts.ps1`, `check-sdk-clients.ps1` e documentazione client.
- **N-16** - RISOLTO: migrazioni tenant/RLS aggiunte per audit, config, notifications, permissions, storage, subscriptions e webhooks.
- **N-17** - RISOLTO: aggiunto `shared/postgres.WithTenantTx` per transazioni PostgreSQL tenant-aware.
- **N-18** - RISOLTO: affidabilita' operativa documentata con pattern idempotenza, valutazione outbox, security baseline e runbook tenant/webhook/job/pagamenti/audit/storage.
- **N-19** - RISOLTO: workflow locale rafforzato con Compose senza `version` obsoleto, `.env` opzionale, smoke check, script build completo e README operativo aggiornato.
- **N-20** - RISOLTO: developer experience completata con Makefile, generator `scripts/scaffold-service.ps1`, ADR, changelog, semantic versioning e gestione breaking changes.
- **N-21** - RISOLTO: Phase 9 avviata con superficie gateway frontend-facing, OpenAPI iniziale, CORS allowlist, auth error envelope, `/healthz`, `/readyz` e test E2E gateway.
- **N-22** - APERTO: auth frontend flow completo non ancora verificato end-to-end con stack reale, seed tenant/admin e chiamata protetta.
- **N-23** - APERTO: client TypeScript da OpenAPI non ancora generato.
- **N-24** - RISOLTO: README e `.env.example` aggiornati con guida pratica per frontend, setup database, configurazioni ENV, gateway, servizi, migrazioni e bootstrap locale.
- **N-25** - APERTO: Phase 11 OWASP Security Hardening aggiunta; ForgeCore e' security-aware ma non ancora security-verified con test automatici, CI security checks e controlli runtime reali.
- **N-26** - PARZIALMENTE RISOLTO: Phase 10 runtime hardening implementata per health/readiness, outbox SDK, idempotency SDK, metriche operative, CI e recovery runbook; integration runtime DB/NATS/Redis bloccata localmente da porta `5432` gia' occupata.
- **N-27** - RISOLTO: Phase 11 implementata per endpoint map, security tests, token/CSRF policy, CI govulncheck/Trivy, security script, RBAC endpoint-by-endpoint, auth E2E completo, runtime key rotation JWT e audit obbligatorio.
- **N-28** - RISOLTO: Dockerfile dei servizi usavano path legacy (`auth-service`, `api-gateway`, ecc.) e Go 1.24; aggiornati a `services/forgecore-*`, Go 1.26 e aggiunto `check-dockerfiles.ps1`.
- **N-29** - RISOLTO: smoke locale non propagava errori dei check figli, log runtime troppo generici e revoke permissions era un endpoint senza use case reale. Corretto smoke, normalizzati messaggi log principali/health con servizio e addr, aggiunto `RevokePermissionUseCase` e verificati test/build/Docker Compose build.
- **N-30** - RISOLTO: Phase 11 completata con RBAC gateway endpoint-by-endpoint, audit middleware per azioni sensibili, JWT key rotation runtime, auth E2E applicativo completo e check automatico `check-rbac-security.ps1`.
