CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    filename TEXT NOT NULL,
    bucket TEXT NOT NULL,
    key TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_files_tenant_id ON files(tenant_id);
CREATE INDEX idx_files_user ON files(tenant_id, user_id);
CREATE UNIQUE INDEX idx_files_bucket_key ON files(tenant_id, bucket, key);

ALTER TABLE files ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON files
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
