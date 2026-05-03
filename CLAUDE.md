# CLAUDE.md

## Identita' del progetto

Questo repository contiene `ForgeCore`, una monorepo Go per una SDK backend e un set di microservizi production-ready. Lo scopo e' fornire primitive coerenti, riutilizzabili e pronte all'uso per backend multi-tenant: autenticazione, pagamenti, permessi, configurazioni, audit, notifiche, storage, webhook, job asincroni, osservabilita', validazione, eventi e paginazione.

Root di lavoro:

```text
C:\Users\Andrea\Desktop\golang-modules
```

## Regole operative

1. Prima di rispondere a domande di architettura o codice, consulta la knowledge graph in `graphify-out/`.
2. Se esiste `graphify-out/wiki/index.md`, naviga quella wiki prima di leggere file grezzi.
3. Leggi file grezzi solo quando l'utente lo chiede esplicitamente o quando la wiki/graphify non basta per completare una modifica sicura.
4. Prima di modificare codice, consulta `docs/forgecore/issues.md`.
5. Tutti i file di piano devono usare checkbox Markdown per ogni attivita' eseguibile.
6. Ogni task di piano deve essere spuntato da `[ ]` a `[x]` solo dopo verifica effettiva.
7. Non considerare completata una fase se i relativi sotto-task non hanno checkbox spuntate.
8. Dopo ogni modifica rilevante, aggiorna `README.md` e verifica che sia coerente con struttura, naming e stato reale del progetto.
9. Prima di chiudere un task di codice, verifica che le build funzionino per l'area toccata.
10. Dopo modifiche a file di codice, aggiorna graphify con:

```bash
python3 -c "from graphify.watch import _rebuild_code; from pathlib import Path; _rebuild_code(Path('.'))"
```

## Architettura obbligatoria

Ogni servizio segue DDD a quattro livelli:

```text
transport/ -> application/ -> domain/
infrastructure/ -> domain/
```

- `domain/`: entita', value object, repository interface, regole pure. Nessuna dipendenza esterna infrastrutturale.
- `application/`: use case, validazione input, orchestrazione dominio.
- `infrastructure/`: PostgreSQL, Redis, NATS, provider esterni, implementazioni repository.
- `transport/`: REST, gRPC, consumer, router, middleware di ingresso.

## Standard Go

- Go target: `1.26`.
- Ogni funzione deve essere al massimo 50 righe.
- Preferire guard clause a `if` annidati.
- Gestire sempre gli errori in modo esplicito.
- Non usare `_` per ignorare errori.
- Usare costanti nominate al posto di stringhe e numeri magici.
- Non lasciare placeholder come `// TODO implement`.
- Ogni file creato o modificato deve compilare.
- I log applicativi devono essere in italiano.

## Package condivisi SDK

Costruire e mantenere prima `shared/`, perche' e' la SDK usata dai servizi:

```text
shared/
├── proto/
├── events/
├── middleware/
├── validation/
├── crypto/
├── pagination/
├── i18n/
└── observability/
```

Responsabilita':

- `proto`: contratti gRPC e messaggi condivisi.
- `events`: eventi NATS tipizzati e versionati.
- `middleware`: tenant context, auth check, request-id.
- `validation`: wrapper e pattern di validazione.
- `crypto`: helper AES-256 per PII.
- `pagination`: cursor pagination encode/decode.
- `i18n`: locale, date e importi.
- `observability`: slog JSON, Prometheus, graceful shutdown.

## Primo refactor architetturale

Il primo refactor ForgeCore deve introdurre un solo punto di configurazione per tutta la piattaforma.

Regole:

- Esiste un unico servizio/modulo di configurazione: `forgecore-config`.
- Tutti i servizi leggono configurazioni da `forgecore-config` o dalla SDK condivisa di config, non da logiche duplicate locali.
- Le configurazioni devono poter arrivare da YAML e variabili ENV.
- Le variabili ENV hanno precedenza sui valori YAML quando entrambe definiscono la stessa chiave.
- Ogni configurazione deve avere schema, validazione, default espliciti e messaggi di errore chiari.
- I segreti non devono essere loggati e devono avere tipi/config wrapper dedicati.
- Il refactor deve eliminare codice duplicato di config loading, env parsing, default e validazione se gia' presente.
- I package devono avere naming parlante, con responsabilita' piccola e confini chiari.
- Evitare package generici come `utils`, `common`, `helpers` quando si puo' usare un nome di dominio tecnico preciso.

Naming consigliato per config:

```text
shared/configloader
shared/configschema
shared/configsource
services/forgecore-config
```

Il servizio `forgecore-config` governa configurazioni runtime, tenant-specific e distribuite. I package `shared/config*` forniscono API locali per caricare, validare e comporre YAML/ENV in modo uniforme.

## Aree di miglioramento da presidiare

- La cartella `shared/` deve essere una SDK con API stabili, non una raccolta casuale di utility.
- Le API della SDK devono essere ergonomiche: nomi chiari, composizione semplice e comportamento prevedibile.
- I confini dei moduli devono essere verificati da lint, test o regole statiche.
- Errori, configurazione, osservabilita', paginazione, validazione, tenant context e pubblicazione eventi devono avere pattern comuni.
- La configurazione deve essere centralizzata in `forgecore-config` e nei package `shared/config*`, eliminando duplicazioni servizio per servizio.
- Ogni scrittura che produce eventi deve considerare affidabilita', idempotenza e possibilita' di outbox pattern.
- Proto ed eventi devono essere versionati e compatibili tra release.
- I client Go interni per gRPC/REST devono essere generati o standardizzati, non riscritti a mano servizio per servizio.
- Deve esistere un generatore template per nuovi servizi conformi a ForgeCore.
- Le decisioni architetturali importanti devono vivere in una cartella ADR.
- Deve esistere una compatibility matrix tra SDK, proto, servizi e schema DB.
- Ogni servizio deve rispettare una security baseline con threat model minimo.
- I runbook operativi devono spiegare come debuggare tenant, webhook, job, pagamenti, audit e storage.
- Lo stack locale deve includere Docker Compose, seed, migrazioni e healthcheck reali.
- Le release devono seguire changelog, semantic versioning e gestione esplicita delle breaking changes.
- Migrazioni SQL e policy RLS devono seguire naming e struttura coerenti.
- I servizi devono avere ownership dati chiara, senza duplicare responsabilita' di dominio.
- La developer experience deve restare curata con comandi comuni per build, test, lint, migrazioni, proto e scaffolding.

## Multi-tenancy

Ogni tabella applicativa deve avere `tenant_id`, indice dedicato e Row Level Security.

Pattern SQL:

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

Ogni accesso PostgreSQL tenant-aware deve impostare:

```go
conn.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID)
```

## Servizi

| Servizio | REST | gRPC | Focus |
| --- | ---: | ---: | --- |
| `forgecore-gateway` | 8080 | - | proxy, router, middleware |
| `forgecore-auth` | 8081 | 9091 | utenti, sessioni, JWT RS256 |
| `forgecore-payments` | 8082 | 9092 | pagamenti, Stripe |
| `forgecore-notifications` | 8083 | - | notifiche, NATS, SendGrid |
| `forgecore-admin` | 8084 | - | backoffice, client interni |
| `forgecore-audit` | 8085 | 9095 | audit append-only, NATS |
| `forgecore-jobs` | - | - | worker Redis/asynq, scheduler |
| `forgecore-permissions` | 8087 | 9097 | ruoli, permessi, policy |
| `forgecore-config` | 8088 | 9098 | configurazioni, Redis cache |
| `forgecore-webhooks` | 8089 | - | webhook, firme, idempotenza |
| `forgecore-storage` | 8090 | - | MinIO, metadata PostgreSQL |
| `forgecore-subscriptions` | 8091 | 9099 | abbonamenti, Stripe |

## Naming ForgeCore

Non usare nomi generici come `forgecore-payments`, `forgecore-auth` o `forgecore-gateway` per nuovi moduli, container, metriche o documentazione. Ogni servizio deve usare il prefisso `forgecore-`.

Pattern:

```text
forgecore-<bounded-context>
```

Il nome deve essere coerente in cartella servizio, modulo Go, Docker Compose, logger, metriche, healthcheck, documentazione e runbook.

## Modulo servizio

Ogni servizio deve avere un `go.mod` con modulo:

```text
github.com/Andrea-Cavallo/golang-modules/services/<forgecore-service-name>
```

## Repository interface

Per entita' persistenti, usare questo pattern nel dominio:

```go
type EntityRepository interface {
    Create(ctx context.Context, entity *Entity) error
    GetByID(ctx context.Context, id uuid.UUID) (*Entity, error)
    Update(ctx context.Context, entity *Entity) error
    Delete(ctx context.Context, id uuid.UUID) error
    ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Entity, error)
}
```

## Eventi

Ogni evento condiviso deve avere:

- nome costante
- versione
- `tenant_id`
- `correlation_id` o `request_id`
- timestamp
- payload tipizzato

Evitare `map[string]any` quando il contratto e' noto.

## Definition of Done

Un task e' completato solo quando:

- `go build ./...` passa per il codice toccato.
- Non ci sono placeholder.
- Il dominio espone le interfacce necessarie.
- Gli use case sono cablati alle interfacce.
- `go.mod` e `Dockerfile` sono presenti nei servizi creati.
- `README.md` e' coerente con naming, struttura e comandi reali.
- `docs/forgecore/issues.md` non segnala problemi introdotti.
- I problemi corretti vengono marcati come risolti in `issues.md`.

## Ordine di lavoro consigliato

1. Consultare graphify.
2. Consultare `docs/forgecore/issues.md` prima di cambiare codice.
3. Usare come piano canonico `docs/forgecore/plans/2026-05-01-forgecore-refactor-roadmap.md`.
4. Verificare che il piano usi checkbox per ogni task e sotto-task.
5. Completare il primo task aperto del piano canonico.
6. Prima di creare file nuovi, verificare se esiste gia' una base da rifattorizzare.
7. Costruire prima `shared/` e `forgecore-config`, poi lavorare sui servizi.
8. Verificare build e assenza di placeholder.
9. Spuntare nel piano solo cio' che e' stato completato e verificato.
10. Aggiornare `README.md` se naming, struttura, comandi, servizi o stato del progetto sono cambiati.
11. Aggiornare graphify dopo modifiche a codice.

## File di riferimento

- Studio progetto ForgeCore: `docs/forgecore/studio-approfondito-sdk-backend.md`
- Piano canonico refactor ForgeCore: `docs/forgecore/plans/2026-05-01-forgecore-refactor-roadmap.md`
- Piano microservizi: `docs/forgecore/plans/2026-03-30-microservices-implementation.md`
- Specifiche microservizi: `docs/forgecore/specs/2026-03-30-microservices-design.md`
- Registro problemi: `docs/forgecore/issues.md`
