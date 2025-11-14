-- 000003_create_pull_requests_table.down.sql
DROP INDEX IF EXISTS idx_pr_pull_request_id;
DROP INDEX IF EXISTS idx_pr_author_id;
DROP INDEX IF EXISTS idx_pr_status;
DROP TABLE IF EXISTS pull_requests;
DROP TYPE IF EXISTS pr_status;