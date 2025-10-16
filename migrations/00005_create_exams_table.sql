-- +goose Up
-- Create exams table
CREATE TABLE exams (
    id BIGSERIAL PRIMARY KEY,
    exam_uuid UUID NOT NULL DEFAULT gen_random_uuid(), -- uuid for external reference
    public BOOLEAN NOT NULL DEFAULT TRUE,
    difficulty difficulty NOT NULL,
    time_limit INTEGER, -- time limit in seconds, null = no limit
    access_password TEXT, -- password for private exams
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- 冗余统计字段
    attempts_count BIGINT NOT NULL DEFAULT 0, -- 总参与次数
    total_correct_answers BIGINT NOT NULL DEFAULT 0, -- 总答对题数
    highest_score INTEGER NOT NULL DEFAULT 0, -- 最高分
    shortest_time INTEGER, -- 最短完成时长（秒）
    average_score DECIMAL(5,2) NOT NULL DEFAULT 0, -- 平均分
    pass_rate DECIMAL(5,2) NOT NULL DEFAULT 0 -- 通过率（百分比）
);

-- Create exam translations table
CREATE TABLE exam_translations(
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL, -- 'zh-CN','en-US' ...
    title TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(exam_id, language)
);

-- Create exam questions table (questions included in exam)
CREATE TABLE exam_questions (
    id BIGSERIAL PRIMARY KEY,
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    question_order INTEGER NOT NULL DEFAULT 0, -- order in exam
    points INTEGER NOT NULL DEFAULT 1, -- points for this question
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(exam_id, question_id), -- prevent duplicate questions in same exam
    UNIQUE(exam_id, question_order) -- prevent duplicate order in same exam
);

-- Create exam attempts table (user exam participation records)
CREATE TABLE exam_attempts (
    id BIGSERIAL PRIMARY KEY,
    attempt_uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    exam_id BIGINT NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    time_taken INTEGER NOT NULL, -- seconds taken to complete (from frontend)
    total_score INTEGER NOT NULL DEFAULT 0,
    max_score INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP -- completion time
);

-- Create exam answers table (user answers for each question in exam)
CREATE TABLE exam_answers (
    id BIGSERIAL PRIMARY KEY,
    attempt_id BIGINT NOT NULL REFERENCES exam_attempts(id) ON DELETE CASCADE,
    question_id BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    selected_option_ids BIGINT[], -- array of selected option IDs (supports multiple choice)
    time_taken INTEGER, -- seconds taken for this question (for future analytics)
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(attempt_id, question_id) -- one answer per question per attempt
);

-- Create indexes for better performance
-- Exams table indexes
CREATE INDEX idx_exams_uuid ON exams(exam_uuid);
CREATE INDEX idx_exams_public ON exams(public);
CREATE INDEX idx_exams_difficulty ON exams(difficulty);
CREATE INDEX idx_exams_created_by ON exams(created_by);
CREATE INDEX idx_exams_created_at ON exams(created_at);

-- Exam translations indexes
CREATE INDEX idx_exam_translations_exam_language ON exam_translations(exam_id, language);

-- Exam questions indexes
CREATE INDEX idx_exam_questions_exam_id ON exam_questions(exam_id);
CREATE INDEX idx_exam_questions_question_id ON exam_questions(question_id);
CREATE INDEX idx_exam_questions_exam_order ON exam_questions(exam_id, question_order);

-- Exam attempts indexes
CREATE INDEX idx_exam_attempts_uuid ON exam_attempts(attempt_uuid);
CREATE INDEX idx_exam_attempts_exam_id ON exam_attempts(exam_id);
CREATE INDEX idx_exam_attempts_user_id ON exam_attempts(user_id);
CREATE INDEX idx_exam_attempts_user_exam ON exam_attempts(user_id, exam_id);
CREATE INDEX idx_exam_attempts_score ON exam_attempts(total_score DESC); -- for leaderboard
CREATE INDEX idx_exam_attempts_created_at ON exam_attempts(created_at); -- for time-based queries

-- Exam answers indexes
CREATE INDEX idx_exam_answers_attempt_id ON exam_answers(attempt_id);
CREATE INDEX idx_exam_answers_question_id ON exam_answers(question_id);

-- +goose Down
-- Drop indexes
DROP INDEX IF EXISTS idx_exam_answers_question_id;
DROP INDEX IF EXISTS idx_exam_answers_attempt_id;
DROP INDEX IF EXISTS idx_exam_attempts_created_at;
DROP INDEX IF EXISTS idx_exam_attempts_score;
DROP INDEX IF EXISTS idx_exam_attempts_user_exam;
DROP INDEX IF EXISTS idx_exam_attempts_user_id;
DROP INDEX IF EXISTS idx_exam_attempts_exam_id;
DROP INDEX IF EXISTS idx_exam_attempts_uuid;
DROP INDEX IF EXISTS idx_exam_questions_exam_order;
DROP INDEX IF EXISTS idx_exam_questions_question_id;
DROP INDEX IF EXISTS idx_exam_questions_exam_id;
DROP INDEX IF EXISTS idx_exam_translations_exam_language;
DROP INDEX IF EXISTS idx_exams_created_at;
DROP INDEX IF EXISTS idx_exams_created_by;
DROP INDEX IF EXISTS idx_exams_difficulty;
DROP INDEX IF EXISTS idx_exams_public;
DROP INDEX IF EXISTS idx_exams_uuid;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS exam_answers;
DROP TABLE IF EXISTS exam_attempts;
DROP TABLE IF EXISTS exam_questions;
DROP TABLE IF EXISTS exam_translations;
DROP TABLE IF EXISTS exams;
