-- 000004_create_pr_reviewers_table.up.sql
CREATE TABLE IF NOT EXISTS pr_reviewers (
    id SERIAL PRIMARY KEY,
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_user_id VARCHAR(255) NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(pull_request_id, reviewer_user_id)
    );

CREATE INDEX idx_pr_reviewers_pr_id ON pr_reviewers(pull_request_id);
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(reviewer_user_id);