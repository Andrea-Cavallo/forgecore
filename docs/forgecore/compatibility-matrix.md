# ForgeCore Compatibility Matrix

| Area | Versione | Compatibilita' | Note |
| --- | --- | --- | --- |
| SDK shared | `v1` | Backward compatible dentro Go 1.26 | API pubbliche interne documentate nel README |
| Proto | `*.v1` | Campi esistenti non vanno rinumerati o rimossi | Aggiunte solo con nuovi field number |
| Eventi NATS | `*.v1` | Event name versionato con suffisso `.v1` | Payload mantiene `tenant_id`, `correlation_id`, `occurred_at` |
| Schema DB | `000001+` | Migrazioni append-only | RLS obbligatoria per tabelle tenant-aware |
| Servizi | `forgecore-*` | Import path Go allineati al naming ForgeCore | Docker Compose usa rete `forgecore` |

Regole:

- Non riutilizzare numeri campo proto.
- Non rimuovere campi JSON gia' pubblicati negli eventi.
- Aggiungere nuovi eventi con nuovo nome versionato quando cambia semantica.
- Ogni tabella applicativa deve avere `tenant_id`, indice, RLS e policy `tenant_isolation`.
