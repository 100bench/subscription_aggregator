-- Remove indexes added in 0002_add_indexes.up.sql
DROP INDEX IF EXISTS idx_subs_dates;
DROP INDEX IF EXISTS idx_subs_user_created_at;
