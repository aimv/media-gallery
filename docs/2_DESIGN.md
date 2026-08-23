# Архитектура микросервиса медиагалереи с CMS

## 1. Введение

Документ описывает глобальную архитектуру микросервиса медиагалереи с CMS. Проектирование выполнено в соответствии с принципами Clean Architecture и DDD, с учётом анализа из `docs/1_ANALYSIS.md`. Основное внимание уделено структуре Go-приложения, интерфейсам портов инфраструктуры и схеме базы данных.

---
## 2. Структура Go-приложения (Clean Architecture)

```
.
├── cmd
│   ├── api
│   │   └── main.go              # Точка входа HTTP API
│   └── worker
│       └── main.go              # Точка входа фонового обработчика
├── internal
│   ├── domain                    # Слой домена (Entities, Value Objects, Ports)
│   │   ├── entity
│   │   │   ├── media_asset.go
│   │   │   ├── processing_job.go
│   │   │   ├── content_block.go
│   │   │   └── enums.go         # Перечисления статусов, типов
│   │   ├── repository            # Интерфейсы репозиториев (порты для БД)
│   │   │   ├── media_repository.go
│   │   │   ├── job_queue.go
│   │   │   ├── cms_repository.go
│   │   │   └── transaction.go   # Опциональный интерфейс транзакций
│   │   └── service               # Интерфейсы внешних сервисов (порты для инфраструктуры)
│   │       ├── video_processor.go
│   │       └── file_storage.go
│   ├── usecase                   # Слой прикладной логики (Use Cases)
│   │   ├── media
│   │   │   ├── upload.go
│   │   │   ├── get_media.go
│   │   │   ├── list_media.go
│   │   │   ├── delete_media.go
│   │   │   └── dto.go
│   │   ├── processing
│   │   │   ├── start_processing.go
│   │   │   ├── poll_jobs.go
│   │   │   └── dto.go
│   │   └── cms
│   │       ├── manage_blocks.go
│   │       ├── link_media.go
│   │       └── dto.go
│   ├── infrastructure            # Слой инфраструктуры (реализации портов)
│   │   ├── persistence
│   │   │   ├── postgres
│   │   │   │   ├── media_repository_pg.go
│   │   │   │   ├── job_queue_pg.go
│   │   │   │   ├── cms_repository_pg.go
│   │   │   │   ├── transaction_pg.go
│   │   │   │   └── migrations/   # SQL-миграции
│   │   │   └── models            # ORM-модели (если используются)
│   │   ├── filestorage
│   │   │   └── local_fs.go
│   │   ├── video
│   │   │   ├── ffmpeg_processor.go
│   │   │   └── ffprobe.go
│   │   ├── http
│   │   │   ├── handler          # HTTP-хендлеры
│   │   │   ├── middleware
│   │   │   ├── router.go
│   │   │   └── dto.go
│   │   └── config
│   │       └── config.go         # Загрузка конфигурации из ENV
│   └── pkg                       # Общие утилиты (логирование, ошибки)
│       ├── logger
│       └── apperror
├── docs
│   ├── 1_ANALYSIS.md
│   └── 2_DESIGN.md
├── migrations                    # Альтернативное расположение миграций
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.worker
├── go.mod
└── go.sum
```

**Примечания по структуре:**

- Каталог `internal/domain` содержит только бизнес-сущности и интерфейсы портов; он не должен импортировать внешние пакеты (кроме стандартной библиотеки).
- Каталог `internal/usecase` реализует прикладные сценарии, используя порты из `domain`.
- Каталог `internal/infrastructure` содержит конкретные реализации: PostgreSQL, локальная файловая система, ffmpeg, HTTP-сервер.
- `cmd/api` и `cmd/worker` отвечают за композицию зависимостей и запуск соответствующих компонентов.

---

## 3. Интерфейсы портов инфраструктуры

Порты определены в `internal/domain/repository` и `internal/domain/service`. Они представляют собой чистые интерфейсы Go, не зависящие от конкретных технологий.

### 3.1. MediaRepository

Управление метаданными медиафайлов.

```go
type MediaRepository interface {
    // Create сохраняет новый медиафайл и возвращает его ID.
    Create(ctx context.Context, asset *entity.MediaAsset) (string, error)

    // GetByID возвращает медиафайл по ID.
    GetByID(ctx context.Context, id string) (*entity.MediaAsset, error)

    // List возвращает список медиафайлов с фильтрацией и пагинацией.
    List(ctx context.Context, filter MediaFilter, limit, offset int) ([]*entity.MediaAsset, error)

    // UpdateStatus атомарно изменяет статус медиафайла.
    UpdateStatus(ctx context.Context, id string, oldStatus, newStatus entity.MediaStatus) error

    // UpdateMetadata обновляет метаданные (размер, чексумму, параметры видео).
    UpdateMetadata(ctx context.Context, id string, metadata map[string]interface{}) error

    // SetHlsPath сохраняет путь к HLS-плейлисту после успешной обработки.
    SetHlsPath(ctx context.Context, id string, hlsPath string) error

    // Delete мягко удаляет запись (помечает как удалённую или полностью удаляет, если нет связей).
    Delete(ctx context.Context, id string) error
}

type MediaFilter struct {
    Status   *entity.MediaStatus
    Type     *entity.MediaType
    PageKey  *string
    Search   *string
}
```

### 3.2. JobQueue

Очередь задач на базе PostgreSQL (SKIP LOCKED). Реализует паттерн «надёжной очереди».

```go
type JobQueue interface {
    // Enqueue добавляет новую задачу в очередь.
    Enqueue(ctx context.Context, assetID string) error

    // ClaimNext атомарно выбирает следующую задачу со статусом 'queued',
    // блокируя её за воркером (owner) с установкой lease.
    // Возвращает nil, если задач нет.
    ClaimNext(ctx context.Context, owner string) (*entity.ProcessingJob, error)

    // Heartbeat продлевает lease задачи, пока она выполняется.
    Heartbeat(ctx context.Context, jobID string, owner string, newExpireAt time.Time) error

    // Complete завершает задачу с указанием успеха или ошибки.
    // Если success=false, увеличивает счётчик попыток и при необходимости возвращает задачу в 'queued' (retry)
    // или окончательно помечает как 'failed'.
    Complete(ctx context.Context, jobID string, success bool, errorMsg string) error

    // Requeue вручную возвращает задачу в очередь (например, после сбоя).
    Requeue(ctx context.Context, jobID string) error

    // GetByAssetID возвращает активную задачу для медиафайла (если есть).
    GetByAssetID(ctx context.Context, assetID string) (*entity.ProcessingJob, error)
}
```

### 3.3. VideoProcessor

Сервис для работы с видео через ffmpeg/ffprobe.

```go
type VideoProcessor interface {
    // Validate выполняет базовую проверку файла с помощью ffprobe:
    // контейнер MP4, кодеки, длительность, наличие видео/аудио потока.
    // Возвращает структуру VideoMetadata с параметрами.
    Validate(ctx context.Context, filePath string) (*entity.VideoMetadata, error)

    // ProcessToHLS конвертирует видео в HLS-сегменты и плейлист.
    // Входной файл filePath, выходная директория outputDir (должна быть временной).
    // Функция должна быть атомарной: либо все файлы сгенерированы, либо возвращается ошибка.
    // Прогресс можно логировать, но не обязательно.
    ProcessToHLS(ctx context.Context, filePath, outputDir string) error

    // ProbeMetadata извлекает детальные метаданные видео (разрешение, битрейт, длительность).
    ProbeMetadata(ctx context.Context, filePath string) (*entity.VideoMetadata, error)
}
```

### 3.4. FileStorage

Абстракция над файловым хранилищем (локальный диск, S3 в будущем).

```go
type FileStorage interface {
    // Save сохраняет данные из src в файл по пути path.
    // Если файл существует, перезаписывает его.
    Save(ctx context.Context, src io.Reader, path string) error

    // Open открывает файл для чтения.
    Open(ctx context.Context, path string) (io.ReadCloser, error)

    // Delete удаляет файл или директорию по пути.
    Delete(ctx context.Context, path string) error

    // Move атомарно перемещает файл/директорию из src в dst (в пределах одного хранилища).
    Move(ctx context.Context, src, dst string) error

    // Exists проверяет существование файла/директории.
    Exists(ctx context.Context, path string) (bool, error)

    // EnsureDir создаёт директорию (и все родительские), если её нет.
    EnsureDir(ctx context.Context, path string) error
}
```

### 3.5. CMSRepository

Управление динамическими блоками страниц и связями с медиафайлами.

```go
type CMSRepository interface {
    // CreateBlock создаёт новый контентный блок.
    CreateBlock(ctx context.Context, block *entity.ContentBlock) (string, error)

    // GetBlock возвращает блок по ID.
    GetBlock(ctx context.Context, id string) (*entity.ContentBlock, error)

    // ListBlocksByPage возвращает все блоки для указанной страницы, отсортированные по SortOrder.
    ListBlocksByPage(ctx context.Context, pageKey string) ([]*entity.ContentBlock, error)

    // UpdateBlock обновляет данные блока (Data, SortOrder, Type).
    UpdateBlock(ctx context.Context, block *entity.ContentBlock) error

    // DeleteBlock удаляет блок и все его связи с медиафайлами.
    DeleteBlock(ctx context.Context, id string) error

    // LinkMedia добавляет связь между блоком и медиафайлом с указанием позиции.
    LinkMedia(ctx context.Context, blockID, assetID string, position int) error

    // UnlinkMedia удаляет связь.
    UnlinkMedia(ctx context.Context, blockID, assetID string) error

    // ListMediaByBlock возвращает список медиафайлов, привязанных к блоку, отсортированных по позиции.
    ListMediaByBlock(ctx context.Context, blockID string) ([]*entity.MediaAsset, error)

    // ListBlocksByMedia возвращает блоки, использующие данный медиафайл.
    ListBlocksByMedia(ctx context.Context, assetID string) ([]*entity.ContentBlock, error)
}
```
---
## 4. Схема базы данных

База данных PostgreSQL используется как источник истины для метаданных и очередь задач. Все таблицы имеют `created_at` и `updated_at`. Для идемпотентности и аудита.

### 4.1. Таблица `media_assets`

Хранит метаданные о загруженных медиафайлах.

```sql
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

-- Индексы
CREATE INDEX idx_media_assets_status ON media_assets (status) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_type ON media_assets (media_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_created_at ON media_assets (created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_media_assets_checksum ON media_assets (checksum_sha256) WHERE deleted_at IS NULL;
```

**Примечания:**
- `status` переходит по жизненному циклу: `uploaded` → `queued` (если видео) или сразу `ready` (для изображений) → `processing` → `ready` / `failed`; при удалении — `deleting`.
- `storage_path` и `hls_path` содержат относительные пути, например: `originals/{uuid}/file.mp4`, `hls/{uuid}/master.m3u8`.
- `metadata` — JSONB для хранения произвольных метаданных (длительность, разрешение и т.п.).
- `deleted_at` используется для мягкого удаления; реальные файлы удаляются после успешного обновления статуса `deleting`.

### 4.2. Таблица `processing_jobs`

Очередь задач обработки видео. Реализует механизм `FOR UPDATE SKIP LOCKED` для конкурентного выбора.

```sql
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
```

**Пояснения:**
- `FOR UPDATE SKIP LOCKED` используется в запросе на выборку задач (см. SQL ниже).
- `lease_owner` и `lease_expires_at` обеспечивают механизм перехвата зависших задач.
- `attempt` увеличивается при каждом неудачном завершении; при превышении `max_attempts` статус становится `failed`.
- Уникальный индекс на `asset_id` с условием `status IN ('queued','processing')` предотвращает создание дублирующих задач для одного медиафайла.
- При успешной обработке статус меняется на `success`, а связанный `media_assets.status` — на `ready`.

**Пример запроса для claim задачи:**

```sql
UPDATE processing_jobs
SET status = 'processing',
    lease_owner = $1,
    lease_expires_at = now() + interval '5 minutes',
    started_at = COALESCE(started_at, now()),
    updated_at = now()
WHERE id = (
    SELECT id
    FROM processing_jobs
    WHERE status = 'queued'
    ORDER BY created_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING *;
```

### 4.3. Таблица `content_blocks`

Хранит динамические блоки страниц CMS.

```sql
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
```

### 4.4. Таблица `block_media_links`

Связь многие-ко-многим между контентными блоками и медиафайлами.

```sql
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
```

**Внешний ключ `asset_id` с `ON DELETE RESTRICT`** гарантирует, что медиафайл, используемый в CMS, не будет удалён из `media_assets` без явного удаления связей.

---

## 5. Миграции

Миграции управляются инструментом [golang-migrate](https://github.com/golang-migrate/migrate) или аналогичным. Файлы миграций расположены в `internal/infrastructure/persistence/postgres/migrations/` и нумеруются последовательно. При старте сервиса миграции применяются автоматически.

Пример первых файлов:

- `0001_init.up.sql` – создание enum-типов, таблиц `media_assets`, `processing_jobs`, `content_blocks`, `block_media_links`, индексов и ограничений.
- `0001_init.down.sql` – обратные операции.

---

## 6. Ключевые проектные решения (Architectural Decisions)

В данном разделе фиксируются и обосновываются важные архитектурные решения, принятые при проектировании микросервиса. Каждое решение описывается в формате «Решение — Обоснование — Альтернативы».

### 6.1. Использование UUIDv4 в качестве первичных ключей

**Решение:** Для всех таблиц (`media_assets`, `processing_jobs`, `content_blocks`, `block_media_links`) первичным ключом используется `UUID` версии 4, генерируемый на стороне приложения или БД (`gen_random_uuid()`).

**Обоснование:**
- **Безопасность API.** Последовательные целочисленные идентификаторы (`BIGSERIAL`) легко перебираются, что упрощает несанкционированный доступ к ресурсам и утечку информации о количестве записей. UUID непредсказуемы, что снижает риск атак перебора.
- **Распределённость и отказоустойчивость.** UUID могут генерироваться независимо в разных экземплярах приложения без координации с БД, что критично при горизонтальном масштабировании API и воркеров. Это устраняет необходимость в централизованном генераторе ID.
- **Простота горизонтального партиционирования.** В будущем при переходе на шардирование или распределённые хранилища UUID позволяют избежать коллизий и не требуют сложных схем выделения диапазонов ключей.
- **Идемпотентность операций.** При использовании UUID клиент может заранее генерировать ID для повторных запросов, что упрощает реализацию идемпотентных загрузок (хотя в MVP данная возможность не используется).

**Альтернативы:**
- `BIGSERIAL` / `IDENTITY` — просты и компактны, но раскрывают внутреннюю информацию и создают узкое место при генерации.
- `ULID` / `UUIDv7` — сортируемые по времени, но требуют дополнительных библиотек; для текущих задач не дают существенного преимущества перед UUIDv4.


### 6.2. Потоковая обработка файлов через `io.Reader` в `FileStorage`

**Решение:** Интерфейс `FileStorage.Save` принимает `io.Reader` в качестве источника данных, а не путь к временному файлу или байтовый слайс.

```go
Save(ctx context.Context, src io.Reader, path string) error
```

**Обоснование:**
- **Эффективное использование памяти.** Загрузка больших видеофайлов (до 500 МБ) целиком в память в виде `[]byte` может привести к исчерпанию RAM и падению сервиса. Потоковая передача через `io.Reader` позволяет копировать данные чанками фиксированного размера, сохраняя константное потребление памяти независимо от размера файла.
- **Совместимость с HTTP.** Обработчик `multipart/form-data` предоставляет `io.Reader` для каждого загружаемого файла, что устраняет необходимость во временном сохранении на диск перед записью в хранилище.
- **Гибкость реализации.** Реализация `LocalFileStorage` может использовать `io.Copy` для эффективного копирования, а будущая реализация для S3 — multipart upload, также принимающий поток.
- **Возможность вычисления хеша на лету.** Обёртывание `io.Reader` в `io.TeeReader` позволяет одновременно вычислять SHA-256 и записывать файл, не сохраняя его в память.

**Альтернативы:**
- Сохранение временного файла и передача пути — создаёт дополнительные операции ввода-вывода и усложняет управление временными файлами.
- Чтение всего файла в `[]byte` — неприемлемо для видеофайлов большого размера.


### 6.3. Мягкое удаление медиафайлов через `deleted_at`

**Решение:** Для таблицы `media_assets` вводится колонка `deleted_at TIMESTAMPTZ NULL`. Удаление медиафайла не приводит к немедленному физическому удалению строки и файлов; вместо этого:
1. Статус меняется на `deleting`.
2. Выполняется асинхронное (или синхронное в рамках запроса) удаление файлов с диска.
3. После успешного удаления файлов строка помечается `deleted_at = now()` (мягкое удаление) либо физически удаляется (при отсутствии связей с CMS).

В MVP допустимо использовать комбинированный подход: строка остаётся с `deleted_at` для аудита и предотвращения повторного использования ID, но исключается из всех обычных выборок.

**Обоснование:**
- **Предотвращение висячих ссылок в CMS.** При использовании внешнего ключа `ON DELETE RESTRICT` в таблице `block_media_links` физическое удаление медиафайла, на который есть ссылки, будет заблокировано на уровне БД. Мягкое удаление позволяет сохранить целостность связей и отложить решение о судьбе связанного контента.
- **Отказоустойчивость при удалении файлов.** Операция удаления файлов с диска может завершиться ошибкой (например, из-за сетевого сбоя при использовании S3). Мягкое удаление даёт возможность повторить попытку физического удаления позже, не теряя информацию о том, что ресурс помечен как удалённый.
- **Аудит и восстановление.** Хранение удалённых записей позволяет отслеживать историю операций, а при необходимости — восстанавливать ошибочно удалённые метаданные (хотя файлы уже могут быть утеряны).
- **Упрощение конкурентных сценариев.** Если во время обработки видео приходит запрос на удаление, статус `deleting` сигнализирует воркеру о необходимости прекратить обработку и корректно завершить контекст. После завершения обработки удаление может быть выполнено безопасно.

**Альтернативы:**
- Физическое удаление без `deleted_at` — проще, но рискованно при наличии связей и внешних ошибок.
- Полное каскадное удаление (`ON DELETE CASCADE`) — приводит к неявной потере данных CMS-блоков, что неприемлемо.


### 6.4. Генерация HLS в временной директории с атомарным перемещением

**Решение:** При обработке видео ffmpeg записывает все выходные файлы (`.m3u8` и `.ts`) в отдельную временную директорию внутри хранилища (например, `tmp/{assetID}/hls/`). После успешного завершения процесса и проверки целостности директория атомарно перемещается в финальное расположение `hls/{assetID}/`.

**Обоснование:**
- **Предотвращение видимости частично записанных данных.** HLS-стрим состоит из множества файлов; если ffmpeg упадёт на середине генерации, клиент, обратившийся к `master.m3u8`, получит ссылки на несуществующие сегменты. Атомарный `rename` скрывает все изменения до полной готовности.
- **Идемпотентность повторной обработки.** Если по какой-то причине задача будет переобработана (retry после сбоя), временная директория может быть перезаписана без влияния на уже готовый результат. Финальная директория остаётся неизменной до полного успеха.
- **Упрощение очистки.** При ошибке достаточно удалить только временную директорию, не рискуя затронуть предыдущую версию HLS.

**Альтернативы:**
- Прямая запись в финальную директорию — приводит к «битым» плейлистам и нарушению консистентности.
- Использование уникальных версий каталогов — усложняет маршрутизацию и требует дополнительного указателя в БД.


### 6.5. Отдельные процессы API и Worker в Docker Compose

**Решение:** В архитектуре выделены два независимых процесса: `api` и `worker`. Оба запускаются из одного кодовой базы, но используют разные точки входа (`cmd/api` и `cmd/worker`). В `docker-compose.yml` они описаны как отдельные сервисы, использующие один и тот же образ (возможно, с разными command).

**Обоснование:**
- **Изоляция ресурсоёмких задач.** Обработка видео (ffmpeg) потребляет значительные CPU и память. Запуск воркера отдельным контейнером позволяет ограничить его ресурсы (через Docker limits) и не допустить деградации HTTP API при пиковых нагрузках.
- **Независимое масштабирование.** Можно увеличивать количество реплик воркера для ускорения обработки очереди, не затрагивая API.
- **Устойчивость к сбоям.** Падение воркера не влияет на доступность API; клиенты по-прежнему могут загружать файлы и получать метаданные.
- **Упрощение graceful shutdown.** Разные стратегии остановки: API завершает обработку HTTP-запросов, воркер — активные ffmpeg-процессы.

**Альтернативы:**
- Встроенный воркер в API (горутины) — проще для локальной разработки, но создаёт проблемы с ресурсами и горизонтальным масштабированием.
- Отдельный микросервис с собственной кодовой базой — неоправданное усложнение на этапе MVP.


### 6.6. Паттерн «Порты и адаптеры» для внешних зависимостей

**Решение:** Все взаимодействия с PostgreSQL, файловой системой, ffmpeg/ffprobe и HTTP-обработчиками абстрагированы через интерфейсы, определённые в слое `domain` (`MediaRepository`, `JobQueue`, `VideoProcessor`, `FileStorage`, `CMSRepository`). Реализации находятся в `internal/infrastructure`.

**Обоснование:**
- **Тестируемость.** Use-case слой можно тестировать с моками портов без реальной БД и внешних процессов.
- **Заменяемость.** Будущая миграция с локального диска на S3 потребует только новой реализации `FileStorage` без изменения бизнес-логики.
- **Соответствие Clean Architecture.** Зависимости направлены внутрь: `interfaces → application → domain`; `domain` не знает о технологиях.
- **Развитие без регрессий.** Добавление нового брокера очередей (Kafka) не затрагивает доменные правила.

**Альтернативы:**
- Прямое использование драйверов и SDK в use-case — приводит к сильной связанности и усложняет тестирование.
- Использование фреймворков (GORM, Echo) с прямым внедрением — нарушает инверсию зависимостей.


### 6.7. Использование JSONB для контентных блоков CMS

**Решение:** Таблица `content_blocks` содержит колонку `data JSONB`, в которой хранится произвольная структура данных, специфичная для каждого типа блока (`text`, `carousel`, `reviews`, `about`). Связи с медиафайлами вынесены в отдельную таблицу `block_media_links`.

**Обоснование:**
- **Гибкость схемы.** Разные типы блоков имеют разные поля: текст — `content`, карусель — список `asset_id`, отзывы — массив отзывов и т.д. JSONB позволяет хранить их без создания множества таблиц с наследованием.
- **Простота расширения.** Добавление нового типа блока не требует миграций схемы, достаточно добавить валидацию в приложении.
- **Производительность.** PostgreSQL поддерживает индексацию JSONB (GIN) и частичные выборки, чего достаточно для CMS-нагрузки MVP.
- **Разделение ответственности.** Связи с медиафайлами хранятся отдельно в нормализованном виде, что позволяет эффективно выполнять запросы вида «найти все блоки, использующие данный медиафайл».

**Альтернативы:**
- Отдельные таблицы для каждого типа блока — строгая схема, но избыточна и усложняет API.
- Хранение JSON в `TEXT` — теряет возможности запросов и индексации.


### 6.8. Ограничение размера загружаемых файлов на уровне API

**Решение:** Лимит на размер тела запроса устанавливается в HTTP-сервере (например, `http.MaxBytesReader`) и дополнительно проверяется после получения файла. Максимальные значения: изображения — 10 МБ, видео — 500 МБ (конфигурируются через переменные окружения).

**Обоснование:**
- **Защита от DoS.** Предотвращает переполнение диска и исчерпание памяти при загрузке аномально больших файлов.
- **Предсказуемое потребление ресурсов.** Позволяет оценить объём хранилища и нагрузку на сеть.
- **Ранний отказ.** Клиент получает ошибку `413 Payload Too Large` до того, как файл будет полностью загружен, что экономит пропускную способность.

**Альтернативы:**
- Отсутствие лимитов — риск полного исчерпания диска и недоступности сервиса.
- Проверка только после загрузки — не предотвращает передачу лишних данных.


### 6.9. Использование `exec.CommandContext` для управления ffmpeg

**Решение:** Все вызовы ffmpeg/ffprobe осуществляются через Go-пакет `os/exec` с обязательным использованием `exec.CommandContext`. Контекст создаётся с таймаутом и привязывается к жизненному циклу задачи. При отмене контекста процесс получает `SIGTERM`, а по истечении grace-периода — `SIGKILL`.

**Обоснование:**
- **Контроль времени выполнения.** Таймаут предотвращает бесконечное зависание воркера на сломанном файле.
- **Корректное завершение при остановке сервиса.** При graceful shutdown все активные ffmpeg-процессы получают сигнал завершения, что исключает «зомби» и утечки ресурсов.
- **Единый механизм отмены.** Контекст позволяет связать отмену задачи (например, при истечении lease) с завершением внешнего процесса.
- **Безопасность процессов.** При установке `Setpgid` можно убивать всю группу процессов, если ffmpeg порождает дочерние процессы.

**Альтернативы:**
- Запуск без контекста — риск зависших процессов и невозможность graceful shutdown.
- Использование библиотек-обёрток (например, `ffmpeg-go`) — добавляет зависимость, но не решает фундаментальных проблем управления процессами.


### 6.10. Валидация MIME-типа по содержимому (magic bytes)

**Решение:** При загрузке файла проверяется не только заголовок `Content-Type` и расширение, но и первые байты файла с помощью `http.DetectContentType` (для изображений) и `ffprobe` (для видео). Только после этого файл сохраняется.

**Обоснование:**
- **Защита от подделки расширений.** Злоумышленник может загрузить исполняемый файл, назвав его `photo.jpg`; проверка сигнатур предотвращает это.
- **Соответствие ожидаемым форматам.** Гарантируется, что в хранилище попадают только валидные JPEG/PNG/MP4.
- **Снижение риска обработки вредоносных файлов.** ffmpeg имеет историю уязвимостей; предварительная проверка контейнера снижает поверхность атаки.

**Альтернативы:**
- Доверие расширению и `Content-Type` — небезопасно, легко обходится.
- Полный разбор файла до загрузки — требует дополнительных ресурсов, но `ffprobe` уже выполняется для видео.


### 6.11. Автоматическое применение миграций при старте

**Решение:** При запуске API и worker миграции базы данных применяются автоматически с использованием инструмента `golang-migrate` (или аналогичного). Миграции хранятся в отдельной директории и накатываются последовательно.

**Обоснование:**
- **Простота деплоя.** Нет необходимости в отдельном шаге `migrate` в CI/CD или ручных командах; `docker-compose up` сразу приводит БД в актуальное состояние.
- **Согласованность окружений.** Разработка, тестирование и production используют одни и те же миграции, что снижает риск рассинхронизации схемы.
- **Версионирование схемы.** Каждая миграция соответствует определённой версии приложения.

**Альтернативы:**
- Ручное применение миграций — подвержено ошибкам и замедляет развёртывание.
- Миграции на уровне ORM (GORM AutoMigrate) — менее контролируемо и не подходит для сложных изменений.


### 6.12. Хранение конфигурации исключительно в переменных окружения

**Решение:** Все параметры конфигурации (DSN, пути хранилища, лимиты, количество воркеров, таймауты) загружаются из переменных окружения. В коде отсутствуют жёстко закодированные значения.

**Обоснование:**
- **Соответствие 12-factor app.** Позволяет изменять конфигурацию без пересборки образа.
- **Безопасность.** Секреты (пароли, ключи API) не попадают в код и систему контроля версий.
- **Гибкость деплоя.** Один и тот же образ может использоваться в разных окружениях (dev, staging, prod) с разными параметрами.

**Альтернативы:**
- Конфигурационные файлы (YAML, JSON) — требуют монтирования в контейнер и усложняют управление.
- Аргументы командной строки — менее удобны для контейнерной среды.

---

Данный перечень проектных решений фиксирует ключевые принципы, заложенные в архитектуру микросервиса. Они обеспечивают баланс между простотой MVP, надёжностью и возможностью дальнейшего масштабирования.

