# Studio approfondito del progetto ForgeCore SDK Backend

## Scopo

`ForgeCore` e' una monorepo Go orientata a microservizi che deve fornire una SDK backend pronta all'uso. La SDK deve raccogliere funzionalita' comuni, coerenti e riutilizzabili per costruire servizi backend multi-tenant, sicuri, osservabili e facilmente componibili.

L'obiettivo non e' solo avere servizi separati, ma creare un sistema in cui ogni servizio usa le stesse primitive: identita' tenant, validazione, paginazione, eventi, logging, crittografia, middleware, protobuf, error handling e integrazioni infrastrutturali.

## Visione architetturale

La struttura segue un modello DDD a quattro livelli:

- `domain/`: entita', value object, regole di dominio e interfacce repository.
- `application/`: use case, orchestrazione di dominio, input/output applicativi.
- `infrastructure/`: implementazioni concrete verso PostgreSQL, Redis, NATS, provider esterni.
- `transport/`: REST, gRPC, consumer NATS, router e handler.

Le dipendenze devono puntare verso il dominio:

```text
transport -> application -> domain
infrastructure -> domain
```

Il dominio non deve importare librerie infrastrutturali. L'application layer dipende solo dal dominio e da porte applicative esplicite. L'infrastructure layer implementa le interfacce, senza far trapelare dettagli tecnici verso gli use case.

## Monorepo

La root di lavoro e':

```text
C:\Users\Andrea\Desktop\golang-modules
```

Il progetto `ForgeCore` contiene dodici moduli servizio con naming esplicito e non generico:

- `forgecore-gateway`
- `forgecore-auth`
- `forgecore-payments`
- `forgecore-notifications`
- `forgecore-admin`
- `forgecore-audit`
- `forgecore-jobs`
- `forgecore-permissions`
- `forgecore-config`
- `forgecore-webhooks`
- `forgecore-storage`
- `forgecore-subscriptions`

Ogni servizio deve essere autonomo, compilabile e dotato di:

- `go.mod`
- `Dockerfile`
- struttura DDD coerente
- repository interface nel dominio quando il servizio possiede dati
- use case cablati alle interfacce
- error handling esplicito
- log in italiano

## SDK condivisa

La SDK backend nasce soprattutto dai package condivisi in `shared/`. Questi package devono essere costruiti prima dei servizi, perche' rappresentano il contratto comune del sistema.

Package attesi:

- `shared/proto`: contratti `.proto` per i servizi con interfacce gRPC.
- `shared/events`: eventi tipizzati per NATS e integrazioni asincrone.
- `shared/middleware`: tenant context, auth check, request-id.
- `shared/validation`: wrapper coerenti su validator.
- `shared/crypto`: cifratura AES-256 per PII e segreti applicativi.
- `shared/pagination`: cursor pagination stabile, con encode/decode.
- `shared/i18n`: locale helper, formattazione importi e date.
- `shared/observability`: slog JSON, metriche Prometheus, graceful shutdown.

Questi package devono essere progettati come API pubbliche interne: nomi stabili, errori leggibili, documentazione minima ma utile, test mirati sulle regole critiche.

## Primo refactor: configurazione centralizzata

Il primo refactor architetturale di ForgeCore deve creare un solo punto di configurazione per l'intera piattaforma.

`forgecore-config` e' il bounded context dedicato alla configurazione runtime, tenant-specific e distribuita. Tutti gli altri servizi devono consumare configurazione tramite questo servizio o tramite package SDK condivisi, evitando loader locali duplicati.

Sorgenti supportate:

- file YAML
- variabili ENV
- configurazioni distribuite esposte da `forgecore-config`

Regole:

- ENV prevale su YAML quando entrambi definiscono la stessa chiave.
- Ogni configurazione deve avere schema, validazione e default espliciti.
- I messaggi di errore devono spiegare quale chiave e' invalida o mancante.
- I segreti devono usare tipi dedicati e non devono mai finire nei log.
- Il codice duplicato di env parsing, default, validazione e binding config deve essere eliminato.
- I package devono essere piccoli, con nomi parlanti e responsabilita' nette.

Naming consigliato:

```text
shared/configloader
shared/configschema
shared/configsource
services/forgecore-config
```

Evitare nomi generici come `utils`, `common`, `helpers`, `misc` o `shared/common`. Se un package carica YAML si chiama `configloader`; se descrive sorgenti si chiama `configsource`; se valida contratti si chiama `configschema`.

## Multi-tenancy

Ogni tabella applicativa deve contenere `tenant_id` e abilitare Row Level Security PostgreSQL.

Schema base:

```sql
CREATE TABLE <name> (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON <name>(tenant_id);

ALTER TABLE <name> ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON <name>
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
```

Ogni accesso DB deve impostare il tenant sulla connessione:

```go
conn.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
```

La SDK deve rendere questa pratica difficile da sbagliare, esponendo helper o middleware che propagano il tenant nel `context.Context` e lo applicano alle transazioni PostgreSQL.

## Principi implementativi

- Ogni funzione deve restare entro 50 righe.
- Usare guard clause invece di annidamenti profondi.
- Gestire sempre gli errori in modo esplicito.
- Evitare `_` per errori.
- Usare costanti nominate al posto di stringhe o numeri magici.
- Vietati placeholder come `// TODO implement`.
- Ogni file creato deve compilare.
- I log devono essere in italiano.

## Contratti di dominio

Per le entita' persistenti, il dominio espone repository interface con shape prevedibile:

```go
type EntityRepository interface {
    Create(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
    Update(ctx context.Context, entity *Entity) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Entity, error)
}
```

Gli use case ricevono interfacce, validano input, applicano regole di business e pubblicano eventi quando necessario.

## Responsabilita' dei servizi

### forgecore-gateway

Punto di ingresso HTTP. Gestisce routing, proxy verso servizi interni, middleware tenant/auth/request-id e mapping uniforme degli errori.

### forgecore-auth

Gestisce identita', autenticazione, JWT RS256, sessioni e Redis. E' fonte primaria per utenti, credenziali e token.

### forgecore-payments

Gestisce pagamenti e integrazione Stripe. Deve separare bene dominio pagamento, provider esterni e webhook.

### forgecore-notifications

Invio notifiche e gestione template/canali. Dipende da NATS e provider come SendGrid.

### forgecore-admin

Orchestrazione amministrativa. Consuma client verso auth e payment, espone funzioni operative per backoffice.

### forgecore-audit

Registro append-only degli eventi critici. Deve privilegiare immutabilita', tracciabilita' e ingestione asincrona.

### forgecore-jobs

Worker asincrono basato su Redis/asynq e NATS. La struttura principale e' `jobs/` e `scheduler/`, non necessariamente REST/gRPC.

### forgecore-permissions

Autorizzazioni, ruoli, policy e permessi tenant-aware.

### forgecore-config

Configurazioni applicative e tenant-specific, con caching Redis.

### forgecore-webhooks

Ricezione, verifica e inoltro webhook. Deve trattare firma, idempotenza e pubblicazione eventi come parti centrali.

### forgecore-storage

Gestione file, metadata PostgreSQL e oggetti MinIO. Deve separare metadata, policy e provider storage.

### forgecore-subscriptions

Abbonamenti, piani, stato subscription e integrazione Stripe.

## Naming ForgeCore

I servizi non devono usare nomi generici come `forgecore-payments` o `forgecore-auth`. Ogni modulo operativo deve portare il prefisso `forgecore-` per rendere chiara l'appartenenza alla piattaforma e ridurre ambiguita' in Docker, log, metriche, repository, moduli Go e documentazione.

Pattern:

```text
forgecore-<bounded-context>
```

Esempi:

- `forgecore-auth`
- `forgecore-payments`
- `forgecore-webhooks`
- `forgecore-subscriptions`

I nomi di container, metriche, logger, moduli, cartelle servizio e target build devono seguire questa convenzione.

## Eventi

Gli eventi devono essere tipizzati, versionati e pensati per NATS. La SDK deve evitare payload anonimi o mappe non tipizzate quando il dominio e' noto.

Ogni evento dovrebbe avere:

- nome costante
- versione
- tenant id
- correlation/request id
- timestamp
- payload tipizzato

## Osservabilita'

Ogni servizio deve usare logging strutturato, metriche e shutdown controllato. La SDK condivisa deve fornire setup coerente per:

- logger JSON
- request id
- graceful shutdown
- metriche Prometheus
- health/readiness endpoints

## Criteri di prontezza

Un task e' pronto solo quando:

- `go build ./...` passa per l'area toccata.
- Non restano placeholder.
- Le interfacce di dominio sono presenti.
- Gli use case usano interfacce e non implementazioni concrete.
- `go.mod` e `Dockerfile` sono validi.
- Non vengono introdotti problemi presenti in `docs/forgecore/issues.md`.
- Il piano relativo contiene checkbox Markdown per task e sotto-task.
- Le checkbox vengono spuntate solo dopo verifica effettiva, non per intenzione o avanzamento parziale.

## Regola piani

Ogni piano ForgeCore deve essere tracciabile tramite checkbox Markdown:

```markdown
- [ ] Task da completare
- [ ] Sotto-task verificabile
- [x] Task completato e verificato
```

Le sezioni descrittive possono restare in testo libero, ma ogni attivita' operativa deve avere una checkbox. Questo rende possibile lavorare per piccoli passi, riprendere il contesto senza ambiguita' e sapere sempre cosa e' davvero completato.

## Rischi principali

- Divergenza tra servizi se i package condivisi non sono costruiti prima.
- Duplicazione di middleware tenant/auth.
- Uso incoerente degli errori applicativi.
- Bypass accidentale della RLS per assenza di `SET LOCAL app.tenant_id`.
- Eventi NATS non versionati.
- Dipendenze infrastrutturali importate nel dominio.
- Skeleton che compila ma non rappresenta contratti utili.

## Aree di miglioramento

- Definire i contratti pubblici della SDK, distinguendo API stabili e dettagli interni.
- Curare la SDK ergonomics: API belle da usare, coerenti nei nomi e non solo tecnicamente corrette.
- Introdurre module boundaries enforcement con lint, test o regole statiche che impediscano import sbagliati tra layer.
- Introdurre un error model unico con codice, messaggio, wrapping, mapping HTTP e mapping gRPC.
- Centralizzare la configurazione in `forgecore-config` e nei package `shared/config*`, eliminando loader duplicati nei servizi.
- Rendere la tenant safety un comportamento di default tramite wrapper transazionali tenant-aware.
- Aggiungere un pattern condiviso per idempotenza in webhook, pagamenti, subscription, notifiche e job.
- Valutare un outbox pattern per pubblicare eventi NATS in modo affidabile dopo scritture DB.
- Standardizzare migrazioni SQL, RLS, indici, rollback e naming.
- Consolidare una policy di sicurezza SDK per crypto, segreti, JWT, firme webhook, rate limit, audit e log sanitizzati.
- Definire una strategia test comune: unit test dominio/application, integration test PostgreSQL con RLS, contract test per proto ed eventi.
- Versionare proto ed eventi con regole di compatibilita' backward e forward.
- Generare client Go interni per gRPC/REST, evitando client manuali divergenti.
- Aggiungere un service template generator per creare nuovi servizi gia' conformi alla struttura ForgeCore.
- Creare una cartella ADR per decisioni architetturali versionate e verificabili nel tempo.
- Mantenere una compatibility matrix tra versioni SDK, proto, servizi e schema DB.
- Definire una security baseline con threat model minimo per ogni servizio.
- Scrivere operational runbooks per debug di tenant, webhook, job, pagamenti, audit e storage.
- Rendere serio lo stack locale con Docker Compose, seed, migrazioni e healthcheck reali.
- Introdurre una release strategy con changelog, semantic versioning e gestione breaking changes.
- Rendere identica l'osservabilita' tra servizi: log, metriche, trace id, correlation id, health, readiness e shutdown.
- Usare una configurazione uniforme con env vars, default, validazione e gestione segreti.
- Chiarire ownership dei dati tra servizi, specialmente subscription/payment, auth/permission, storage/audit.
- Migliorare la developer experience con comandi comuni per build, test, lint, migrazioni, proto e scaffolding nuovi servizi.

## Direzione consigliata

La prima fase deve consolidare `shared/`. Solo dopo ha senso generare o completare i servizi. Ogni servizio dovrebbe essere piccolo ma completo: dominio chiaro, use case minimi, trasporti iniziali, Dockerfile e build pulita.

La SDK deve diventare il linguaggio comune del backend: ogni servizio deve sembrare scritto dalla stessa mano, con le stesse regole, gli stessi errori e gli stessi contratti.
