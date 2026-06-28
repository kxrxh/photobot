ALTER TABLE users ADD COLUMN IF NOT EXISTS login VARCHAR(64) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_messenger_id_check;
ALTER TABLE users ADD CONSTRAINT users_identity_check CHECK (
    telegram_id IS NOT NULL OR max_id IS NOT NULL OR login IS NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_login ON users(login);

CREATE TABLE IF NOT EXISTS user_recovery_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_recovery_codes_user_id ON user_recovery_codes(user_id);
