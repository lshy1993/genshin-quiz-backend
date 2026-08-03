-- +goose Up
-- Create user profile table
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    -- Personal information
    gender SMALLINT NOT NULL DEFAULT 0,
    country CHAR(2),
    timezone VARCHAR(50),
    birthday DATE,

    -- Social links
    website TEXT,
    twitter TEXT,
    discord TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_profiles.gender IS 'Gender: 0=Unknown, 1=Male, 2=Female, 3=Other';
COMMENT ON COLUMN user_profiles.country IS 'ISO 3166-1 alpha-2 country code';

-- Create indexes
CREATE INDEX idx_user_profiles_country ON user_profiles(country);

-- +goose Down
DROP INDEX IF EXISTS idx_user_profiles_country;

DROP TABLE IF EXISTS user_profiles;
