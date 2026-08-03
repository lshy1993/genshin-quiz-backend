-- +goose Up
-- Create user_login_logs table
CREATE TABLE user_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,

    -- 客户端设备信息
    user_agent TEXT,

    -- 登录渠道
    credential_type SMALLINT NOT NULL DEFAULT 0,

    -- 状态：'SUCCESS', 'FAILED'
    status SMALLINT NOT NULL DEFAULT 0,

    login_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_login_logs.credential_type IS 'Login provider: 0=password, 1=google, 2=apple, 3=github';
COMMENT ON COLUMN user_login_logs.status IS 'Login result: 0=success, 1=failed, 2=blocked';

-- create index
CREATE INDEX idx_user_login_logs_user_time ON user_login_logs(user_id, login_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_user_login_logs_user_time;

DROP TABLE IF EXISTS user_login_logs;