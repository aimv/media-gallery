CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Таблица media_assets
-- ============================================================

CREATE TYPE media_type AS ENUM ('image/jpeg', 'image/png', 'video/mp4');
CREATE TYPE media_status AS ENUM ('uploaded', 'queued', 'processing', 'ready', 'failed', 'deleting');

CREATE TABLE media_assets (
      id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      original_filename  TEXT NOT NULL,
      media_type         media_type NOT NULL,
      status             media_status NOT NULL DEFAULT 'uploaded',
      storage_path       TEXT NOT NULL,             -- путь к оригиналу относительно корня хранилища
      hls_path           TEXT,                      -- путь к master.m3u8 (NULL до готовности)
      size_bytes         BIGINT NOT NULL,
      checksum_sha256    VARCHAR(64),               -- может заполняться после загрузки
      width              INTEGER,
      height             INTEGER,
      duration_ms        BIGINT,
      codec              TEXT,
      metadata           JSONB NOT NULL DEFAULT '{}'::jsonb,
      created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
      deleted_at         TIMESTAMPTZ                -- мягкое удаление (NULL если активен)
);

CREATE INDEX idx_media_assets_status ON media_assets (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_type ON media_assets (media_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_created_at ON media_assets (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_checksum ON media_assets (checksum_sha256) WHERE deleted_at IS NULL;

-- ============================================================
-- Таблица processing_jobs
-- ============================================================

CREATE TYPE job_status AS ENUM ('queued', 'processing', 'success', 'failed');

CREATE TABLE processing_jobs (
      id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      asset_id         UUID NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
      status           job_status NOT NULL DEFAULT 'queued',
      attempt          INTEGER NOT NULL DEFAULT 0,
      max_attempts     INTEGER NOT NULL DEFAULT 3,
      lease_owner      TEXT,                         -- идентификатор воркера (hostname:pid)
      lease_expires_at TIMESTAMPTZ,
      error            TEXT,
      created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
      started_at       TIMESTAMPTZ,
      finished_at      TIMESTAMPTZ,
      updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Гарантируем одну активную задачу на медиафайл
CREATE UNIQUE INDEX uq_processing_jobs_asset_active ON processing_jobs (asset_id)
    WHERE status IN ('queued', 'processing');

-- Индекс для выборки задач с SKIP LOCKED
CREATE INDEX idx_processing_jobs_queue ON processing_jobs (status, created_at)
    WHERE status = 'queued';

-- Индекс для поиска просроченных lease
CREATE INDEX idx_processing_jobs_lease ON processing_jobs (lease_expires_at)
    WHERE status = 'processing';

-- ============================================================
-- Таблица content_blocks (CMS)
-- ============================================================

CREATE TYPE block_type AS ENUM ('text', 'carousel', 'reviews', 'about');

CREATE TABLE content_blocks (
      id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      page_key    TEXT NOT NULL,
      type        block_type NOT NULL,
      data        JSONB NOT NULL DEFAULT '{}'::jsonb,   -- содержимое блока, специфичное для типа
      sort_order  INTEGER NOT NULL DEFAULT 0,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Индекс для выборки блоков страницы с сортировкой
CREATE INDEX idx_content_blocks_page ON content_blocks (page_key, sort_order);

-- ============================================================
-- Таблица block_media_links - связь контентных блоков с медиафайлами
-- ============================================================

CREATE TABLE block_media_links (
      id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
      block_id    UUID NOT NULL REFERENCES content_blocks(id) ON DELETE CASCADE,
      asset_id    UUID NOT NULL REFERENCES media_assets(id) ON DELETE RESTRICT,
      position    INTEGER NOT NULL DEFAULT 0,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
      UNIQUE (block_id, asset_id)   -- один медиафайл может быть привязан к блоку только один раз
);

-- Индекс для быстрого поиска блоков по медиафайлу
CREATE INDEX idx_block_media_links_asset ON block_media_links (asset_id);
-- Индекс для выборки медиафайлов блока с сортировкой
CREATE INDEX idx_block_media_links_block ON block_media_links (block_id, position);

