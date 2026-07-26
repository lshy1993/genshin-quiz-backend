-- +goose Up
-- Create poll comments table for future use
CREATE TABLE poll_comments (
    id BIGSERIAL PRIMARY KEY,
    comment_uuid UUID NOT NULL DEFAULT gen_random_uuid(),

    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,

    parent_id BIGINT REFERENCES poll_comments(id) ON DELETE CASCADE,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for poll comments
CREATE UNIQUE INDEX idx_poll_comments_uuid ON poll_comments(comment_uuid);
CREATE INDEX idx_poll_comments_poll_created_at ON poll_comments(poll_id, created_at);
CREATE INDEX idx_poll_comments_user_id ON poll_comments(user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_poll_comments_user_id;
DROP INDEX IF EXISTS idx_poll_comments_poll_created_at;
DROP INDEX IF EXISTS idx_poll_comments_uuid;
DROP TABLE IF EXISTS poll_comments;
