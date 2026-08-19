CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Таблица media_assets
-- ============================================================
CREATE TABLE media_assets (
                              id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              original_file_name  TEXT NOT NULL,
                              media_type          VARCHAR(16) NOT NULL
                                  CHECK (media_type IN ('image/jpeg', 'image/png', 'video/mp4')),
                              status              VARCHAR(32) NOT NULL DEFAULT 'uploaded'
                                  CHECK (status IN ('uploaded', 'queued', 'processing', 'ready', 'failed', 'deleting')),
                              storage_path        TEXT NOT NULL,
                              hls_path            TEXT,
                              size_bytes          BIGINT NOT NULL CHECK (size_bytes >= 0),
                              checksum_sha256     CHAR(64) NOT NULL,
                              metadata            JSONB NOT NULL DEFAULT '{}'::jsonb,
                              created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
                              updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_media_assets_status_type
    ON media_assets(status, media_type);

CREATE INDEX idx_media_assets_created_at
    ON media_assets(created_at DESC);

-- ============================================================
-- Таблица processing_jobs
-- ============================================================
CREATE TABLE processing_jobs (
                                 id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 asset_id            UUID NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
                                 status              VARCHAR(32) NOT NULL DEFAULT 'queued'
                                     CHECK (status IN ('queued', 'processing', 'done', 'failed')),
                                 attempt             INT NOT NULL DEFAULT 0 CHECK (attempt >= 0),
                                 max_attempts        INT NOT NULL DEFAULT 3 CHECK (max_attempts > 0),
                                 lease_owner         VARCHAR(64),
                                 lease_expires_at    TIMESTAMPTZ,
                                 last_error          TEXT,
                                 created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
                                 started_at          TIMESTAMPTZ,
                                 finished_at         TIMESTAMPTZ,
                                 updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Критически важный индекс для выборки задач с SKIP LOCKED
CREATE INDEX idx_processing_jobs_queued
    ON processing_jobs(created_at)
    WHERE status = 'queued';

CREATE INDEX idx_processing_jobs_lease
    ON processing_jobs(lease_expires_at)
    WHERE status = 'processing';

CREATE INDEX idx_processing_jobs_asset_id
    ON processing_jobs(asset_id);

-- ============================================================
-- Таблица content_blocks (CMS)
-- ============================================================
CREATE TABLE content_blocks (
                                id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                page_key            VARCHAR(128) NOT NULL,
                                block_type          VARCHAR(32) NOT NULL
                                    CHECK (block_type IN ('text', 'carousel', 'reviews', 'about')),
                                data                JSONB NOT NULL DEFAULT '{}'::jsonb,
                                sort_order          INT NOT NULL DEFAULT 0,
                                created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
                                updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_content_blocks_page_sort
    ON content_blocks(page_key, sort_order);

CREATE INDEX idx_content_blocks_type
    ON content_blocks(block_type);

-- ============================================================
-- Связь контентных блоков с медиафайлами
-- ============================================================
CREATE TABLE content_block_media_links (
                                           block_id            UUID NOT NULL REFERENCES content_blocks(id) ON DELETE CASCADE,
                                           asset_id            UUID NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
                                           position            INT NOT NULL DEFAULT 0,
                                           PRIMARY KEY (block_id, asset_id),
                                           UNIQUE (block_id, position)
);

-- ============================================================
-- Триггер для обновления updated_at
-- ============================================================
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_media_assets_updated_at
    BEFORE UPDATE ON media_assets
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_processing_jobs_updated_at
    BEFORE UPDATE ON processing_jobs
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_content_blocks_updated_at
    BEFORE UPDATE ON content_blocks
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();