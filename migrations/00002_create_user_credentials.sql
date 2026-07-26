-- +goose Up
-- Create separate table to store user password hashes and related metadata
CREATE TABLE user_credentials (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- 认证方式与标识
  identity_type VARCHAR(32) NOT NULL, -- 'password', 'google', 'github', 'phone', 'wechat'
  identifier VARCHAR(255) NOT NULL,   -- 账号标识（如 email、phone、OAuth Unique ID/OpenID）
  credential TEXT,                    -- 密码存 password_hash；第三方可存 token 或设为 NULL
  extra_data JSONB,                   -- 存放第三方返回的原生 Profile 额外信息（如 GitHub avatar/nickname）

  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT uk_identity_type_identifier UNIQUE (identity_type, identifier)
);

CREATE INDEX idx_user_credentials_user_id ON user_credentials(user_id);

CREATE TABLE user_tokens (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  -- Token 用途分类
  token_type VARCHAR(32) NOT NULL, -- 'password_reset', 'email_verification', 'magic_link'  
  -- 安全存储：强烈建议存 Token 的 SHA-256 哈希值，而不是明文 Token！
  token_hash VARCHAR(64) NOT NULL UNIQUE,

  -- 状态与有效期
  is_used BOOLEAN NOT NULL DEFAULT false,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_tokens_lookup ON user_tokens(user_id, token_type) WHERE is_used = false;

-- +goose Down
DROP INDEX IF EXISTS idx_user_tokens_lookup;
DROP TABLE IF EXISTS user_tokens;
DROP INDEX IF EXISTS idx_user_credentials_user_id;
DROP TABLE IF EXISTS user_credentials;
