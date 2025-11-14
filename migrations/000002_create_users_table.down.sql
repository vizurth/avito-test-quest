-- 000002_create_users_table.down.sql
DROP INDEX IF EXISTS idx_users_user_id;
DROP INDEX IF EXISTS idx_users_team_id;
DROP INDEX IF EXISTS idx_users_is_active;
DROP TABLE IF EXISTS users;