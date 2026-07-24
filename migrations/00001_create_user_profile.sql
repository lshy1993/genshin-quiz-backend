-- +goose Up
-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY, -- internal identifier
    user_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- public identifier

    -- 基础展示与账号信息
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    avatar_url TEXT,
    biography TEXT,

    -- 个人偏好与设置
    gender VARCHAR(20),
    country VARCHAR(2),
    timezone VARCHAR(50),
    language VARCHAR(10) DEFAULT 'zh-CN',
    show_email BOOLEAN NOT NULL DEFAULT false,

    created_ip INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_uuid ON users(user_uuid);

-- +goose Down
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_uuid;
DROP TABLE IF EXISTS users;