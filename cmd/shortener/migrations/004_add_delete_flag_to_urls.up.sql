-- Добавление колонки delete_flag в таблицу urls
ALTER TABLE urls ADD COLUMN IF NOT EXISTS delete_flag BOOLEAN DEFAULT FALSE;