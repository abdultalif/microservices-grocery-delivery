ALTER TABLE oauth_providers
ADD CONSTRAINT chk_provider_valid CHECK (
    provider IN (
        'google',
        'facebook',
        'twitter',
        'github',
        'apple'
    )
);

ALTER TABLE oauth_activity_logs
ADD CONSTRAINT chk_action_valid CHECK (
    action IN (
        'login',
        'logout',
        'link',
        'unlink',
        'refresh',
        'revoke'
    )
);

ALTER TABLE oauth_activity_logs
ADD CONSTRAINT chk_status_valid CHECK (
    status IN ('success', 'failed')
);