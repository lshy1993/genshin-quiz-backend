-- +goose Up
-- Create user statistics table
CREATE TABLE user_stats (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- 你的回答
    total_submissions BIGINT NOT NULL DEFAULT 0,
    correct_submissions BIGINT NOT NULL DEFAULT 0,

    -- 你创建的
    questions_created BIGINT NOT NULL DEFAULT 0,

    -- 你投出的票
    votes_cast BIGINT NOT NULL DEFAULT 0,

    -- 你创建的投票
    polls_created BIGINT NOT NULL DEFAULT 0,

    -- 你创建的收到的赞
    likes_received BIGINT NOT NULL DEFAULT 0,

    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for statistics queries
CREATE INDEX idx_user_stats_total_submissions ON user_stats(total_submissions DESC);

CREATE INDEX idx_user_stats_questions_created ON user_stats(questions_created DESC);

CREATE INDEX idx_user_stats_polls_created ON user_stats(polls_created DESC);

CREATE INDEX idx_user_stats_likes_received ON user_stats(likes_received DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_user_stats_likes_received;
DROP INDEX IF EXISTS idx_user_stats_polls_created;
DROP INDEX IF EXISTS idx_user_stats_questions_created;
DROP INDEX IF EXISTS idx_user_stats_total_submissions;

DROP TABLE IF EXISTS user_stats;