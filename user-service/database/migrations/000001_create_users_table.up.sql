CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NULL,
    phone VARCHAR(20) NULL,
    photo VARCHAR(255) NULL,
    address TEXT NULL,
    lat VARCHAR(50) NULL,
    lng VARCHAR(50) NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    oauth_only BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL,
    deleted_at TIMESTAMP NULL
);

CREATE INDEX idx_users_email ON users (email);

CREATE INDEX idx_users_is_verified ON users (is_verified);