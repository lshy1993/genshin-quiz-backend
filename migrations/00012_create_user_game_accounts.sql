-- +goose Up
-- Create user statistics table
CREATE TABLE user_game_accounts (
    id BIGSERIAL PRIMARY KEY,

    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- 0=GI 1=HSR 2=ZZZ 3=wuwa 4=nte
    game SMALLINT NOT NULL,      
    server SMALLINT NOT NULL,

    game_uid VARCHAR(20) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(game, server, game_uid)
);

-- +goose Down
DROP TABLE IF EXISTS user_game_accounts;