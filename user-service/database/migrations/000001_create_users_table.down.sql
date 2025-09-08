-- Drop indexes first
DROP INDEX IF EXISTS idx_users_is_verified;

DROP INDEX IF EXISTS idx_users_email;

-- Drop table
DROP TABLE IF EXISTS users;