-- +goose Up
-- 跨站单点登录的一次性授权码（authorization code + PKCE）：只保存哈希，60 秒内有效，只能兑换一次。
CREATE TABLE sso_auth_codes (
    id BIGSERIAL PRIMARY KEY,
    code_hash BYTEA NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client_id TEXT NOT NULL,
    redirect_uri TEXT NOT NULL,
    code_challenge TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

-- +goose Down
DROP TABLE IF EXISTS sso_auth_codes;
