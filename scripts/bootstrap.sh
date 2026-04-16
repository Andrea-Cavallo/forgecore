#!/usr/bin/env bash
# bootstrap.sh — avvia l'intero stack Superpowers in ordine corretto.
# Uso: ./scripts/bootstrap.sh [--skip-vault] [--skip-nats] [--skip-seed]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(dirname "$SCRIPT_DIR")"

SKIP_VAULT=false
SKIP_NATS=false
SKIP_SEED=false

for arg in "$@"; do
  case "$arg" in
    --skip-vault) SKIP_VAULT=true ;;
    --skip-nats)  SKIP_NATS=true ;;
    --skip-seed)  SKIP_SEED=true ;;
  esac
done

log() { echo "[bootstrap] $*"; }

# 1. Vault init
if [ "$SKIP_VAULT" = false ]; then
  log "=== Inizializzazione Vault ==="
  bash "$SCRIPT_DIR/vault-init.sh"
fi

# 2. NATS JetStream streams
if [ "$SKIP_NATS" = false ]; then
  log "=== Inizializzazione NATS JetStream ==="
  bash "$SCRIPT_DIR/nats-init.sh"
fi

# 3. Migrazioni database (tutti i servizi)
log "=== Esecuzione migrazioni PostgreSQL ==="
bash "$SCRIPT_DIR/migrate.sh" all up

# 4. Seed super-admin (primo tenant + utente owner)
if [ "$SKIP_SEED" = false ]; then
  log "=== Seed super-admin ==="
  ADMIN_EMAIL="${SUPERADMIN_EMAIL:-admin@superpowers.io}"
  ADMIN_PASSWORD="${SUPERADMIN_PASSWORD:-}"
  if [ -z "$ADMIN_PASSWORD" ]; then
    log "ATTENZIONE: SUPERADMIN_PASSWORD non impostata — seed saltato."
  else
    curl -sf -X POST "http://localhost:8081/v1/auth/register" \
      -H "Content-Type: application/json" \
      -d "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$ADMIN_PASSWORD\",\"tenant_name\":\"superpowers\"}" \
      && log "Super-admin creato: $ADMIN_EMAIL" \
      || log "ATTENZIONE: registrazione super-admin fallita (forse già esiste)."
  fi
fi

log "=== Bootstrap completato ==="
