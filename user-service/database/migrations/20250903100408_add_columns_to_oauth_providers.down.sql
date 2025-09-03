DROP INDEX IF EXISTS idx_oauth_providers_expired;

DROP INDEX IF EXISTS idx_oauth_providers_user_revoked;

DROP INDEX IF EXISTS idx_user_id_provider;

DROP INDEX IF EXISTS idx_token_expires_at;

DROP INDEX IF EXISTS idx_is_revoked;

ALTER TABLE oauth_providers DROP COLUMN IF EXISTS is_revoked;