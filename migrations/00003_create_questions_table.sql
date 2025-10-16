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
CREATE TABLE IF NOT EXISTS questions (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE, -- uuid 
    public BOOLEAN NOT NULL DEFAULT TRUE, -- is this question public
    question_type question_type NOT NULL, -- multiple choice, true/false, etc.
    category category NOT NULL,
    difficulty difficulty NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    published_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
      -- 快速读取用的冗余统计（可选）
    answer_count BIGINT NOT NULL DEFAULT 0,
    correct_count BIGINT NOT NULL DEFAULT 0,
    likes BIGINT NOT NULL DEFAULT 0
);

-- Create question options table
CREATE TABLE IF NOT EXISTS question_options (
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    option_type question_option_type NOT NULL DEFAULT 'text',
    option_text TEXT NOT NULL,
    img_url TEXT, -- optional image URL
    is_answer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE question_translations(
    id BIGSERIAL PRIMARY KEY,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- 'zh-CN','en-US' ...
    question_text VARCHAR(500) NOT NULL, -- text
    description TEXT, -- detailed explanation or context
    explanation TEXT, -- explanation for the correct answer
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(question_id, language)
)

-- Create question submissions table to track user question solve attempts
CREATE TABLE question_submissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    is_practice BOOLEAN NOT NULL DEFAULT FALSE, -- true if this was a practice attempt
    time_taken INTEGER, -- seconds
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX idx_quizzes_category ON quizzes(category);
CREATE INDEX idx_quizzes_difficulty ON quizzes(difficulty);
CREATE INDEX idx_quizzes_created_by ON quizzes(created_by);
CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX idx_questions_order ON questions(quiz_id, order_index);
CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX idx_user_answers_question_id ON user_answers(question_id);

-- +goose Down
DROP TRIGGER IF EXISTS update_questions_updated_at ON questions;
DROP TRIGGER IF EXISTS update_quizzes_updated_at ON quizzes;

DROP INDEX IF EXISTS idx_user_answers_question_id;
DROP INDEX IF EXISTS idx_user_answers_attempt_id;
DROP INDEX IF EXISTS idx_quiz_attempts_quiz_id;
DROP INDEX IF EXISTS idx_quiz_attempts_user_id;
DROP INDEX IF EXISTS idx_questions_order;
DROP INDEX IF EXISTS idx_questions_quiz_id;
DROP INDEX IF EXISTS idx_quizzes_created_by;
DROP INDEX IF EXISTS idx_quizzes_difficulty;
DROP INDEX IF EXISTS idx_quizzes_category;

DROP TABLE user_answers;
DROP TABLE quiz_attempts;
DROP TABLE questions;
DROP TABLE quizzes;

DROP TYPE question_type;
DROP TYPE quiz_difficulty;
DROP TYPE quiz_category;