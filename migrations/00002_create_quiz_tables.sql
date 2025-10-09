-- +goose Up
-- Create quiz categories enum
CREATE TYPE quiz_category AS ENUM ('characters', 'weapons', 'artifacts', 'lore', 'gameplay');

-- Create quiz difficulty enum
CREATE TYPE quiz_difficulty AS ENUM ('easy', 'medium', 'hard');

-- Create question types enum
CREATE TYPE question_type AS ENUM ('multiple_choice', 'true_false', 'fill_in_blank');

-- Create quizzes table
CREATE TABLE quizzes (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    category quiz_category NOT NULL,
    difficulty quiz_difficulty NOT NULL,
    time_limit INTEGER DEFAULT 300, -- seconds
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create questions table
CREATE TABLE questions (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text VARCHAR(500) NOT NULL,
    question_type question_type NOT NULL,
    options TEXT[], -- JSON array for multiple choice options
    correct_answer TEXT NOT NULL,
    explanation TEXT,
    points INTEGER NOT NULL DEFAULT 10,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id, order_index)
);

-- Create quiz attempts table to track user quiz attempts
CREATE TABLE quiz_attempts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    score INTEGER NOT NULL DEFAULT 0,
    max_score INTEGER NOT NULL,
    time_taken INTEGER, -- seconds
    completed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create user answers table to track individual question answers
CREATE TABLE user_answers (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_answer TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    points_earned INTEGER NOT NULL DEFAULT 0,
    answered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attempt_id, question_id)
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