-- +goose Up
-- Create quizzes table
CREATE TABLE quizzes (
    id BIGSERIAL PRIMARY KEY,
    quiz_uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(), -- uuid for external reference
    public BOOLEAN NOT NULL DEFAULT TRUE,
    difficulty difficulty NOT NULL,
    time_limit INTEGER, -- time limit in seconds, null = no limit
    access_credentials TEXT, -- password for private quizzes
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create quiz translations table
CREATE TABLE quiz_translations(
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- 'zh-CN','en-US' ...
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id, language)
);

-- Create quiz questions table (questions included in quiz)
CREATE TABLE quiz_questions (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    question_order INTEGER NOT NULL DEFAULT 0, -- order in quiz
    points INTEGER NOT NULL DEFAULT 1, -- points for this question
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(quiz_id, question_id), -- prevent duplicate questions in same quiz
    UNIQUE(quiz_id, question_order) -- prevent duplicate order in same quiz
);

-- Create quiz attempts table (user quiz participation records)
CREATE TABLE quiz_attempts (
    id BIGSERIAL PRIMARY KEY,
    attempt_uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    started_at TIMESTAMPTZ NOT NULL, -- 从 Redis 读出的服务端开始时间，随提交一起写入
    total_score INTEGER NOT NULL,
    max_score INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 提交时间

    CHECK (created_at >= started_at),
    CHECK (total_score <= max_score)
);

-- Create quiz answers table (user answers for each question in quiz)
CREATE TABLE quiz_answers (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    -- array of selected option IDs (supports multiple choice)
    selected_option_ids BIGINT[],
    -- seconds taken for this question (for future analytics)
    time_taken INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attempt_id, question_id) -- one answer per question per attempt
);

-- Create quiz statistics table
CREATE TABLE quiz_stats (
    quiz_id BIGINT PRIMARY KEY REFERENCES quizzes(id) ON DELETE CASCADE,
    attempts_count BIGINT NOT NULL DEFAULT 0, -- 总参与次数
    total_correct_answers BIGINT NOT NULL DEFAULT 0, -- 总答对题数
    highest_score INTEGER NOT NULL DEFAULT 0, -- 最高分
    shortest_time INTEGER, -- 最短完成时长（秒）
    average_score DECIMAL(5,2) NOT NULL DEFAULT 0 CHECK (average_score BETWEEN 0 AND 100), -- 平均分
    pass_rate DECIMAL(5,2) NOT NULL DEFAULT 0 CHECK (pass_rate BETWEEN 0 AND 100), -- 通过率（百分比）
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
-- quizzes table indexes
CREATE INDEX idx_quizzes_uuid ON quizzes(quiz_uuid);
CREATE INDEX idx_quizzes_public ON quizzes(public);
CREATE INDEX idx_quizzes_difficulty ON quizzes(difficulty);
CREATE INDEX idx_quizzes_created_by ON quizzes(created_by);
CREATE INDEX idx_quizzes_created_at ON quizzes(created_at);

-- quiz translations indexes
CREATE INDEX idx_quiz_translations_quiz_language ON quiz_translations(quiz_id, language);

-- Quiz questions indexes
CREATE INDEX idx_quiz_questions_quiz_id ON quiz_questions(quiz_id);
CREATE INDEX idx_quiz_questions_question_id ON quiz_questions(question_id);
CREATE INDEX idx_quiz_questions_quiz_order ON quiz_questions(quiz_id, question_order);

-- Quiz attempts indexes
CREATE INDEX idx_quiz_attempts_uuid ON quiz_attempts(attempt_uuid);
CREATE INDEX idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX idx_quiz_attempts_user_quiz ON quiz_attempts(user_id, quiz_id);
CREATE INDEX idx_quiz_attempts_score ON quiz_attempts(total_score DESC); -- for leaderboard
CREATE INDEX idx_quiz_attempts_created_at ON quiz_attempts(created_at); -- for time-based queries

-- Quiz answers indexes
CREATE INDEX idx_quiz_answers_attempt_id ON quiz_answers(attempt_id);
CREATE INDEX idx_quiz_answers_question_id ON quiz_answers(question_id);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_quiz_answers_question_id;
DROP INDEX IF EXISTS idx_quiz_answers_attempt_id;
DROP INDEX IF EXISTS idx_quiz_attempts_created_at;
DROP INDEX IF EXISTS idx_quiz_attempts_score;
DROP INDEX IF EXISTS idx_quiz_attempts_user_quiz;
DROP INDEX IF EXISTS idx_quiz_attempts_user_id;
DROP INDEX IF EXISTS idx_quiz_attempts_quiz_id;
DROP INDEX IF EXISTS idx_quiz_attempts_uuid;
DROP INDEX IF EXISTS idx_quiz_questions_quiz_order;
DROP INDEX IF EXISTS idx_quiz_questions_question_id;
DROP INDEX IF EXISTS idx_quiz_questions_quiz_id;
DROP INDEX IF EXISTS idx_quiz_translations_quiz_language;
DROP INDEX IF EXISTS idx_quizzes_created_at;
DROP INDEX IF EXISTS idx_quizzes_created_by;
DROP INDEX IF EXISTS idx_quizzes_difficulty;
DROP INDEX IF EXISTS idx_quizzes_public;
DROP INDEX IF EXISTS idx_quizzes_uuid;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS quiz_answers;
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quiz_translations;
DROP TABLE IF EXISTS quiz_stats;
DROP TABLE IF EXISTS quizzes;