-- +goose Up
-- Create separate table to store user password hashes and related metadata
CREATE TABLE IF NOT EXISTS user_passwords (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  password_hash TEXT NOT NULL,
  password_algorithm VARCHAR(50) NOT NULL DEFAULT 'bcrypt',
  password_reset_token TEXT,
  password_reset_expires_at TIMESTAMP WITH TIME ZONE,
  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_user_passwords_user_id ON user_passwords(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_passwords_user_id;
DROP TABLE IF EXISTS user_passwords;
