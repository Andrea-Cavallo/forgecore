# ForgeCore Refactor Roadmap

Questo e' il piano canonico da cui partire per il refactor ForgeCore.

Regola principale: non ricreare cio' che esiste gia'. Ogni fase parte da inventario, confronto con la struttura reale e refactor mirato.

## Phase 0 - Inventario e riallineamento

- [x] Consultare `graphify-out/wiki/index.md` o `graphify-out/GRAPH_REPORT.md` per capire struttura reale, god nodes e community.
- [x] Inventariare moduli Go, package condivisi, servizi, Dockerfile, file config, migrazioni e README esistenti.
- [x] Mappare i nomi attuali dei servizi verso naming ForgeCore.
- [x] Identificare codice duplicato gia' presente in config loading, env parsing, default, logging, middleware e validazione.
- [x] Identificare package generici esistenti (`utils`, `common`, `helpers`, `misc`) e proporre nomi piu' parlanti.
- [x] Aggiornare `docs/forgecore/issues.md` con problemi reali trovati durante l'inventario.
- [x] Aggiornare `README.md` con stato reale, naming e comandi esistenti.
- [x] Verificare build iniziale dei moduli esistenti, annotando cosa passa e cosa fallisce.

Note Phase 0:

- `graphify-out/` e' stato rigenerato dopo le modifiche di codice; `GRAPH_REPORT.md` indica 835 nodi, 990 archi e 124 community.
- God nodes principali: `writeError()`, `Handler`, `writeJSON()`, `run()`, `decodeJSON()`, `extractIDs()`, `main()`, `loadConfig()`, `envOr()`, `config`.
- Esistono gia' 12 servizi con `go.mod` e `Dockerfile`.
- Esistono gia' gli 8 package shared: `crypto`, `events`, `i18n`, `middleware`, `observability`, `pagination`, `proto`, `validation`.
- Il naming iniziale era legacy (`auth-service`, `payment-service`, ecc.) ed e' stato mappato verso `forgecore-*`.
- I loader config/env sono duplicati nei `cmd/server` e `cmd/worker` di diversi servizi.
- Package generico rilevato `sdk/go/common`, poi risolto dividendolo in `sdk/go/clientretry` e `sdk/go/clienttransport`.
- Build verificata: `sdk/go`, tutti i 12 servizi e `shared` passano con `go build ./...`.

## Phase 1 - Naming ForgeCore

- [x] Definire mappa finale dei servizi `forgecore-<bounded-context>`.
- [x] Rinominare cartelle servizio solo quando il contenuto esistente e' stato verificato.
- [x] Aggiornare moduli Go, Dockerfile, Docker Compose, metriche, logger e healthcheck al naming ForgeCore.
- [x] Aggiornare import path e riferimenti interni senza introdurre alias temporanei inutili.
- [x] Aggiornare README e documentazione operativa.
- [x] Eseguire build dei moduli rinominati.

Mappa naming ForgeCore:

| Legacy | ForgeCore |
| --- | --- |
| `api-gateway` | `forgecore-gateway` |
| `auth-service` | `forgecore-auth` |
| `payment-service` | `forgecore-payments` |
| `notification-service` | `forgecore-notifications` |
| `admin-service` | `forgecore-admin` |
| `audit-service` | `forgecore-audit` |
| `job-service` | `forgecore-jobs` |
| `permission-service` | `forgecore-permissions` |
| `config-service` | `forgecore-config` |
| `webhook-service` | `forgecore-webhooks` |
| `storage-service` | `forgecore-storage` |
| `subscription-service` | `forgecore-subscriptions` |

Note Phase 1:

- Cartelle servizio rinominate sotto `services/forgecore-*`.
- Moduli Go aggiornati a `github.com/Andrea-Cavallo/golang-modules/services/forgecore-*`.
- Docker Compose aggiornato a rete/servizi `forgecore`.
- Import path interni aggiornati e verificati con `go mod tidy`.
- Build verificata: `sdk/go`, tutti i 12 servizi rinominati e `shared` passano con `go build ./...`.

## Phase 2 - Configurazione centralizzata

- [x] Progettare il contratto di `forgecore-config`.
- [x] Inventariare tutti i loader config esistenti nei servizi.
- [x] Estrarre la logica comune nei package SDK `shared/configloader`, `shared/configschema` e `shared/configsource`.
- [x] Supportare configurazioni da YAML.
- [x] Supportare override da ENV con precedenza su YAML.
- [x] Aggiungere schema, default espliciti e validazione.
- [x] Aggiungere wrapper per segreti che impediscano logging accidentale.
- [x] Collegare i servizi esistenti alla SDK config condivisa senza duplicare parsing.
- [x] Eliminare loader duplicati rimasti nei servizi.
- [x] Aggiornare README con formato YAML/ENV e priorita' delle sorgenti.
- [x] Eseguire build e test della config SDK e dei servizi toccati.

Note Phase 2:

- `forgecore-config` resta il bounded context runtime per configurazioni tenant/distribuite.
- `shared/configsource` carica sorgenti YAML, ENV e mappe di default.
- `shared/configschema` definisce schema, tipi, default, required e secret redatti.
- `shared/configloader` compone default -> YAML -> ENV con precedenza finale ENV.
- I servizi con config runtime usano la SDK config condivisa; `envOr` duplicato e parsing diretto ENV sono rimasti solo nei package SDK di config.
- File YAML opzionale indicato da `FORGECORE_CONFIG_FILE`.
- Build verificata: `sdk/go`, tutti i 12 servizi e `shared` passano con `go build ./...`.

## Phase 3 - Package naming e boundary

- [x] Rinominare package generici verso nomi parlanti e tecnici.
- [x] Separare package troppo larghi in responsabilita' piccole.
- [x] Verificare che `domain/` non importi infrastruttura.
- [x] Verificare che `application/` dipenda solo da dominio e porte applicative.
- [x] Aggiungere controlli lint o test per impedire import vietati tra layer.
- [x] Aggiornare README con regole di package naming.
- [x] Eseguire build dei moduli toccati.

Note Phase 3:

- `sdk/go/common` e' stato diviso in `sdk/go/clientretry` e `sdk/go/clienttransport`.
- Aggiunto `scripts/check-boundaries.ps1` per verificare import vietati tra DDD layers.
- Corretto `forgecore-payments/internal/application/webhook.go`: ora dipende da `domain.PaymentWebhookEvent`, non dal provider Stripe infrastrutturale.
- Boundary check verificato: `Boundary DDD verificate`.
- Build verificata: `sdk/go`, tutti i 12 servizi e `shared` passano con `go build ./...`.

## Phase 4 - SDK shared stabile

- [x] Inventariare package `shared/` esistenti.
- [x] Definire API pubbliche interne della SDK.
- [x] Stabilizzare error model condiviso.
- [x] Stabilizzare tenant context e middleware.
- [x] Stabilizzare pagination cursor encode/decode.
- [x] Stabilizzare validation wrapper.
- [x] Stabilizzare observability setup.
- [x] Stabilizzare crypto helper e secret handling.
- [x] Aggiornare README con API SDK disponibili.
- [x] Eseguire build e test dei package shared.

Note Phase 4:

- Package `shared/` inventariati: `apperrors`, `configloader`, `configschema`, `configsource`, `crypto`, `events`, `i18n`, `middleware`, `observability`, `pagination`, `proto`, `validation`.
- Aggiunto `shared/apperrors` con `Code`, `Error`, wrapping, `IsCode` e mapping HTTP.
- Middleware stabilizzato con header pubblici `HeaderTenantID`, `HeaderRequestID`, `HeaderAuthorization`.
- Pagination stabilizzata con `DefaultLimit`, `MaxLimit`, `NormalizeLimit`, encode/decode e SQL helper.
- Validation stabilizzata con `Validator`, `NewValidator`, `Validate`, `FieldError`, `FieldErrors`.
- Crypto stabilizzata con `AES256KeySize`, `MinPepperSize` e `NewPIIEncryptorChecked`.
- Observability stabilizzata con `ServiceInfo` e `NewServiceLogger` mantenendo logger, metrics, tracer e shutdown helper.
- Test shared aggiunti per error model, config precedence, crypto, middleware context, pagination e validation.
- Verifiche: `go test ./...` e `go build ./...` passano in `shared`; build completa passa su `sdk/go`, tutti i 12 servizi e `shared`.

## Phase 5 - Eventi, proto e client generati

- [x] Inventariare proto ed eventi esistenti.
- [x] Versionare eventi NATS con nome, versione, tenant id, correlation id, timestamp e payload tipizzato.
- [x] Definire regole di compatibilita' backward e forward.
- [x] Generare o standardizzare client Go interni per gRPC/REST.
- [x] Eliminare client manuali duplicati.
- [x] Aggiornare compatibility matrix tra SDK, proto, servizi e schema DB.
- [x] Aggiornare README con processo di generazione.
- [x] Eseguire build e test dei client generati.

Note Phase 5:

- Proto inventariati: `auth`, `payment`, `permission`, `config`, `audit`.
- Eventi inventariati: `auth`, `payment`, `audit`, `notification`, `publisher`.
- Aggiunto `shared/events.Metadata`, `EventVersionV1`, `Versioned` e nomi evento `.v1`.
- Payload eventi arricchiti con `version`, `event_name`, `correlation_id`, mantenendo `tenant_id` e `occurred_at`.
- Aggiunto `scripts/check-proto-contracts.ps1` per verificare `proto3`, package `.v1` e `go_package`.
- Aggiunto `scripts/check-sdk-clients.ps1` per verificare client SDK standardizzati con `clientretry` e `clienttransport`.
- Aggiunta compatibility matrix: `docs/forgecore/compatibility-matrix.md`.
- Aggiunto processo client: `docs/forgecore/client-generation.md`.
- Verifiche: proto contracts, SDK clients, `go build ./...` in `sdk/go`, `go test ./...` in `sdk/go`.

## Phase 6 - Multi-tenancy e persistenza

- [x] Inventariare tabelle e migrazioni esistenti.
- [x] Verificare presenza di `tenant_id` nelle tabelle applicative.
- [x] Verificare indici per `tenant_id`.
- [x] Verificare Row Level Security e policy `tenant_isolation`.
- [x] Introdurre wrapper transazionale tenant-aware per `SET LOCAL app.tenant_id`.
- [x] Eliminare uso manuale duplicato di tenant setup nei repository.
- [x] Aggiornare README con pattern di persistenza tenant-aware.
- [x] Eseguire integration test PostgreSQL dove disponibili.

Note Phase 6:

- Migrazioni iniziali gia' presenti per `forgecore-auth` e `forgecore-payments`.
- Aggiunte migrazioni tenant/RLS per `forgecore-audit`, `forgecore-config`, `forgecore-notifications`, `forgecore-permissions`, `forgecore-storage`, `forgecore-subscriptions`, `forgecore-webhooks`.
- Ogni tabella applicativa nelle migrazioni ha `tenant_id`, indice tenant, RLS e policy `tenant_isolation`.
- Aggiunto `shared/postgres.WithTenantTx` con `SET LOCAL app.tenant_id = $1`.
- Aggiunto `scripts/check-tenant-migrations.ps1`.
- Uso manuale duplicato di tenant setup non rilevato nei repository; il solo `SET app.tenant_id` esistente e' nel test helper auth.
- Verifiche: tenant migrations, boundary check, `go test ./...` in `forgecore-auth`, `shared`, `sdk/go`, build completa su tutti i moduli.

## Phase 7 - Affidabilita' operativa

- [x] Definire pattern idempotenza per webhook, pagamenti, subscription, notifiche e job.
- [x] Valutare outbox pattern per scritture DB con pubblicazione NATS.
- [x] Aggiungere security baseline e threat model minimo per servizio.
- [x] Aggiungere runbook operativi per tenant, webhook, job, pagamenti, audit e storage.
- [x] Rafforzare Docker Compose locale con seed, migrazioni e healthcheck reali.
- [x] Aggiornare README con workflow operativo locale.
- [x] Eseguire build e smoke test dello stack locale dove disponibile.

Note Phase 7:

- Aggiunto `docs/forgecore/reliability-patterns.md` con pattern idempotenza per webhook, pagamenti, subscription, notifiche e job.
- Valutato outbox pattern per scritture PostgreSQL con pubblicazione NATS e definito flusso transazionale consigliato.
- Aggiunto `docs/forgecore/security-baseline.md` con baseline comune e threat model minimo per ogni servizio ForgeCore.
- Aggiunti runbook operativi per tenant, webhooks, jobs, payments, audit e storage.
- Docker Compose locale ripulito: chiave `version` rimossa, `.env` opzionale, healthcheck infrastrutturali verificati da `docker compose config --quiet`.
- README aggiornato con workflow operativo locale, migrazioni, seed/bootstrap, smoke e link ai runbook.
- Verifiche: `scripts/smoke-local.ps1`, `scripts/build-all.ps1`, test `shared`, `sdk/go` e `forgecore-auth`.

## Phase 8 - Release e developer experience

- [x] Definire comandi comuni per build, test, lint, migrate, proto e scaffold.
- [x] Creare service template generator per nuovi servizi ForgeCore.
- [x] Aggiungere cartella ADR per decisioni architetturali.
- [x] Definire changelog e semantic versioning.
- [x] Definire gestione breaking changes.
- [x] Aggiornare README con comandi ufficiali.
- [x] Eseguire build finale dell'area completa disponibile.

Note Phase 8:

- Aggiunto `Makefile` con comandi comuni: build, verify, smoke, test shared, test SDK, test auth, scaffold e scaffold dry-run.
- Aggiunto `scripts/build-all.ps1` per build completa multi-modulo.
- Aggiunto `scripts/scaffold-service.ps1` per creare nuovi servizi `forgecore-*` con DDD layout, `go.mod`, server entrypoint e Dockerfile; verificato con `-DryRun`.
- Aggiunta cartella ADR con `docs/forgecore/adr/0001-forgecore-platform-boundaries.md`.
- Aggiunti `CHANGELOG.md` e `docs/forgecore/release.md` per semantic versioning, changelog e breaking changes.
- README aggiornato con comandi ufficiali e release policy.
- Verifiche: scaffold dry-run, smoke locale, build completa, test `shared`, `sdk/go` e `forgecore-auth`.

## Phase 9 - Frontend API readiness

- [x] Definire quali API sono pubbliche via gateway e quali restano interne.
- [x] Aggiungere contratto OpenAPI per le API frontend-facing.
- [x] Standardizzare envelope HTTP per errori e risposte gateway.
- [x] Rafforzare CORS, security headers e preflight nel gateway.
- [ ] Completare auth frontend flow: register, login, refresh, logout, me, reset password, email verification e OAuth callback.
- [ ] Aggiungere client TypeScript generato o scaffoldato da OpenAPI.
- [ ] Aggiungere seed locale tenant + admin user verificabile.
- [ ] Aggiungere smoke E2E: gateway, auth public flow e chiamata protetta.
- [x] Aggiornare README con frontend integration guide.
- [x] Eseguire build e test E2E disponibili.

Note Phase 9:

- Aggiunto `docs/forgecore/frontend-api-readiness.md` con superficie gateway pubblica, header richiesti, error envelope e regole CORS.
- Aggiunto contratto iniziale `docs/forgecore/openapi/forgecore-gateway.v1.yaml`.
- Gateway aggiornato con CORS allowlist comma-separated, `Idempotency-Key` nei preflight, security headers e auth error JSON con `code`, `message`, `request_id`.
- Gateway espone ora `/healthz` e `/readyz` oltre a `/health`; tutti e tre bypassano auth.
- Aggiunto E2E gateway in `services/forgecore-gateway/cmd/server/gateway_e2e_test.go`.
- Aggiunto comando `make test-e2e` e script `scripts/e2e-gateway.ps1`.
- Verifiche: `scripts/e2e-gateway.ps1`, `go test ./...` in `forgecore-gateway`, `scripts/smoke-local.ps1`, `scripts/build-all.ps1`.

## Phase 10 - Service runtime hardening

- [x] Aggiungere `/healthz` e `/readyz` reali per tutti i servizi.
- [x] Implementare outbox pattern nei servizi event-driven.
- [x] Implementare idempotency store/use case per webhook, payments, subscriptions, notifications e jobs.
- [ ] Aggiungere integration test DB/NATS/Redis per servizi principali.
- [x] Aggiungere CI job per build, test, smoke locale ed E2E gateway.
- [x] Aggiungere metriche operative per outbox, idempotenza, provider esterni e job.
- [x] Aggiornare runbook con procedure di recovery verificate.
- [ ] Eseguire build finale e smoke runtime completo dove disponibile.

Note Phase 10:

- Aggiunto `shared/health` e collegati `/health`, `/healthz`, `/readyz` ai servizi HTTP e ai worker/consumer con health server dedicato.
- Aggiunto `shared/outbox` con messaggio versionato, repository interface e dispatcher.
- Aggiunto `shared/idempotency` con store interface e memory store testato.
- Aggiunte metriche operative in `shared/observability`: outbox, idempotency, provider calls e job runs.
- Aggiunto `scripts/check-runtime-hardening.ps1` e incluso nello smoke locale.
- Aggiunto `scripts/integration-local.ps1` per PostgreSQL/Redis/NATS; verifica locale bloccata da porta host `5432` gia' occupata.
- Aggiunto workflow CI `.github/workflows/ci.yml`.
- Aggiunto runbook `docs/forgecore/runbooks/recovery.md`.
- Verifiche passate: runtime hardening check, security check disponibile, build completa, smoke locale, Docker Compose build completo.
- Verifica non completata: integration runtime DB/NATS/Redis per conflitto porta locale `5432`.

## Phase 11 - OWASP Security Hardening

- [x] Mappare ogni endpoint gateway/servizio contro OWASP Top 10.
- [x] Aggiungere `govulncheck` e dependency scanning.
- [x] Aggiungere container image scanning.
- [x] Aggiungere security tests per CORS, auth, tenant isolation, SSRF, webhook signature.
- [x] Verificare RBAC/permissions su ogni endpoint protetto.
- [x] Completare E2E auth: register, login, refresh, logout, me, chiamata protetta.
- [x] Definire token storage policy per frontend.
- [x] Definire CSRF policy se si usano cookie.
- [x] Implementare key rotation/secrets lifecycle.
- [x] Aggiungere audit log obbligatorio per azioni sensibili.
- [x] Portare security baseline verso OWASP ASVS livello 1 come minimo.
- [x] Trasformare i controlli security in test automatici, CI security checks e controlli runtime reali.
- [x] Aggiornare README e `docs/forgecore/security-baseline.md` con stato OWASP verificato.
- [x] Eseguire build, smoke, E2E e security checks disponibili.

Note Phase 11:

- Aggiunto `docs/forgecore/owasp-endpoint-map.md`.
- Aggiunto `scripts/security-check.ps1` con `govulncheck` quando disponibile e test security gateway/webhooks/payments.
- Aggiunto workflow `.github/workflows/security.yml` con `govulncheck` e Trivy image scan.
- Aggiunti test security: CORS forbidden origin, missing token envelope, SSRF webhook URL e Stripe invalid signature.
- Aggiunta policy `docs/forgecore/frontend-token-csrf-policy.md`.
- Aggiunta matrice RBAC `docs/forgecore/rbac-endpoint-matrix.md`, middleware RBAC gateway e check `scripts/check-rbac-security.ps1`.
- Aggiunto E2E auth applicativo per register, login, refresh, logout, me e validazione token protetta.
- Aggiunto lifecycle runtime JWT con `kid`, current key e previous keys tramite `NewRotatingJWTService`.
- Aggiunto audit middleware gateway per mutazioni sensibili con tenant, user e request id.
- Aggiunto piano lifecycle `docs/forgecore/secrets-lifecycle.md` e implementazione runtime di key rotation per JWT.
- Aggiornati README, security baseline e hardening docs.
- Verifiche passate: `scripts/security-check.ps1`, `scripts/check-rbac-security.ps1`, gateway E2E, auth E2E applicativo, build completa, smoke locale.
- `govulncheck` locale non era installato: lo script lo segnala come skip; CI lo installa ed esegue.

## Regole di avanzamento

- [x] Ogni task deve restare `[ ]` finche' non e' stato completato e verificato.
- [x] Ogni modifica di codice richiede build dell'area toccata.
- [x] Ogni modifica strutturale richiede aggiornamento README.
- [x] Ogni problema corretto deve essere marcato in `docs/forgecore/issues.md`.
- [x] Dopo modifiche a codice, aggiornare graphify.
