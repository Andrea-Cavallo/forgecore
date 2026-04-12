#!/bin/bash
# Usage: ./scripts/migrate.sh <service> <direction>
# Example: ./scripts/migrate.sh auth-service up

set -euo pipefail

SERVICE=${1:-}
DIRECTION=${2:-up}

if [ -z "$SERVICE" ]; then
  echo "Uso: $0 <service> [up|down]"
  echo "Servizi: auth-service payment-service notification-service admin-service audit-service"
  echo "         job-service api-gateway permission-service config-service webhook-service"
  echo "         storage-service subscription-service"
  exit 1
fi

case "$SERVICE" in
  auth-service)         DB_URL="${AUTH_DATABASE_URL:?AUTH_DATABASE_URL non impostata}" ;;
  payment-service)      DB_URL="${PAYMENT_DATABASE_URL:?PAYMENT_DATABASE_URL non impostata}" ;;
  notification-service) DB_URL="${NOTIFICATION_DATABASE_URL:?NOTIFICATION_DATABASE_URL non impostata}" ;;
  admin-service)        DB_URL="${ADMIN_DATABASE_URL:?ADMIN_DATABASE_URL non impostata}" ;;
  audit-service)        DB_URL="${AUDIT_DATABASE_URL:?AUDIT_DATABASE_URL non impostata}" ;;
  job-service)          DB_URL="${JOB_DATABASE_URL:?JOB_DATABASE_URL non impostata}" ;;
  api-gateway)          DB_URL="${GATEWAY_DATABASE_URL:?GATEWAY_DATABASE_URL non impostata}" ;;
  permission-service)   DB_URL="${PERMISSION_DATABASE_URL:?PERMISSION_DATABASE_URL non impostata}" ;;
  config-service)       DB_URL="${CONFIG_DATABASE_URL:?CONFIG_DATABASE_URL non impostata}" ;;
  webhook-service)      DB_URL="${WEBHOOK_DATABASE_URL:?WEBHOOK_DATABASE_URL non impostata}" ;;
  storage-service)      DB_URL="${STORAGE_DATABASE_URL:?STORAGE_DATABASE_URL non impostata}" ;;
  subscription-service) DB_URL="${SUBSCRIPTION_DATABASE_URL:?SUBSCRIPTION_DATABASE_URL non impostata}" ;;
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
