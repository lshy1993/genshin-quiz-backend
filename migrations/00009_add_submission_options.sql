-- +goose Up
-- Add table to track selected options for each submission
CREATE TABLE question_submission_options (
    submission_id BIGINT NOT NULL REFERENCES question_submissions(id) ON DELETE CASCADE,
    option_id BIGINT NOT NULL REFERENCES question_options(id) ON DELETE CASCADE,

    UNIQUE(submission_id, option_id), -- 防止同一次提交重复选择同一个选项
    PRIMARY KEY(submission_id, option_id)
);

-- Create indexes for better query performance
CREATE INDEX idx_submission_options_option_id ON question_submission_options(option_id);

-- +goose Down
-- Remove table and indexes

DROP INDEX IF EXISTS idx_submission_options_option_id;
DROP TABLE IF EXISTS question_submission_options;