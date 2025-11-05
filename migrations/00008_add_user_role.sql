-- Add user role column to users table
-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN user_role INTEGER DEFAULT 0;

-- 添加注释说明角色定义
COMMENT ON COLUMN users.user_role IS 'User role: 0=regular user, 1=admin, 2=moderator';

-- 可选：为现有用户设置默认角色
UPDATE users SET user_role = 0 WHERE user_role IS NULL;

-- 添加索引以提高查询性能
CREATE INDEX idx_users_user_role ON users(user_role);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_user_role;
ALTER TABLE users DROP COLUMN user_role;
-- +goose StatementEnd