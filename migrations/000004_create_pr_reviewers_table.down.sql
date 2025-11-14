-- 000004_create_pr_reviewers_table.down.sql
DROP INDEX IF EXISTS idx_pr_reviewers_pr_id;
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;
DROP TABLE IF EXISTS pr_reviewers;