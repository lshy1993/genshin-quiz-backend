-- +goose Up
-- Create vote comments table for future use
CREATE TABLE vote_comments (
    id BIGSERIAL PRIMARY KEY,
    vote_id BIGINT NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for vote comments
CREATE INDEX idx_vote_comments_vote_id ON vote_comments(vote_id);
CREATE INDEX idx_vote_comments_user_id ON vote_comments(user_id);
CREATE INDEX idx_vote_comments_vote_created_at ON vote_comments(vote_id, created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_vote_comments_vote_created_at;
DROP INDEX IF EXISTS idx_vote_comments_user_id;
DROP INDEX IF EXISTS idx_vote_comments_vote_id;
DROP TABLE IF EXISTS vote_comments;
