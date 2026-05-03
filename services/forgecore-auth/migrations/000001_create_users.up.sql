CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL,
    email_enc      BYTEA NOT NULL,
    email_hash     VARCHAR(64) NOT NULL,
    password_hash  VARCHAR(255),
    roles          TEXT[] NOT NULL DEFAULT '{"user"}',
    mfa_enabled    BOOLEAN NOT NULL DEFAULT false,
    mfa_secret     BYTEA,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    locked_until   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ,
    UNIQUE(tenant_id, email_hash)
);
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.tenant_id')::uuid);

CREATE TABLE sessions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id    VARCHAR(255) NOT NULL,
    user_agent   TEXT,
    ip_address   INET,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_sessions_user_id ON sessions(tenant_id, user_id);
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON sessions
    USING (tenant_id = current_setting('app.tenant_id')::uuid);
