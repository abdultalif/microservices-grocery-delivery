CREATE TABLE IF NOT EXISTS oauth_providers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL, -- 'google', 'facebook', etc
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255) NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_picture VARCHAR(500) NULL,
    access_token TEXT NULL,
    refresh_token TEXT NULL,
    token_expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL,
    UNIQUE (provider, provider_user_id),
    UNIQUE (provider, provider_email)
);

CREATE INDEX idx_oauth_providers_user_id ON oauth_providers (user_id);

CREATE INDEX idx_oauth_providers_provider ON oauth_providers (provider);

CREATE INDEX idx_oauth_providers_email ON oauth_providers (provider_email);