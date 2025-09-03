CREATE TABLE IF NOT EXISTS oauth_activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(20) NOT NULL,
    error_msg TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_id ON oauth_activity_logs (user_id);

CREATE INDEX idx_status ON oauth_activity_logs (status);

CREATE INDEX idx_created_at ON oauth_activity_logs (created_at);

CREATE INDEX idx_provider_action ON oauth_activity_logs (provider, action);