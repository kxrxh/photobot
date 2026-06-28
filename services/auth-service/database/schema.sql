CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE,
    max_id BIGINT UNIQUE,
    login VARCHAR(64) UNIQUE,
    password_hash VARCHAR(255),
    organization_name VARCHAR(128),
    inn VARCHAR(12),
    full_name VARCHAR(128),
    phone_number VARCHAR(32) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_identity_check CHECK (
        telegram_id IS NOT NULL OR max_id IS NOT NULL OR login IS NOT NULL
    )
);

-- Indexes for users table
CREATE INDEX idx_users_telegram_id ON users(telegram_id);
CREATE INDEX idx_users_login ON users(login);
CREATE INDEX idx_users_inn ON users(inn);
CREATE INDEX idx_users_phone_number ON users(phone_number);
CREATE INDEX idx_users_organization_name ON users(organization_name);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_users_full_name_gin ON users USING gin(to_tsvector('russian', full_name)); -- Full-text search

-- Roles table
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) UNIQUE NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for roles table
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_created_at ON roles(created_at);

-- User roles junction table
CREATE TABLE user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- Indexes for user_roles table
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

-- Bots table
CREATE TABLE bots (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL CHECK (LENGTH(TRIM(name)) > 0),
    token BYTEA NOT NULL, -- Encrypted token
    platform VARCHAR(32) NOT NULL DEFAULT 'telegram',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT bots_name_platform_key UNIQUE (name, platform)
);

-- Indexes for bots table
CREATE INDEX idx_bots_name ON bots(name);
CREATE INDEX idx_bots_created_at ON bots(created_at);

-- Improved function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    -- Only update if the row actually changed
    IF OLD IS DISTINCT FROM NEW THEN
        NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_bots_updated_at
    BEFORE UPDATE ON bots
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Service table for OAuth2 Client Credentials flow
CREATE TABLE services (
    service_id VARCHAR(128) UNIQUE NOT NULL CHECK (LENGTH(TRIM(service_id)) > 0),
    service_secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for services table
CREATE INDEX idx_services_service_id ON services(service_id);
CREATE INDEX idx_services_created_at ON services(created_at);

-- Trigger for services updated_at
CREATE TRIGGER update_services_updated_at
    BEFORE UPDATE ON services
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Web account recovery codes (one-time use)
CREATE TABLE user_recovery_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_recovery_codes_user_id ON user_recovery_codes(user_id);
