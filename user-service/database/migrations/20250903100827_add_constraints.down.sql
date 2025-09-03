ALTER TABLE oauth_activity_logs
DROP CONSTRAINT IF EXISTS chk_status_valid;

ALTER TABLE oauth_activity_logs
DROP CONSTRAINT IF EXISTS chk_action_valid;

ALTER TABLE oauth_providers
DROP CONSTRAINT IF EXISTS chk_provider_valid;