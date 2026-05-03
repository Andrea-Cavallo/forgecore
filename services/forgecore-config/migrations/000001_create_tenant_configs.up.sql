CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS tenant_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, key)
);

CREATE INDEX idx_tenant_configs_tenant_id ON tenant_configs(tenant_id);

ALTER TABLE tenant_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON tenant_configs
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
