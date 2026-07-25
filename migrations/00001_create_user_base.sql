-- +goose Up
-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY, -- internal identifier
    user_uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), -- public identifier

    email VARCHAR(255) UNIQUE NOT NULL,

    -- 基础账号信息
    nickname VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    biography TEXT,

    -- 用户偏好语言
    language VARCHAR(10) NOT NULL DEFAULT 'zh-CN',

    -- 用户创建信息
    user_role SMALLINT NOT NULL DEFAULT 0, -- 0 user 1 admin
    email_verified BOOLEAN NOT NULL default false,

    -- 用户状态
    status SMALLINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    
    created_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN users.user_role IS '0=regular user, 1=admin, 2=moderator';
COMMENT ON COLUMN users.status IS '0=active, 1=suspended, 2=deleted';
COMMENT ON COLUMN users.language IS 'IETF BCP 47 language tag';

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;