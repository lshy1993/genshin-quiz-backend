-- +goose Up
-- Create votes table
CREATE TABLE votes (
    id BIGSERIAL PRIMARY KEY,
    vote_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- uuid for external reference
    public BOOLEAN NOT NULL DEFAULT TRUE,
    category category NOT NULL, -- reuse existing category enum
    start_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ, -- nullable, no expiration if null
    votes_per_user INTEGER NOT NULL DEFAULT 1, -- max votes per user
    votes_per_option INTEGER DEFAULT 1, -- 0 = unlimited
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 冗余统计字段（用于快速查询）
    participants_count BIGINT NOT NULL DEFAULT 0,
    total_votes_count BIGINT NOT NULL DEFAULT 0,
    likes_count BIGINT NOT NULL DEFAULT 0
);

-- Create vote translations table
CREATE TABLE vote_translations(
    id BIGSERIAL PRIMARY KEY,
    vote_id BIGINT NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- 'zh-CN','en-US' ...
    title TEXT NOT NULL, -- title
    description TEXT, -- detailed explanation or context
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vote_id, language)
);

-- Create vote options table
CREATE TABLE vote_options (
    id BIGSERIAL PRIMARY KEY,
    option_uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    vote_id BIGINT NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    option_text TEXT NOT NULL,
    option_order INTEGER NOT NULL DEFAULT 0, -- display order
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 冗余统计字段
    vote_count BIGINT NOT NULL DEFAULT 0
);

-- Create user votes table (tracks individual user votes)
CREATE TABLE user_votes (
    id BIGSERIAL PRIMARY KEY,
    vote_id BIGINT NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_id BIGINT NOT NULL REFERENCES vote_options(id) ON DELETE CASCADE,
    vote_count INTEGER NOT NULL DEFAULT 1, -- number of votes for this option
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vote_id, user_id, option_id) -- one record per user-option combination
);

-- Create vote likes table
CREATE TABLE vote_likes (
    id BIGSERIAL PRIMARY KEY,
    vote_id BIGINT NOT NULL REFERENCES votes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value SMALLINT NOT NULL, -- 1 = like, -1 = dislike
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(vote_id, user_id)
);

-- Create indexes for better performance
-- Votes table indexes
CREATE INDEX idx_votes_uuid ON votes(vote_uuid);
CREATE INDEX idx_votes_public_category ON votes(public, category);
CREATE INDEX idx_votes_start_expires ON votes(start_at, expires_at);
CREATE INDEX idx_votes_created_by ON votes(created_by);
CREATE INDEX idx_votes_created_at ON votes(created_at);

-- Vote options indexes
CREATE INDEX idx_vote_options_uuid ON vote_options(option_uuid);
CREATE INDEX idx_vote_options_vote_id ON vote_options(vote_id);
CREATE INDEX idx_vote_options_vote_order ON vote_options(vote_id, option_order);

-- User votes indexes
CREATE INDEX idx_user_votes_vote_id ON user_votes(vote_id);
CREATE INDEX idx_user_votes_user_id ON user_votes(user_id);
CREATE INDEX idx_user_votes_option_id ON user_votes(option_id);
CREATE INDEX idx_user_votes_user_vote ON user_votes(user_id, vote_id);

-- Vote likes indexes
CREATE INDEX idx_vote_likes_vote_id ON vote_likes(vote_id);
CREATE INDEX idx_vote_likes_user_id ON vote_likes(user_id);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_vote_likes_user_id;
DROP INDEX IF EXISTS idx_vote_likes_vote_id;
DROP INDEX IF EXISTS idx_user_votes_user_vote;
DROP INDEX IF EXISTS idx_user_votes_option_id;
DROP INDEX IF EXISTS idx_user_votes_user_id;
DROP INDEX IF EXISTS idx_user_votes_vote_id;
DROP INDEX IF EXISTS idx_vote_options_vote_order;
DROP INDEX IF EXISTS idx_vote_options_vote_id;
DROP INDEX IF EXISTS idx_vote_options_uuid;
DROP INDEX IF EXISTS idx_votes_created_at;
DROP INDEX IF EXISTS idx_votes_created_by;
DROP INDEX IF EXISTS idx_votes_start_expires;
DROP INDEX IF EXISTS idx_votes_public_category;
DROP INDEX IF EXISTS idx_votes_uuid;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS vote_likes;
DROP TABLE IF EXISTS user_votes;
DROP TABLE IF EXISTS vote_options;
DROP TABLE IF EXISTS votes;