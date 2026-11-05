-- +goose Up
-- 登录会话：浏览器只持有不透明的 HttpOnly Cookie，数据库只保存其 SHA-256，
-- 可随时撤销（登出 / 全部登出），并用于换取短期 access token。
CREATE TABLE user_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash BYTEA NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    ip TEXT,
    user_agent TEXT
);

CREATE INDEX idx_user_sessions_user_active ON user_sessions(user_id) WHERE revoked_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_user_sessions_user_active;
DROP TABLE IF EXISTS user_sessions;
