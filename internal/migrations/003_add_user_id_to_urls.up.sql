-- Добавление колонки user_id в таблицу urls
ALTER TABLE urls ADD COLUMN IF NOT EXISTS user_id UUID;

-- Создание индекса для user_id
CREATE INDEX IF NOT EXISTS idx_urls_user_id ON urls (user_id);