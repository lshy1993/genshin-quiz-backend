-- +goose Up
-- Create user_login_logs table
CREATE TABLE user_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    user_agent TEXT,                        -- 客户端设备信息（可选）
    login_type VARCHAR(32) DEFAULT 'password', -- 登录渠道（可选，如 'password', 'google'）
    status VARCHAR(20) NOT NULL DEFAULT 'SUCCESS', -- 状态：'SUCCESS', 'FAILED'（可选）
    login_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_login_logs_user_time ON user_login_logs(user_id, login_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_user_login_logs_user_time;

DROP TABLE IF EXISTS user_login_logs;