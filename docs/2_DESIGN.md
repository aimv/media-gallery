# Документ: docs/2_DESIGN.md

**Файл:** `docs/2_DESIGN.md`  
**Статус:** Утвержден для реализации  
**Роль:** Агент-Архитектор  
**Модуль:** `github.com/yourorg/media-gallery`

---

## 1. Структура Go-приложения

Проект следует принципам Clean Architecture. Зависимости направлены внутрь:  
`infrastructure -> usecase -> domain`

```text
media-gallery/
├── cmd/
│   ├── api/
│   │   └── main.go                     # Точка входа HTTP API
│   └── worker/
│       └── main.go                     # Точка входа фонового обработчика
├── internal/
│   ├── domain/
│   │   ├── media/
│   │   │   ├── asset.go                # Агрегат MediaAsset, статусы, типы
│   │   │   ├── job.go                  # Агрегат ProcessingJob
│   │   │   ├── ports.go                # Интерфейсы MediaRepository, JobQueue, VideoProcessor, FileStorage
│   │   │   └── errors.go               # Доменные ошибки
│   │   └── cms/
│   │       ├── block.go                # Агрегат ContentBlock
│   │       ├── ports.go                # Интерфейсы CMS-репозиториев
│   │       └── errors.go
│   ├── usecase/
│   │   ├── media/
│   │   │   ├── upload.go               # Use Case загрузки
│   │   │   ├── get.go                  # Use Case получения метаданных
│   │   │   ├── delete.go               # Use Case удаления
│   │   │   ├── process.go              # Use Case обработки видео
│   │   │   └── dto.go                  # DTO для слоя использования
│   │   └── cms/
│   │       ├── create_block.go
│   │       ├── update_block.go
│   │       ├── delete_block.go
│   │       ├── list_blocks.go
│   │       └── dto.go
│   ├── infrastructure/
│   │   ├── config/
│   │   │   └── config.go               # Загрузка ENV-конфигурации
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── migrations/
│   │   │   │   │   ├── 000001_init.up.sql
│   │   │   │   │   └── 000001_init.down.sql
│   │   │   │   ├── media_repository.go # Реализация MediaRepository
│   │   │   │   ├── job_queue.go        # Реализация JobQueue поверх PostgreSQL
│   │   │   │   └── cms_repository.go   # Реализация CMS-репозиториев
│   │   ├── filestorage/
│   │   │   └── local/
│   │   │       └── storage.go          # Локальная реализация FileStorage
│   │   ├── video/
│   │   │   └── ffmpeg/
│   │   │       └── processor.go        # Реализация VideoProcessor через ffmpeg/ffprobe
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── media.go            # HTTP handlers /v1/media
│   │   │   │   ├── cms.go              # HTTP handlers /v1/cms
│   │   │   │   └── health.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go             # Статический API key
│   │   │   │   ├── logging.go          # Request ID + structured logs
│   │   │   │   └── recovery.go
│   │   │   └── router.go               # Роутер, DTO-маппинг
│   │   └── worker/
│   │       ├── runner.go               # Poller очереди
│   │       └── processor.go            # Вызов usecase обработки
│   └── pkg/
│       ├── logger/
│       │   └── logger.go               # Обертка над slog/zerolog
│       └── apperror/
│           └── error.go                # Единый формат ошибок API
├── docs/
│   ├── 1_ANALYSIS.md
│   └── 2_DESIGN.md                     # Настоящий документ
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.worker
├── Makefile
├── go.mod
├── go.sum
├── .env.example
└── README.md
```

---

## 2. SQL-миграции PostgreSQL

### 2.1. Up-миграция: `000001_init.up.sql`

```sql
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
```

### 2.2. Down-миграция: `000001_init.down.sql`

```sql
DROP TRIGGER IF EXISTS trg_content_blocks_updated_at ON content_blocks;
DROP TRIGGER IF EXISTS trg_processing_jobs_updated_at ON processing_jobs;
DROP TRIGGER IF EXISTS trg_media_assets_updated_at ON media_assets;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS content_block_media_links;
DROP TABLE IF EXISTS content_blocks;
DROP TABLE IF EXISTS processing_jobs;
DROP TABLE IF EXISTS media_assets;
```

---

## 3. Go-интерфейсы портов инфраструктуры

Интерфейсы определяются на уровне `domain`, чтобы usecase-слой не зависел от реализаций. Реализации находятся в `infrastructure`.

### 3.1. Доменные типы (`internal/domain/media/asset.go`)

```go
package media

import "time"

type MediaType string

const (
    MediaTypeJPEG MediaType = "image/jpeg"
    MediaTypePNG  MediaType = "image/png"
    MediaTypeMP4  MediaType = "video/mp4"
)

type AssetStatus string

const (
    StatusUploaded    AssetStatus = "uploaded"
    StatusQueued      AssetStatus = "queued"
    StatusProcessing  AssetStatus = "processing"
    StatusReady       AssetStatus = "ready"
    StatusFailed      AssetStatus = "failed"
    StatusDeleting    AssetStatus = "deleting"
)

type MediaAsset struct {
    ID               string
    OriginalFileName string
    MediaType        MediaType
    Status           AssetStatus
    StoragePath      string
    HlsPath          *string
    SizeBytes        int64
    ChecksumSHA256   string
    Metadata         map[string]any
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### 3.2. Доменный тип задачи (`internal/domain/media/job.go`)

```go
package media

import "time"

type JobStatus string

const (
    JobStatusQueued     JobStatus = "queued"
    JobStatusProcessing JobStatus = "processing"
    JobStatusDone       JobStatus = "done"
    JobStatusFailed     JobStatus = "failed"
)

type ProcessingJob struct {
    ID             string
    AssetID        string
    Status         JobStatus
    Attempt        int
    MaxAttempts    int
    LeaseOwner     string
    LeaseExpiresAt time.Time
    LastError      string
    CreatedAt      time.Time
    StartedAt      *time.Time
    FinishedAt     *time.Time
}
```

### 3.3. Порты инфраструктуры (`internal/domain/media/ports.go`)

```go
package media

import (
    "context"
    "io"
    "time"
)

// MediaRepository — хранилище метаданных медиафайлов.
type MediaRepository interface {
    Save(ctx context.Context, asset *MediaAsset) error
    GetByID(ctx context.Context, id string) (*MediaAsset, error)
    List(ctx context.Context, filter ListFilter) ([]*MediaAsset, error)
    UpdateStatus(ctx context.Context, id string, status AssetStatus) error
    Delete(ctx context.Context, id string) error
}

type ListFilter struct {
    Status *AssetStatus
    Type   *MediaType
    Limit  int
    Offset int
}

// JobQueue — очередь фоновой обработки на базе PostgreSQL SKIP LOCKED.
type JobQueue interface {
    Enqueue(ctx context.Context, assetID string) (*ProcessingJob, error)
    Claim(ctx context.Context, owner string, leaseDuration time.Duration) (*ProcessingJob, error)
    Heartbeat(ctx context.Context, jobID string, owner string, leaseDuration time.Duration) error
    MarkDone(ctx context.Context, jobID string, owner string) error
    MarkFailed(ctx context.Context, jobID string, owner string, errMsg string) error
}

// VideoProcessor — интерфейс внешнего процесса обработки видео.
type VideoProcessor interface {
    Probe(ctx context.Context, filePath string) (*VideoMetadata, error)
    ConvertToHLS(ctx context.Context, inputPath, outputDir string) error
}

type VideoMetadata struct {
    Duration    time.Duration
    Width       int
    Height      int
    VideoCodec  string
    AudioCodec  string
}

// FileStorage — файловое хранилище (local/S3), изолируется за портом.
type FileStorage interface {
    Save(ctx context.Context, src io.Reader, dstPath string) error
    Delete(ctx context.Context, path string) error
    Open(ctx context.Context, path string) (io.ReadCloser, error)
    Move(ctx context.Context, src, dst string) error
}
```

### 3.4. Порты CMS (`internal/domain/cms/ports.go`)

```go
package cms

import "context"

type ContentBlockRepository interface {
    Save(ctx context.Context, block *ContentBlock) error
    GetByID(ctx context.Context, id string) (*ContentBlock, error)
    ListByPage(ctx context.Context, pageKey string) ([]*ContentBlock, error)
    Delete(ctx context.Context, id string) error
}

type MediaLinkRepository interface {
    Add(ctx context.Context, blockID, assetID string, position int) error
    Remove(ctx context.Context, blockID, assetID string) error
    ListAssetIDs(ctx context.Context, blockID string) ([]string, error)
}
```

---

## 4. Ключевые проектные решения

| Решение | Обоснование |
|---------|-------------|
| Очередь на PostgreSQL с `FOR UPDATE SKIP LOCKED` | Отсутствие внешней MQ для MVP, гарантия отсутствия двойной обработки |
| `lease_owner` + `lease_expires_at` | Перехват зависших задач при падении воркера |
| Частичный индекс `WHERE status = 'queued'` | Эффективный claim задач из очереди |
| `ON DELETE RESTRICT` для `asset_id` в CMS-ссылках | Защита от появления висячих ссылок на медиафайлы |
| JSONB для `data` в `content_blocks` | Гибкая схема для различных типов блоков |
| Интерфейс `FileStorage` | Возможность миграции с локального диска на S3 без изменения usecase |
| `VideoProcessor` как порт | Изоляция ffmpeg/ffprobe, тестируемость usecase |
| Статусная модель `media_assets` | Полный жизненный цикл файла от загрузки до удаления |

Документ готов к использованию на этапе реализации.