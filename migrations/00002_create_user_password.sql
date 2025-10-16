-- +goose Up
-- Create separate table to store user password hashes and related metadata
CREATE TABLE user_passwords (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
  password_hash TEXT NOT NULL,
  password_reset_token TEXT,
  password_reset_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_passwords_user_id ON user_passwords(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_passwords_user_id;
DROP TABLE IF EXISTS user_passwords;
