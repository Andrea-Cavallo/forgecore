-- 000003_add_oauth_fields.up.sql
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS oauth_provider    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS oauth_provider_id TEXT NOT NULL DEFAULT '';

-- Unique constraint: one OAuth identity per tenant per provider
CREATE UNIQUE INDEX IF NOT EXISTS users_oauth_unique
    ON users(tenant_id, oauth_provider, oauth_provider_id)
    WHERE oauth_provider <> '';
