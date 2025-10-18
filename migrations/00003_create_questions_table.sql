-- +goose Up
-- Create quiz categories enum
CREATE TYPE category AS ENUM ( 'character', 'weapon', 'artifact', 'lore', 'gameplay', 'world', 'combat', 'music', 'statistics', 'fun', 'real', 'other');

-- Create quiz difficulty enum
CREATE TYPE difficulty AS ENUM ('easy', 'medium', 'hard');

-- Create question types enum
CREATE TYPE question_type AS ENUM ('multiple_choice', 'single_choice', 'true_false');

-- Create question option types enum
CREATE TYPE question_option_type AS ENUM ('text', 'image', 'audio');

-- Create questions table
CREATE TABLE questions (
    id BIGSERIAL PRIMARY KEY, -- inner id
    question_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- uuid, expose
    public BOOLEAN NOT NULL DEFAULT TRUE, -- is this question public
    question_type question_type NOT NULL, -- multiple choice, true/false, etc.
    category category NOT NULL,
    difficulty difficulty NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMPTZ,
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 快速读取用的冗余统计（可选）
    submit_count BIGINT NOT NULL DEFAULT 0,
    correct_count BIGINT NOT NULL DEFAULT 0,
    likes BIGINT NOT NULL DEFAULT 0
);

-- Create question options table
CREATE TABLE question_options (
    id BIGSERIAL PRIMARY KEY,
    option_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- uuid, expose
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_type question_option_type NOT NULL DEFAULT 'text',
    img_url TEXT, -- optional image URL
    is_answer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        -- 快速读取用的冗余统计（可选）
    selected_count BIGINT NOT NULL DEFAULT 0
);

-- Create question translations table
CREATE TABLE question_translations(
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- 'zh-CN','en-US' ...
    question_text TEXT NOT NULL, -- text
    description TEXT, -- detailed explanation or context
    explanation TEXT, -- explanation for the correct answer
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, language)
);

-- Create question option translations table
CREATE TABLE option_translations (
    id BIGSERIAL PRIMARY KEY,
    option_id BIGINT NOT NULL REFERENCES question_options(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,
    option_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(option_id, language)
);

-- Create question submissions table to track user question solve attempts
CREATE TABLE question_submissions (
    id BIGSERIAL PRIMARY KEY,
    submission_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- uuid, expose
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    is_practice BOOLEAN NOT NULL DEFAULT FALSE, -- practice attempt won't affect stats
    time_taken INTEGER, -- seconds
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create question likes table to track user likes/dislikes on questions
CREATE TABLE question_likes (
  question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  value SMALLINT NOT NULL, -- 1 = like, -1 = dislike
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (question_id, user_id)
);

-- Create comments table for questions
CREATE TABLE question_comments (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
-- 索引：questions（注意表中字段名为 question_id，按现有定义建立索引）
CREATE INDEX idx_questions_id ON questions(question_uuid);
CREATE INDEX idx_questions_public_published_on_date ON questions(public, is_published, published_at);
CREATE INDEX idx_questions_public_published_likes ON questions (public, is_published, likes DESC);
CREATE INDEX idx_questions_category_difficulty ON questions(category, difficulty);

-- 索引：options（按 question 快速查、按是否为正确项过滤）
CREATE INDEX idx_question_options_question_id ON question_options(question_id);
CREATE INDEX idx_question_options_question_is_answer ON question_options(question_id, is_answered);

-- 索引：翻译表按语言和 question/option 快速查
CREATE INDEX idx_question_translations_question_language ON question_translations(question_id, language);
CREATE INDEX idx_option_translations_option_language ON option_translations(option_id, language);

-- 索引：提交（按 user、按 question、按时间范围查询常用）
CREATE INDEX idx_question_submissions_user_id ON question_submissions(user_id);
CREATE INDEX idx_question_submissions_question_id ON question_submissions(question_id);
CREATE INDEX idx_question_submissions_question_created_at ON question_submissions(question_id, created_at);

-- 索引：likes / comments（按 question 快速聚合，按 user 用于查用户行为）
CREATE INDEX idx_question_likes_question_id ON question_likes(question_id);
CREATE INDEX idx_question_likes_question_user ON question_likes(question_id, user_id);
CREATE INDEX idx_question_comments_question_id ON question_comments(question_id);
CREATE INDEX idx_question_comments_user_id ON question_comments(user_id);


-- +goose Down
DROP INDEX IF EXISTS idx_question_comments_user_id;
DROP INDEX IF EXISTS idx_question_comments_question_id;
DROP INDEX IF EXISTS idx_question_likes_question_user;
DROP INDEX IF EXISTS idx_question_likes_question_id;
DROP INDEX IF EXISTS idx_question_submissions_question_created_at;
DROP INDEX IF EXISTS idx_question_submissions_question_id;
DROP INDEX IF EXISTS idx_question_submissions_user_id;
DROP INDEX IF EXISTS idx_option_translations_option_language;
DROP INDEX IF EXISTS idx_question_translations_question_language;
DROP INDEX IF EXISTS idx_question_options_question_is_answer;
DROP INDEX IF EXISTS idx_question_options_question_id;
DROP INDEX IF EXISTS idx_questions_category_difficulty;
DROP INDEX IF EXISTS idx_questions_public_published_on_date;
DROP INDEX IF EXISTS idx_questions_quiz_id;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS question_comments;
DROP TABLE IF EXISTS question_likes;
DROP TABLE IF EXISTS question_submissions;
DROP TABLE IF EXISTS option_translations;
DROP TABLE IF EXISTS question_translations;
DROP TABLE IF EXISTS question_options;
DROP TABLE IF EXISTS questions;

-- Drop enum types
DROP TYPE IF EXISTS question_option_type;
DROP TYPE IF EXISTS question_type;
DROP TYPE IF EXISTS difficulty;
DROP TYPE IF EXISTS category;
