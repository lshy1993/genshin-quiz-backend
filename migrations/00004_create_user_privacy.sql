-- +goose Up
-- Create user privacies table
CREATE TABLE user_privacies (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

    email_visibility SMALLINT NOT NULL DEFAULT 0,
    birthday_visibility SMALLINT NOT NULL DEFAULT 0,
    gender_visibility SMALLINT NOT NULL DEFAULT 0,
    country_visibility SMALLINT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON COLUMN user_privacies.email_visibility IS 'Visibility: 0=private, 1=public, 2=friends';
COMMENT ON COLUMN user_privacies.birthday_visibility IS 'Visibility: 0=private, 1=public, 2=friends';
COMMENT ON COLUMN user_privacies.gender_visibility IS 'Visibility: 0=private, 1=public, 2=friends';
COMMENT ON COLUMN user_privacies.country_visibility IS 'Visibility: 0=private, 1=public, 2=friends';

-- +goose Down
DROP TABLE IF EXISTS user_privacies;