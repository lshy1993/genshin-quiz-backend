-- +goose Up
-- Add user statistics fields to users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_submissions BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS correct_submissions BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS questions_created BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_votes BIGINT NOT NULL DEFAULT 0;

-- Create indexes for statistics queries
CREATE INDEX IF NOT EXISTS idx_users_total_submissions ON users(total_submissions);
CREATE INDEX IF NOT EXISTS idx_users_correct_submissions ON users(correct_submissions);

-- +goose Down
DROP INDEX IF EXISTS idx_users_correct_submissions;
DROP INDEX IF EXISTS idx_users_total_submissions;

ALTER TABLE users DROP COLUMN IF EXISTS total_votes;
ALTER TABLE users DROP COLUMN IF EXISTS questions_created;
ALTER TABLE users DROP COLUMN IF EXISTS correct_submissions;
ALTER TABLE users DROP COLUMN IF EXISTS total_submissions;