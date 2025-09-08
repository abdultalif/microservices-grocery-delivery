CREATE TABLE IF NOT EXISTS oauth_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'facebook', 'github', 'apple', etc
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_picture VARCHAR(500) NULL,
    access_token TEXT NULL,
    refresh_token TEXT NULL,
    token_expires_at TIMESTAMP NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL,

-- Constraints
UNIQUE (provider, provider_user_id),
UNIQUE (provider, provider_email),

-- Add constraint for supported providers
CONSTRAINT chk_provider_valid CHECK (provider IN ('google', 'facebook', 'twitter', 'github', 'apple'))
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_oauth_providers_user_id ON oauth_providers (user_id);

CREATE INDEX IF NOT EXISTS idx_oauth_providers_provider ON oauth_providers (provider);

CREATE INDEX IF NOT EXISTS idx_oauth_providers_email ON oauth_providers (provider_email);

-- Additional indexes for production features
CREATE INDEX IF NOT EXISTS idx_oauth_providers_is_revoked ON oauth_providers (is_revoked);

CREATE INDEX IF NOT EXISTS idx_oauth_providers_token_expires ON oauth_providers (token_expires_at)
WHERE
    token_expires_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_oauth_providers_user_revoked ON oauth_providers (user_id, is_revoked);

CREATE INDEX IF NOT EXISTS idx_oauth_providers_expired_tokens ON oauth_providers (
    token_expires_at,
    refresh_token
)
WHERE
    token_expires_at IS NOT NULL
    AND refresh_token IS NOT NULL;