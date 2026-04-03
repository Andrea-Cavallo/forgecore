# Issues — Go Code Review (2026-04-03)

Generato da `/go-review project`. Da risolvere prima di ogni merge/deploy.

---

## CRITICAL (blocca compilazione o introduce vulnerabilità)

- **C-1** — `go.mod` × 10 servizi: manca `replace github.com/yourorg/golang-modules/shared => ../../shared`. Il progetto non compila.
- **C-2** — `shared/validation/validator.go:36`: bare type assertion `err.(validator.ValidationErrors)` — panica su errori inattesi. Usare `errors.As`.
- **C-3** — `webhook-service/internal/application/register_endpoint.go:19`: nessuna validazione di schema/host nell'URL — **SSRF**.
- **C-4** — `shared/crypto/pii.go:64`: SHA-256 senza salt/pepper sull'email — vulnerabile a rainbow table. Usare HMAC-SHA256 con secret pepper.
- **C-5** — `shared/observability/shutdown.go:21`: `os.Exit(0)` salta tutti i `defer` (disconnect DB, flush span, ecc.). Tornare normalmente e lasciare che `main` esca.

---

## HIGH (bug silenzioso o violazione architetturale)

- **H-1** — `auth-service/internal/application/login.go:76`: `_ = sessions.Create(...)` — la sessione non viene persistita silenziosamente.
- **H-2** — `payment-service/internal/application/create_payment.go:72`: record di fallimento pagamento silenziosamente perso.
- **H-3** — `webhook-service/internal/application/deliver.go:64`: `_ = deliveries.Create(...)` — record di consegna perso silenziosamente.
- **H-4** — `auth-service/internal/application/register.go:54`: errore DB trattato come "utente non trovato" → duplicati.
- **H-5** — `subscription-service/internal/application/subscribe.go:48`: errore DB trattato come "nessun abbonamento" → doppia fatturazione.
- **H-6** — `shared/observability/metrics.go:32`: `prometheus.MustRegister` panica nei test se chiamato più volte.
- **H-7** — `api-gateway/internal/middleware/ratelimit.go:53`: `r.RemoteAddr` include la porta + nessuna eviction → memory leak illimitato.
- **H-8** — `payment-service/internal/application/create_payment.go:51`: funzione di 51 righe (limite ≤50).
- **H-9** — `domain/repository.go` × 7 servizi: il layer domain importa `shared/pagination` — **violazione DDD** (domain deve avere zero dipendenze esterne).

---

## MEDIUM

- **M-1** — `shared/observability/shutdown.go:13`: doppia registrazione signal handler.
- **M-2** — `shared/middleware/tenant.go:37`: valore raw dell'header echeggiato nella risposta 400 (reflection risk).
- **M-3** — `shared/pagination/cursor.go:29`: errore di `json.Marshal` scartato con `_`.
- **M-4** — `webhook-service/internal/application/deliver.go:91`: valore di ritorno di `mac.Write` non gestito.
- **M-5** — `api-gateway/internal/router/router.go:25`: errore di `w.Write` scartato.
- **M-6** — `shared/observability/tracer.go:17`: `WithInsecure()` hardcoded — non sicuro in produzione.
- **M-7** — `shared/validation/validator.go:10`: validator come variabile package-level mutabile (singleton globale).
- **M-8** — `notification-service/internal/application/send.go:12`: `maxRetryAttempts` dichiarato ma mai usato.
- **M-9** — `shared/i18n/locale.go:19`: locale IT identica al default (dead code + separatore decimale sbagliato).
- **M-10** — Intero repo: **zero file di test** — copertura 0% vs. requisito 80%.

---

## LOW

- **L-1** — `job-service/internal/jobs/registry.go:41`: `MustMarshal` (solo per test) nel package di produzione.
- **L-2** — `auth-service/internal/application/login.go:74`: magic number `7 * 24 * time.Hour` — usare costante nominata.
- **L-3** — `payment-service/internal/application/create_payment.go:63`: magic string `"stripe"` — usare `domain.ProviderStripe`.
- **L-4** — `config-service/internal/application/get_config.go:29`: `cacheTTLSeconds` non condiviso.
- **L-5** — `shared/validation/validator.go:47`: messaggi di validazione in inglese (convenzione progetto: italiano).

---

## Azioni minime prima di sbloccare il progetto

1. Aggiungere `replace` directive in tutti e 10 i `go.mod` dei servizi + `go mod tidy` (C-1)
2. `errors.As` al posto della bare assertion in `validator.go:36` (C-2)
3. Validazione URL con allowlist scheme HTTPS + blocco IP privati/loopback in `register_endpoint.go` (C-3)
4. HMAC-SHA256 con pepper in `pii.go` (C-4)
5. Rimuovere `os.Exit` da `shutdown.go` (C-5)
6. Gestire tutti gli errori scartati con `_` in auth, payment, webhook (H-1 → H-5)
7. Rate limiter: estrarre solo IP con `net.SplitHostPort` + eviction periodica (H-7)
