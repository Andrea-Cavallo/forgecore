#!/bin/bash
# Usage: ./scripts/migrate.sh <service> <direction>
# Example: ./scripts/migrate.sh forgecore-auth up

set -euo pipefail

SERVICE=${1:-}
DIRECTION=${2:-up}

if [ -z "$SERVICE" ]; then
  echo "Uso: $0 <service> [up|down]"
  echo "Servizi: forgecore-auth forgecore-payments forgecore-notifications forgecore-admin forgecore-audit"
  echo "         forgecore-jobs forgecore-gateway forgecore-permissions forgecore-config forgecore-webhooks"
  echo "         forgecore-storage forgecore-subscriptions"
  exit 1
fi

case "$SERVICE" in
  forgecore-auth)         DB_URL="${AUTH_DATABASE_URL:?AUTH_DATABASE_URL non impostata}" ;;
  forgecore-payments)      DB_URL="${PAYMENT_DATABASE_URL:?PAYMENT_DATABASE_URL non impostata}" ;;
  forgecore-notifications) DB_URL="${NOTIFICATION_DATABASE_URL:?NOTIFICATION_DATABASE_URL non impostata}" ;;
  forgecore-admin)        DB_URL="${ADMIN_DATABASE_URL:?ADMIN_DATABASE_URL non impostata}" ;;
  forgecore-audit)        DB_URL="${AUDIT_DATABASE_URL:?AUDIT_DATABASE_URL non impostata}" ;;
  forgecore-jobs)          DB_URL="${JOB_DATABASE_URL:?JOB_DATABASE_URL non impostata}" ;;
  forgecore-gateway)          DB_URL="${GATEWAY_DATABASE_URL:?GATEWAY_DATABASE_URL non impostata}" ;;
  forgecore-permissions)   DB_URL="${PERMISSION_DATABASE_URL:?PERMISSION_DATABASE_URL non impostata}" ;;
  forgecore-config)       DB_URL="${CONFIG_DATABASE_URL:?CONFIG_DATABASE_URL non impostata}" ;;
  forgecore-webhooks)      DB_URL="${WEBHOOK_DATABASE_URL:?WEBHOOK_DATABASE_URL non impostata}" ;;
  forgecore-storage)      DB_URL="${STORAGE_DATABASE_URL:?STORAGE_DATABASE_URL non impostata}" ;;
  forgecore-subscriptions) DB_URL="${SUBSCRIPTION_DATABASE_URL:?SUBSCRIPTION_DATABASE_URL non impostata}" ;;
  *)
    echo "Servizio sconosciuto: $SERVICE"
    exit 1
    ;;
esac

MIGRATIONS_PATH="services/${SERVICE}/migrations"

if [ ! -d "$MIGRATIONS_PATH" ]; then
  echo "Directory migrazioni non trovata: $MIGRATIONS_PATH"
  exit 1
fi

echo "Esecuzione migrazione $DIRECTION per $SERVICE..."
migrate -path "$MIGRATIONS_PATH" -database "$DB_URL" "$DIRECTION"
echo "Migrazione completata."
