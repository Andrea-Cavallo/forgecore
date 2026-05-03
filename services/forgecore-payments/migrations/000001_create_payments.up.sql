-- 000001_create_payments.up.sql
CREATE TABLE IF NOT EXISTS payments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    user_id        UUID NOT NULL,
    amount         BIGINT NOT NULL CHECK (amount > 0),
    currency       TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending',
    provider       TEXT NOT NULL,
    provider_id    TEXT NOT NULL DEFAULT '',
    failure_reason TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON payments(tenant_id);
CREATE INDEX ON payments(user_id, tenant_id);
CREATE UNIQUE INDEX payments_provider_id_idx ON payments(provider, provider_id) WHERE provider_id <> '';
CREATE UNIQUE INDEX payments_idempotency_key_idx ON payments(tenant_id, idempotency_key) WHERE idempotency_key <> '';

ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON payments
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
