# ForgeCore

![ForgeCore](./forgecore.png)

ForgeCore e' una base backend in Go per costruire piattaforme multi-tenant con microservizi, SDK condivisa e regole operative gia' pronte. Non e' un singolo servizio: e' un monorepo che raccoglie primitive comuni, servizi di riferimento e script di verifica per partire piu' velocemente senza ricostruire ogni volta autenticazione, configurazione, tenant isolation, eventi, osservabilita', paginazione, storage, pagamenti e permessi.

L'obiettivo e' dare a chi sviluppa frontend, backend o prodotti SaaS un backend coerente da usare come fondazione. Il frontend parla con un solo ingresso pubblico, `forgecore-gateway`; i servizi interni restano dietro al gateway e seguono confini DDD chiari.

## Perche' Esiste

Ogni piattaforma seria finisce per riscrivere gli stessi pezzi:

- login, sessioni, JWT, refresh token e ruoli
- configurazioni da YAML ed ENV
- isolamento tenant con PostgreSQL Row Level Security
- errori HTTP coerenti e request id
- validazione input
- eventi NATS tipizzati e versionati
- paginazione a cursore
- cifratura PII e gestione segreti
- metriche Prometheus, log JSON e tracing
- Docker Compose, migrazioni, seed e smoke test
- controlli automatici su boundary, proto, SDK, Dockerfile e sicurezza

ForgeCore mette questi pezzi in una SDK condivisa e in servizi pronti da estendere. Serve a ridurre codice duplicato, mantenere naming coerente e rendere piu' semplice aggiungere nuovi bounded context senza rompere le regole della piattaforma.

## A Chi Serve

ForgeCore e' utile se stai costruendo:

- un SaaS multi-tenant
- una dashboard amministrativa
- un portale clienti con abbonamenti e pagamenti
- un marketplace con vendor e backoffice
- un'app con storage, webhook, audit e notifiche
- una piattaforma interna con RBAC e configurazioni tenant-specific

Se vuoi solo una piccola API monolitica, ForgeCore puo' essere piu' grande del necessario. Se invece vuoi una base production-oriented per servizi separati e frontend moderni, il progetto nasce proprio per quello.

## Stato Frontend

ForgeCore e' gia' utilizzabile per prototipi frontend e sviluppo interno tramite `forgecore-gateway`.

Gia' presente:

- gateway pubblico su `http://localhost:8080`
- CORS allowlist con `CORS_ORIGIN`
- security headers e preflight prima dell'autenticazione
- error envelope per errori auth del gateway
- auth flow applicativo coperto da test per register, login, refresh, logout, me e token validation
- route protette con JWT e RBAC lato gateway
- OpenAPI iniziale in `docs/forgecore/openapi/forgecore-gateway.v1.yaml`
- policy token/CSRF in `docs/forgecore/frontend-token-csrf-policy.md`
- matrice RBAC in `docs/forgecore/rbac-endpoint-matrix.md`
- collection Postman in `docs/forgecore/postman/ForgeCore.postman_collection.json`
- environment Postman in `docs/forgecore/postman/ForgeCore.local.postman_environment.json`

Manca ancora per dichiarare la parte frontend production-ready:

- client TypeScript generato o scaffoldato da OpenAPI
- OpenAPI completa con request e response schema per tutti i servizi
- seed locale unico per tenant, admin, ruoli e piani demo
- verifica E2E runtime completa con Compose, PostgreSQL, Redis, NATS, MinIO e migrazioni reali
- harness frontend con Playwright o Cypress
- pacchetto SDK frontend versionato e compatibile con la versione OpenAPI

Il piano di verifica completo vive in `docs/forgecore/frontend-testplan.md`.

## Cosa Puoi Costruire

Esempi concreti di uso:

1. Dashboard SaaS multi-tenant: login, ruoli, configurazioni tenant, audit log e schermate admin protette.
2. Portale billing: registrazione utente, verifica email, scelta piano, abbonamento, pagamento e retry idempotente.
3. Backoffice marketplace: gestione vendor, upload file, notifiche e webhook di integrazione.
4. Pannello impostazioni embedded: OAuth, feature flag tenant-specific e CORS su origini approvate.
5. Console customer support: ricerca utenti, audit trail, disabilitazione account e permessi granulari.
6. File manager aziendale: storage metadata, presigned URL, paginazione e audit sulle operazioni sensibili.
7. Notification preference center: preferenze notifiche, resend verification e tracking eventi.

## Architettura In Breve

```text
frontend
   |
   v
forgecore-gateway
   |
   +-- forgecore-auth
   +-- forgecore-payments
   +-- forgecore-subscriptions
   +-- forgecore-permissions
   +-- forgecore-config
   +-- forgecore-storage
   +-- forgecore-webhooks
   +-- forgecore-notifications
   +-- forgecore-admin
   +-- forgecore-audit
   +-- forgecore-jobs
```

Ogni servizio segue la direzione:

```text
transport -> application -> domain
infrastructure -> domain
```

`domain/` contiene regole pure e interfacce. `application/` orchestra use case. `infrastructure/` implementa database, Redis, NATS e provider esterni. `transport/` espone REST, gRPC, consumer e handler.

## Struttura Repository

```text
shared/                 SDK backend condivisa
sdk/go/                 client Go interni standardizzati
services/               microservizi ForgeCore
deployments/            Prometheus, Grafana, Traefik, OTEL, alerting
scripts/                build, smoke, check, scaffold, migrazioni
docs/forgecore/         specifiche, piani, runbook, OpenAPI, Postman
docker-compose.yml      stack locale
```

## Servizi

| Servizio | Gateway | REST locale | gRPC locale | Scopo |
| --- | --- | ---: | ---: | --- |
| `forgecore-gateway` | pubblico | 8080 | - | ingresso HTTP per frontend |
| `forgecore-auth` | `/v1/auth/*` | 8081 | 9091 | utenti, sessioni, JWT, MFA, OAuth |
| `forgecore-payments` | `/v1/payments/*` | 8082 | 9092 | pagamenti e Stripe |
| `forgecore-notifications` | `/v1/notifications/*` | 8083 | - | notifiche e SendGrid |
| `forgecore-admin` | `/v1/admin/*` | 8084 | - | backoffice |
| `forgecore-audit` | `/v1/audit/*` | 8085 | 9095 | audit append-only |
| `forgecore-jobs` | interno | - | - | worker Redis/NATS |
| `forgecore-permissions` | `/v1/permissions/*` | 8087 | 9097 | ruoli e permessi |
| `forgecore-config` | `/v1/config/*` | 8088 | 9098 | configurazioni runtime |
| `forgecore-webhooks` | `/v1/webhooks/*` | 8089 | - | endpoint e consegne webhook |
| `forgecore-storage` | `/v1/storage/*` | 8090 | - | metadata file e MinIO |
| `forgecore-subscriptions` | `/v1/subscriptions/*` | 8091 | 9099 | piani e abbonamenti |

Il frontend non dovrebbe chiamare direttamente le porte dei servizi. In uso normale passa sempre da `forgecore-gateway`.

## Avvio Rapido

Prerequisiti consigliati:

- Go target `1.26`
- Docker e Docker Compose
- PowerShell per gli script locali
- Bash, Git Bash o WSL per `scripts/migrate.sh`
- `make`, oppure esecuzione diretta degli script PowerShell

Configura l'ambiente:

```powershell
Copy-Item .env.example .env
```

Imposta le origini frontend consentite:

```env
CORS_ORIGIN=http://localhost:3000,http://localhost:5173
```

Esegui i controlli statici principali:

```powershell
make smoke
```

Avvia le dipendenze:

```powershell
docker compose up -d postgres redis nats minio prometheus grafana
```

Esegui le migrazioni:

```bash
./scripts/migrate.sh all up
```

Esegui il bootstrap locale:

```bash
SUPERADMIN_EMAIL=admin@forgecore.local SUPERADMIN_PASSWORD=ChangeMe123! ./scripts/bootstrap.sh
```

Avvia gateway e auth:

```powershell
docker compose up -d forgecore-auth forgecore-gateway
```

Verifica il gateway:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/readyz
```

Per avviare tutto:

```powershell
docker compose up -d
```

## Usare Da Frontend

Base URL locale:

```text
http://localhost:8080
```

Header principali:

- `Authorization: Bearer <access-token>` per route protette
- `X-Tenant-ID: <tenant-id>` quando il tenant non deriva dal token
- `X-Request-ID: <request-id>` per correlazione log e audit
- `Idempotency-Key: <stable-key>` per mutazioni ritentate

Route pubbliche:

- `GET /healthz`
- `GET /readyz`
- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/forgot-password`
- `POST /v1/auth/reset-password`
- `POST /v1/auth/verify-email`
- `POST /v1/auth/resend-verification`
- `GET /v1/auth/oauth/{provider}`
- `GET /v1/auth/oauth/{provider}/callback`

Documentazione utile:

- `docs/forgecore/frontend-api-readiness.md`
- `docs/forgecore/frontend-token-csrf-policy.md`
- `docs/forgecore/rbac-endpoint-matrix.md`
- `docs/forgecore/openapi/forgecore-gateway.v1.yaml`
- `docs/forgecore/frontend-testplan.md`
- `docs/forgecore/postman/ForgeCore.postman_collection.json`

## Postman

Importa la collection:

```text
docs/forgecore/postman/ForgeCore.postman_collection.json
```

Importa l'environment locale opzionale:

```text
docs/forgecore/postman/ForgeCore.local.postman_environment.json
```

La collection include:

- health e readiness
- auth register/login/refresh
- esempi protetti con bearer token
- pagamenti, subscriptions, config, permissions, webhooks, storage e admin
- richieste negative per missing token e CORS
- variabili Postman per `base_url`, `tenant_id`, `access_token`, `refresh_token`, `user_id`, `request_id` e `idempotency_key`

Il login salva automaticamente `access_token` e `refresh_token` se la risposta li contiene.

## SDK Condivisa

`shared/` contiene i package riutilizzabili:

| Package | Responsabilita' |
| --- | --- |
| `apperrors` | error model e mapping HTTP |
| `configloader` | composizione default, YAML ed ENV |
| `configschema` | schema, required, default e secret |
| `configsource` | sorgenti YAML, ENV e mappe |
| `crypto` | AES-256-GCM e HMAC per PII |
| `events` | eventi NATS tipizzati e versionati |
| `i18n` | formati data/importi |
| `middleware` | tenant, auth claims e request id |
| `observability` | log, metriche, tracing e shutdown |
| `pagination` | cursori e limiti |
| `postgres` | transazioni tenant-aware |
| `proto` | contratti gRPC |
| `validation` | validazione input |

## Configurazione

Le configurazioni sono caricate con questa priorita':

```text
default < YAML < ENV
```

File YAML opzionale:

```powershell
$env:FORGECORE_CONFIG_FILE = "C:\path\to\forgecore.yaml"
```

Esempio:

```yaml
port: ":8088"
database_url: "postgres://postgres:postgres@localhost:5432/config?sslmode=disable"
redis_addr: "localhost:6379"
```

Le variabili piu' importanti sono:

| Variabile | Scopo |
| --- | --- |
| `DATABASE_URL` | PostgreSQL usato dai servizi nel Compose |
| `*_DATABASE_URL` | URL specifici per migrazioni |
| `REDIS_ADDR` | Redis per auth, config e jobs |
| `NATS_URL` | NATS per eventi |
| `CORS_ORIGIN` | origini frontend consentite |
| `AUTH_GRPC_ADDR` | auth gRPC usato dal gateway |
| `*_SERVICE_URL` | upstream HTTP interni del gateway |
| `STRIPE_SECRET_KEY` | pagamenti e subscriptions |
| `STRIPE_WEBHOOK_SECRET` | verifica webhook Stripe |
| `SENDGRID_API_KEY` | invio notifiche |
| `MINIO_ENDPOINT` | storage oggetti |
| `SUPERADMIN_EMAIL` | seed admin locale |
| `SUPERADMIN_PASSWORD` | password admin locale |

I segreti devono essere letti tramite wrapper dedicati, per evitare logging accidentale:

```go
values.Secret("STRIPE_SECRET_KEY").Value()
```

## Sviluppare Un Nuovo Servizio

Genera uno scheletro DDD:

```powershell
make scaffold-dryrun name=forgecore-example
make scaffold name=forgecore-example
```

Regole principali:

- nome servizio `forgecore-<bounded-context>`
- modulo Go sotto `services/forgecore-<name>`
- niente import infrastrutturali in `domain/` o `application/`
- migrazioni tenant-aware con `tenant_id`, indice e RLS
- errori gestiti esplicitamente
- funzioni Go massimo 50 righe
- log applicativi in italiano

## Verifica

Comandi ufficiali:

```powershell
make verify
make build
make test-shared
make test-sdk
make test-auth
make test-e2e
make runtime-check
make dockerfile-check
make rbac-check
make security-check
make integration-check
make smoke
```

Script disponibili:

- `scripts/build-all.ps1`
- `scripts/check-boundaries.ps1`
- `scripts/check-proto-contracts.ps1`
- `scripts/check-sdk-clients.ps1`
- `scripts/check-tenant-migrations.ps1`
- `scripts/check-runtime-hardening.ps1`
- `scripts/check-dockerfiles.ps1`
- `scripts/check-rbac-security.ps1`
- `scripts/security-check.ps1`
- `scripts/e2e-gateway.ps1`
- `scripts/integration-local.ps1`
- `scripts/smoke-local.ps1`
- `scripts/scaffold-service.ps1`

Il test plan frontend completo e' in `docs/forgecore/frontend-testplan.md`.

## Documenti Importanti

- `docs/forgecore/studio-approfondito-sdk-backend.md`
- `docs/forgecore/plans/2026-05-01-forgecore-refactor-roadmap.md`
- `docs/forgecore/issues.md`
- `docs/forgecore/compatibility-matrix.md`
- `docs/forgecore/client-generation.md`
- `docs/forgecore/reliability-patterns.md`
- `docs/forgecore/security-baseline.md`
- `docs/forgecore/owasp-security-hardening.md`
- `docs/forgecore/release.md`
- `docs/forgecore/runbooks/tenant.md`
- `docs/forgecore/runbooks/webhooks.md`
- `docs/forgecore/runbooks/jobs.md`
- `docs/forgecore/runbooks/payments.md`
- `docs/forgecore/runbooks/audit.md`
- `docs/forgecore/runbooks/storage.md`
- `docs/forgecore/runbooks/recovery.md`

## Release

Le release seguono:

- changelog in `CHANGELOG.md`
- strategia in `docs/forgecore/release.md`
- ADR in `docs/forgecore/adr/`
- compatibility matrix aggiornata per breaking changes

## Contribuire

Leggi [CONTRIBUTING.md](./CONTRIBUTING.md) prima di aprire una PR.
