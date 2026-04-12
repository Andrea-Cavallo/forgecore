#!/bin/bash
# Crea gli stream JetStream necessari
# Dipendenze: nats CLI installato, NATS_URL settato

set -euo pipefail

NATS_URL=${NATS_URL:-nats://localhost:4222}

create_stream() {
  local name=$1; shift
  echo "Creating stream $name..."
  nats stream add "$name" "$@" --server "$NATS_URL" 2>/dev/null || \
    echo "  Stream $name già esistente, skip."
}

create_stream AUTH_EVENTS \
  --subjects "auth.>" \
  --retention work \
  --max-age 7d \
  --replicas 1

create_stream PAYMENT_EVENTS \
  --subjects "payment.>" \
  --retention work \
  --max-age 30d \
  --replicas 1

create_stream AUDIT_EVENTS \
  --subjects "audit.>" \
  --retention limits \
  --max-age 365d \
  --replicas 1

create_stream NOTIFICATION_RETRY \
  --subjects "notification.retry" \
  --retention work \
  --max-age 24h \
  --replicas 1

create_stream WEBHOOK_EVENTS \
  --subjects "webhook.>" \
  --retention work \
  --max-age 7d \
  --replicas 1

echo "NATS JetStream stream inizializzati."
