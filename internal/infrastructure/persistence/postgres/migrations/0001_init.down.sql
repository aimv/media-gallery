BEGIN; -- Открываем атомарную транзакцию

-- Удаляем таблицы в строгом обратном порядке (сначала зависимые таблицы связей)
DROP TABLE IF EXISTS block_media_links;
DROP TABLE IF EXISTS content_blocks;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS media_assets;

-- Удаляем кастомные пользовательские типы данных (ENUM)
DROP TYPE IF EXISTS block_type;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS media_status;
DROP TYPE IF EXISTS media_type;

-- Удаляем расширение генерации UUID (опционально, оставляем для чистоты окружения)
DROP EXTENSION IF EXISTS "pgcrypto";

COMMIT; -- Фиксируем изменения только в случае полного успеха!