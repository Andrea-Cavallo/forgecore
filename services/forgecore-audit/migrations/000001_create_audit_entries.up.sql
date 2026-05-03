CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS audit_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    actor_id UUID,
    actor_type TEXT NOT NULL,
    action TEXT NOT NULL,
    resource_id UUID,
    resource_type TEXT NOT NULL DEFAULT '',
    ip_address INET,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_entries_tenant_id ON audit_entries(tenant_id);
CREATE INDEX idx_audit_entries_tenant_time ON audit_entries(tenant_id, occurred_at DESC, id DESC);
CREATE INDEX idx_audit_entries_actor ON audit_entries(tenant_id, actor_id);
CREATE INDEX idx_audit_entries_action ON audit_entries(tenant_id, action);

ALTER TABLE audit_entries ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON audit_entries
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
