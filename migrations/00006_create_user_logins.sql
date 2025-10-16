-- +goose Up
-- Create user_login_logs table
CREATE TABLE user_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL,
    login_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_login_logs_user_id ON user_login_logs(user_id);
CREATE INDEX idx_user_login_logs_login_at ON user_login_logs(login_at);

-- +goose Down
DROP INDEX IF EXISTS idx_user_login_logs_login_at;
DROP INDEX IF EXISTS idx_user_login_logs_user_id;

DROP TABLE user_login_logs;