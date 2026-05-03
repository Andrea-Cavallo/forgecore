-- 000003_add_oauth_fields.down.sql
DROP INDEX IF EXISTS users_oauth_unique;
ALTER TABLE users
    DROP COLUMN IF EXISTS oauth_provider,
    DROP COLUMN IF EXISTS oauth_provider_id;
