COMMENT ON TABLE oauth_providers IS 'Store OAuth provider connections and tokens for users';

COMMENT ON TABLE oauth_activity_logs IS 'Log OAuth-related activities for security monitoring';

COMMENT ON COLUMN oauth_activity_logs.action IS 'login, logout, link, unlink, refresh, revoke';

COMMENT ON COLUMN oauth_activity_logs.status IS 'success, failed';

COMMENT ON COLUMN oauth_activity_logs.ip_address IS 'Support IPv6';