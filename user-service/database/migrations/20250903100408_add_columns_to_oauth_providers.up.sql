ALTER TABLE oauth_providers
ADD COLUMN is_revoked BOOLEAN DEFAULT FALSE;

CREATE INDEX idx_is_revoked ON oauth_providers (is_revoked);

CREATE INDEX idx_token_expires_at ON oauth_providers (token_expires_at);

CREATE INDEX idx_user_id_provider ON oauth_providers (user_id, provider);

CREATE INDEX idx_oauth_providers_user_revoked ON oauth_providers (user_id, is_revoked);

CREATE INDEX idx_oauth_providers_expired ON oauth_providers (token_expires_at)
WHERE
    token_expires_at IS NOT NULL;