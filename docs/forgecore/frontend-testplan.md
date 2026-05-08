# Frontend Test Plan

Questo piano verifica che ForgeCore sia pronto per essere usato da siti e applicazioni frontend attraverso `forgecore-gateway`.

Ogni task va spuntato solo dopo verifica effettiva.

## Obiettivo

- [ ] Confermare che il frontend usi solo `forgecore-gateway` come API pubblica.
- [ ] Confermare che health, readiness, auth, RBAC, tenant isolation, idempotenza, storage, paginazione, osservabilita' e sicurezza siano verificati.
- [ ] Confermare che OpenAPI, Postman e futura SDK TypeScript descrivano la stessa superficie API.

## Prerequisiti

- [ ] Docker e Docker Compose installati.
- [ ] Go target `1.26` disponibile.
- [ ] PowerShell disponibile per gli script locali.
- [ ] Bash, Git Bash o WSL disponibile per `scripts/migrate.sh`.
- [ ] File `.env` creato da `.env.example`.
- [ ] `CORS_ORIGIN` configurato con l'origine frontend usata nei test.
- [ ] Porta locale `5432` libera oppure PostgreSQL locale riconfigurato per evitare conflitti con Compose.

## Verifiche Statiche

- [ ] Eseguire `make smoke`.
- [ ] Eseguire `make rbac-check`.
- [ ] Eseguire `make security-check`.
- [ ] Verificare che boundary DDD, proto, SDK clients, tenant migrations, runtime hardening, Dockerfile e security check passino.
- [ ] Verificare che `docs/forgecore/openapi/forgecore-gateway.v1.yaml` sia coerente con le route documentate.
- [ ] Verificare che `docs/forgecore/rbac-endpoint-matrix.md` copra tutte le route protette frontend-facing.

## Build E Test Go

- [ ] Eseguire `make build`.
- [ ] Eseguire `make test-shared`.
- [ ] Eseguire `make test-sdk`.
- [ ] Eseguire `make test-auth`.
- [ ] Eseguire `make test-e2e`.
- [ ] Confermare che ogni modulo toccato compili con `go build ./...`.

## Runtime Locale

- [ ] Avviare dipendenze con `docker compose up -d postgres redis nats minio prometheus grafana`.
- [ ] Eseguire `./scripts/migrate.sh all up`.
- [ ] Eseguire `SUPERADMIN_EMAIL=admin@forgecore.local SUPERADMIN_PASSWORD=ChangeMe123! ./scripts/bootstrap.sh`.
- [ ] Avviare `forgecore-auth` e `forgecore-gateway`.
- [ ] Verificare `GET http://localhost:8080/healthz`.
- [ ] Verificare `GET http://localhost:8080/readyz`.
- [ ] Avviare lo stack completo con `docker compose up -d`.
- [ ] Eseguire `make integration-check` quando tutte le porte sono libere.

## Postman

- [ ] Importare `docs/forgecore/postman/ForgeCore.postman_collection.json`.
- [ ] Impostare `base_url` a `http://localhost:8080`.
- [ ] Impostare `tenant_id`, `user_id`, `admin_user_id`, `plan_id`, `payment_id`, `subscription_id`, `permission_id` e `file_id` con valori validi o demo.
- [ ] Eseguire la cartella `Health`.
- [ ] Eseguire `Auth / Register`.
- [ ] Eseguire `Auth / Login` e verificare che `access_token` e `refresh_token` vengano salvati nelle variabili.
- [ ] Eseguire `Auth / Refresh`.
- [ ] Eseguire almeno una route protetta con bearer token valido.
- [ ] Eseguire la cartella `Negative And Security`.

## Auth Frontend Flow

- [ ] Registrare un utente con `POST /v1/auth/register`.
- [ ] Eseguire login con `POST /v1/auth/login`.
- [ ] Salvare access token in memoria nel frontend.
- [ ] Salvare refresh token secondo `docs/forgecore/frontend-token-csrf-policy.md`.
- [ ] Eseguire refresh con `POST /v1/auth/refresh`.
- [ ] Verificare logout e revoca sessione quando l'endpoint applicativo e' cablato nel gateway.
- [ ] Verificare `me` o validazione token protetta quando l'endpoint applicativo e' esposto.
- [ ] Verificare forgot password, reset password, verify email e resend verification.
- [ ] Verificare callback OAuth per provider supportati.

## Route Protette

- [ ] Chiamare una route protetta senza token e verificare `401` con `code`, `message` e `request_id`.
- [ ] Chiamare una route protetta con token non valido e verificare `401`.
- [ ] Chiamare una route protetta con token valido ma ruolo insufficiente e verificare `403`.
- [ ] Chiamare una route protetta con token valido e ruolo corretto e verificare successo o errore applicativo atteso.
- [ ] Confermare che il gateway propaghi `X-User-ID`, `X-Tenant-ID` e `X-User-Roles` verso gli upstream.

## Tenant Isolation

- [ ] Creare dati per tenant A.
- [ ] Creare dati per tenant B.
- [ ] Accedere come utente tenant A e verificare che i dati tenant B non siano visibili.
- [ ] Accedere come utente tenant B e verificare che i dati tenant A non siano visibili.
- [ ] Verificare che le migrazioni abbiano `tenant_id`, indice tenant e RLS.

## Idempotenza

- [ ] Inviare due volte una mutazione di pagamento con lo stesso `Idempotency-Key`.
- [ ] Confermare che non vengano creati doppi side effect.
- [ ] Ripetere il test per subscription o webhook-facing command.
- [ ] Verificare metriche o log di idempotency hit/miss.

## Pagamenti E Subscriptions

- [ ] Creare un pagamento con `POST /v1/payments`.
- [ ] Listare pagamenti con `GET /v1/payments`.
- [ ] Eseguire refund con `POST /v1/payments/{id}/refund`.
- [ ] Creare subscription con `POST /v1/subscriptions`.
- [ ] Cancellare subscription con `DELETE /v1/subscriptions/{id}`.
- [ ] Verificare che le mutazioni sensibili producano audit log.

## Config, Permissions E Admin

- [ ] Salvare una configurazione tenant con `PUT /v1/config/{key}`.
- [ ] Leggere una configurazione tenant con `GET /v1/config/{key}`.
- [ ] Eseguire permission check con `POST /v1/permissions/check`.
- [ ] Assegnare permesso con `POST /v1/permissions/grant`.
- [ ] Revocare permesso con `DELETE /v1/permissions/{id}`.
- [ ] Listare tenant admin con `GET /v1/admin/tenants`.
- [ ] Leggere stats admin con `GET /v1/admin/stats`.

## Storage E Webhook

- [ ] Eseguire upload multipart con `POST /v1/storage/upload`.
- [ ] Generare presigned URL con `GET /v1/storage/{id}/presign`.
- [ ] Registrare endpoint webhook con `POST /v1/webhooks/endpoints`.
- [ ] Consegnare evento webhook con `POST /v1/webhooks/deliver`.
- [ ] Verificare blocco URL webhook non sicuri o privati.

## Paginazione

- [ ] Chiamare una lista con limite esplicito.
- [ ] Verificare ordinamento stabile.
- [ ] Verificare cursore successivo.
- [ ] Verificare errore con cursore non valido.
- [ ] Verificare normalizzazione del limite massimo.

## Osservabilita'

- [ ] Inviare `X-Request-ID` dal frontend.
- [ ] Verificare che `X-Request-ID` sia presente nei log gateway.
- [ ] Verificare che audit log includa tenant, user e request id per mutazioni sensibili.
- [ ] Verificare metriche Prometheus per gateway e servizi principali.
- [ ] Verificare assenza di token, password e segreti nei log.

## Sicurezza Browser

- [ ] Verificare che un'origine CORS non consentita venga bloccata.
- [ ] Verificare che un'origine CORS consentita riceva gli header corretti.
- [ ] Verificare security headers su route pubbliche e protette.
- [ ] Verificare policy token e CSRF scelta per il frontend.
- [ ] Verificare che refresh token non venga salvato in `localStorage`.
- [ ] Verificare che state-changing cookie-authenticated requests richiedano CSRF token.

## TypeScript Client

- [ ] Generare client TypeScript da `docs/forgecore/openapi/forgecore-gateway.v1.yaml`.
- [ ] Compilare il client in un'app React o Next.js demo.
- [ ] Eseguire login usando il client.
- [ ] Eseguire refresh token usando il client.
- [ ] Eseguire una route protetta usando il client.
- [ ] Documentare la compatibilita' tra versione OpenAPI e versione client.

## Gate Finale

- [ ] Eseguire `make verify`.
- [ ] Eseguire `make smoke`.
- [ ] Eseguire `make security-check`.
- [ ] Eseguire `make integration-check` con runtime completo disponibile.
- [ ] Eseguire la collection Postman completa.
- [ ] Eseguire test browser Playwright o Cypress quando il frontend harness esiste.
- [ ] Aggiornare README, OpenAPI, Postman e questo test plan se una route o un contratto cambiano.
