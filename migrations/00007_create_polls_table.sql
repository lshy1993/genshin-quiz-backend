-- +goose Up
-- Create polls table
CREATE TABLE polls (
    id BIGSERIAL PRIMARY KEY,
    poll_uuid UUID NOT NULL DEFAULT gen_random_uuid(),

    public BOOLEAN NOT NULL DEFAULT TRUE,
    category category NOT NULL,

    start_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,

    votes_per_user INTEGER NOT NULL DEFAULT 1,
    votes_per_option INTEGER NOT NULL DEFAULT 1,

    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Cached statistics
    participants_count BIGINT NOT NULL DEFAULT 0,
    total_votes_count BIGINT NOT NULL DEFAULT 0,
    likes_count BIGINT NOT NULL DEFAULT 0
);

-- Create poll translations table
CREATE TABLE poll_translations (
    id BIGSERIAL PRIMARY KEY,

    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (poll_id, language)
);

-- Create poll options table
CREATE TABLE poll_options (
    id BIGSERIAL PRIMARY KEY,

    option_uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,

    option_order INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    vote_count BIGINT NOT NULL DEFAULT 0
);

-- Create poll option translations table
CREATE TABLE poll_option_translations (
    id BIGSERIAL PRIMARY KEY,

    option_id BIGINT NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL,
    option_text TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (option_id, language)
);

-- Create user votes table (tracks individual user votes)
CREATE TABLE user_votes (
    id BIGSERIAL PRIMARY KEY,

    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    option_id BIGINT NOT NULL REFERENCES poll_options(id) ON DELETE CASCADE,

    vote_count INTEGER NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (poll_id, user_id, option_id)
);

-- Create poll likes table
CREATE TABLE poll_likes (
    id BIGSERIAL PRIMARY KEY,

    poll_id BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    value SMALLINT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE (poll_id, user_id)
);

-- Create indexes for better performance
-- Poll indexes
CREATE UNIQUE INDEX idx_polls_uuid ON polls(poll_uuid);
CREATE INDEX idx_polls_public_category ON polls(public, category);
CREATE INDEX idx_polls_start_expires ON polls(start_at, expires_at);
CREATE INDEX idx_polls_created_by ON polls(created_by);
CREATE INDEX idx_polls_created_at ON polls(created_at DESC);

-- Poll translations
CREATE INDEX idx_poll_translations_poll_language ON poll_translations(poll_id, language);

-- Poll options
CREATE UNIQUE INDEX idx_poll_options_uuid ON poll_options(option_uuid);
CREATE INDEX idx_poll_options_poll_order ON poll_options(poll_id, option_order);

-- Poll option translations
CREATE INDEX idx_poll_option_translations_option_language ON poll_option_translations(option_id, language);

-- User votes indexes
CREATE INDEX idx_user_votes_poll ON user_votes(poll_id);
CREATE INDEX idx_user_votes_user ON user_votes(user_id);
CREATE INDEX idx_user_votes_option ON user_votes(option_id);
CREATE INDEX idx_user_votes_user_poll ON user_votes(user_id, poll_id);

-- Poll likes
CREATE INDEX idx_poll_likes_poll ON poll_likes(poll_id);
CREATE INDEX idx_poll_likes_user ON poll_likes(user_id);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_poll_likes_user;
DROP INDEX IF EXISTS idx_poll_likes_poll;

DROP INDEX IF EXISTS idx_user_votes_user_poll;
DROP INDEX IF EXISTS idx_user_votes_option;
DROP INDEX IF EXISTS idx_user_votes_user;
DROP INDEX IF EXISTS idx_user_votes_poll;

DROP INDEX IF EXISTS idx_poll_option_translations_option_language;

DROP INDEX IF EXISTS idx_poll_options_poll_order;
DROP INDEX IF EXISTS idx_poll_options_uuid;

DROP INDEX IF EXISTS idx_poll_translations_poll_language;

DROP INDEX IF EXISTS idx_polls_created_at;
DROP INDEX IF EXISTS idx_polls_created_by;
DROP INDEX IF EXISTS idx_polls_start_expires;
DROP INDEX IF EXISTS idx_polls_public_category;
DROP INDEX IF EXISTS idx_polls_uuid;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS poll_likes;
DROP TABLE IF EXISTS user_votes;
DROP TABLE IF EXISTS poll_option_translations;
DROP TABLE IF EXISTS poll_options;
DROP TABLE IF EXISTS poll_translations;
DROP TABLE IF EXISTS polls;
