#!/bin/bash
# Popola Vault con i secrets dell'applicazione
# Eseguire UNA SOLA VOLTA su primo deploy
# Dipendenze: vault CLI installato, VAULT_ADDR e VAULT_TOKEN settati

set -euo pipefail

VAULT_ADDR=${VAULT_ADDR:-http://localhost:8200}
VAULT_TOKEN=${VAULT_TOKEN:-root}
export VAULT_ADDR VAULT_TOKEN

echo "Enabling KV secrets engine..."
vault secrets enable -path=secret kv-v2 || true

echo "Generating RSA key pair for JWT..."
openssl genrsa -out /tmp/jwt_private.pem 4096
openssl rsa -in /tmp/jwt_private.pem -pubout -out /tmp/jwt_public.pem

vault kv put secret/auth/jwt \
  private_key=@/tmp/jwt_private.pem \
  public_key=@/tmp/jwt_public.pem

rm /tmp/jwt_private.pem /tmp/jwt_public.pem

echo "Storing PII encryption key..."
PII_KEY=$(openssl rand -hex 32)
vault kv put secret/crypto pii_key="$PII_KEY"

echo "Storing pepper for HMAC..."
PEPPER=$(openssl rand -hex 32)
vault kv put secret/crypto/pepper value="$PEPPER"

echo "Placeholders for external service keys..."
vault kv put secret/payment/stripe \
  secret_key="sk_test_REPLACE_ME" \
  webhook_secret="whsec_REPLACE_ME"

vault kv put secret/notification/sendgrid \
  api_key="SG.REPLACE_ME"

vault kv put secret/notification/twilio \
  account_sid="AC_REPLACE_ME" \
  auth_token="REPLACE_ME" \
  from_number="+1REPLACE"

echo "Vault inizializzato con successo."
