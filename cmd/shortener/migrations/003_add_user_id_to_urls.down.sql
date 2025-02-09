-- Удаление индекса для user_id
DROP INDEX IF EXISTS idx_urls_user_id;

-- Удаление колонки user_id
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;