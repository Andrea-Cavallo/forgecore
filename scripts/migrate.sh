#!/bin/bash
# Usage: ./scripts/migrate.sh <service> <direction>
# Example: ./scripts/migrate.sh forgecore-auth up

set -euo pipefail

SERVICE=${1:-}
DIRECTION=${2:-up}
SERVICES=(
  forgecore-auth
  forgecore-payments
  forgecore-notifications
  forgecore-audit
  forgecore-permissions
  forgecore-config
  forgecore-webhooks
  forgecore-storage
  forgecore-subscriptions
)

if [ -z "$SERVICE" ]; then
  echo "Uso: $0 <service> [up|down]"
  echo "Servizi: all ${SERVICES[*]}"
  exit 1
fi

database_url_for() {
  case "$1" in
    forgecore-auth)          echo "${AUTH_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-payments)      echo "${PAYMENT_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-notifications) echo "${NOTIFICATION_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-audit)         echo "${AUDIT_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-permissions)   echo "${PERMISSION_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-config)        echo "${CONFIG_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-webhooks)      echo "${WEBHOOK_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-storage)       echo "${STORAGE_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    forgecore-subscriptions) echo "${SUBSCRIPTION_DATABASE_URL:-postgres://forgecore:changeme@localhost:5432/forgecore?sslmode=disable}" ;;
    *) return 1 ;;
  esac
}

run_migration() {
  local service="$1"
  local db_url
  db_url="$(database_url_for "$service")"
  local migrations_path="services/${service}/migrations"

  if [ ! -d "$migrations_path" ]; then
    echo "Directory migrazioni non trovata: $migrations_path"
    exit 1
  fi

  echo "Esecuzione migrazione $DIRECTION per $service..."
  migrate -path "$migrations_path" -database "$db_url" "$DIRECTION"
  echo "Migrazione completata per $service."
}

if [ "$SERVICE" = "all" ]; then
  for svc in "${SERVICES[@]}"; do
    run_migration "$svc"
  done
  exit 0
fi

run_migration "$SERVICE"
