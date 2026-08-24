# Декомпозиция бэклога задач (Task Decomposition)

Документ представляет детальный бэклог атомарных микро-задач для реализации микросервиса медиагалереи с CMS. Задачи сгруппированы по спринтам и соответствуют архитектуре из `docs/2_DESIGN.md`. Каждая задача изолирована, затрагивает 1–2 файла и занимает не более 200–300 строк кода, что обеспечивает плавную историю коммитов.

---

## Спринт 1: Инфраструктурный каркас и контроль качества

- [x] **T1.1: Инициализация go.mod и структуры каталогов**  
  Создать файл `go.mod` с именем модуля `media-gallery`. Создать дерево каталогов в соответствии с разделом 2 документа `docs/2_DESIGN.md`: `cmd/api`, `cmd/worker`, `internal/domain`, `internal/usecase`, `internal/infrastructure`, `internal/pkg`.  
  **DoD:** Проект компилируется без ошибок (`go build ./...`). Структура папок совпадает с дизайном.

- [x] **T1.2: Настройка Makefile**  
  Создать `Makefile` с целями: `build` (сборка API и worker), `test` (запуск тестов), `lint` (запуск golangci-lint), `run-api`, `run-worker`, `migrate-up`, `migrate-down`.  
  **DoD:** Команда `make build` успешно собирает оба бинарных файла.

- [x] **T1.3: Создание .golangci.yml**  
  Добавить конфигурацию линтера `golangci-lint` с включёнными линтерами: `govet`, `errcheck`, `staticcheck`, `ineffassign`, `misspell`, `gocyclo` (порог 15), `revive`.  
  **DoD:** `golangci-lint run` не выявляет ошибок в пустом проекте.

- [x] **T1.4: Реализация структурированного логгера на slog**  
  Создать пакет `internal/pkg/logger`, оборачивающий стандартный `log/slog` с настройкой уровня, формата (JSON/text) и добавлением request ID.  
  **DoD:** Логгер инициализируется из конфигурации и пишет в stdout.

- [x] **T1.5: Создание пакета ошибок**  
  Реализовать `internal/pkg/apperror` с типами: `AppError` (код, сообщение, HTTP-статус), предопределённые ошибки (`ErrNotFound`, `ErrInvalidInput`, `ErrConflict` и т.д.).  
  **DoD:** Ошибки имеют машиночитаемые коды и человекочитаемые сообщения.

- [x] **T1.6: Настройка CI/CD GitHub Actions**  
  Создать `.github/workflows/ci.yml` с шагами: checkout, установка Go, запуск `go vet`, `golangci-lint`, `go test ./...`.  
  **DoD:** Workflow запускается при push и pull request, все шаги проходят.

- [ ] **T1.7: Добавление AI-инфраструктуры (CLAUDE.md, .kilo/)**  
  Создать файлы `CLAUDE.md` (инструкции для ИИ-агента) и каталог `.kilo/` с конфигурацией для генерации кода (опционально, если используется Kilo).  
  **DoD:** Репозиторий содержит базовые инструкции для будущей AI-разработки.

- [ ] **T1.8: Инициализация зависимостей (go.mod)**  
  Добавить основные библиотеки: `github.com/jackc/pgx/v5` (драйвер PostgreSQL), `github.com/go-chi/chi/v5` (роутер), `github.com/golang-migrate/migrate/v4`, `github.com/joho/godotenv`, `github.com/google/uuid`.  
  **DoD:** `go mod tidy` выполняется без конфликтов.

---

## Спринт 2: Persistence Слой и СУБД

- [ ] **T2.1: Docker Compose для PostgreSQL 16**  
  Создать `docker-compose.yml` с сервисом `postgres` (образ `postgres:16-alpine`), environment (POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB), volume для данных и healthcheck (`pg_isready`).  
  **DoD:** `docker compose up postgres` поднимает БД, healthcheck проходит.

- [ ] **T2.2: Загрузка конфигурации из .env**  
  Реализовать пакет `internal/infrastructure/config`, который с помощью `godotenv` загружает `.env` и парсит переменные окружения в структуру `Config` (DSN, порты, пути, лимиты).  
  **DoD:** Конфигурация доступна в main, значения по умолчанию заданы.

- [ ] **T2.3: Интеграция golang-migrate**  
  В `cmd/api/main.go` и `cmd/worker/main.go` добавить вызов применения миграций из директории `internal/infrastructure/persistence/postgres/migrations/`.  
  **DoD:** При старте приложения миграции автоматически накатываются.

- [ ] **T2.4: SQL-миграция 0001_init.up.sql**  
  Создать файл миграции `0001_init.up.sql` со всеми таблицами, enum-типами, индексами и внешними ключами согласно разделу 4 документа `docs/2_DESIGN.md`: `media_assets`, `processing_jobs`, `content_blocks`, `block_media_links`.  
  **DoD:** SQL-скрипт выполняется без ошибок, структура БД соответствует дизайну.

- [ ] **T2.5: SQL-миграция 0001_init.down.sql**  
  Создать обратную миграцию `0001_init.down.sql`, удаляющую все созданные таблицы и enum-типы.  
  **DoD:** `migrate down` корректно откатывает схему.

- [ ] **T2.6: Пул соединений с PostgreSQL**  
  Реализовать `internal/infrastructure/persistence/postgres/db.go` с функцией `NewPool` (используя `pgxpool`), настройкой таймаутов и проверкой соединения.  
  **DoD:** Приложение устанавливает соединение с БД и выполняет `SELECT 1`.

---

## Спринт 3: Модуль загрузки медиафайлов

- [ ] **T3.1: Domain-сущности MediaAsset, MediaStatus, MediaType**  
  В `internal/domain/entity/media_asset.go` и `enums.go` определить структуры `MediaAsset`, `MediaType` (image/jpeg, image/png, video/mp4), `MediaStatus` (uploaded, queued, processing, ready, failed, deleting) и методы перехода статусов.  
  **DoD:** Определены все поля из `docs/2_DESIGN.md`, есть валидация допустимых значений.

- [ ] **T3.2: Интерфейсы MediaRepository и FileStorage**  
  В `internal/domain/repository/media_repository.go` и `internal/domain/service/file_storage.go` объявить интерфейсы с сигнатурами методов (Create, GetByID, List, UpdateStatus и т.д.; Save, Open, Delete, Move, Exists, EnsureDir).  
  **DoD:** Интерфейсы компилируются, domain не импортирует внешние пакеты.

- [ ] **T3.3: Реализация FileStorage для локального диска**  
  Создать `internal/infrastructure/filestorage/local_fs.go` с методами, использующими стандартную библиотеку `os` и `io`. Реализовать потоковое копирование через `io.Copy`, атомарное перемещение (`os.Rename`).  
  **DoD:** Файлы сохраняются, перемещаются, удаляются; пути защищены от path traversal.

- [ ] **T3.4: Реализация MediaRepository для PostgreSQL**  
  Создать `internal/infrastructure/persistence/postgres/media_repository_pg.go` с методами, выполняющими SQL-запросы через `pgxpool`. Включить маппинг строк БД в domain-сущности.  
  **DoD:** Все CRUD-операции работают с БД, тесты на интеграцию проходят.

- [ ] **T3.5: UseCase загрузки медиа**  
  Реализовать `internal/usecase/media/upload.go` с функцией `Upload(ctx, file io.Reader, filename, contentType string, size int64)`. Логика: валидация (расширение, Content-Type, magic bytes), сохранение файла через FileStorage, создание записи в БД, постановка задачи в очередь (для видео).  
  **DoD:** UseCase вызывает порты, не зависит от HTTP.

- [ ] **T3.6: HTTP-хендлер POST /api/v1/media/upload**  
  Создать `internal/infrastructure/http/handler/media_upload.go`, который парсит multipart-форму, ограничивает размер тела, вызывает usecase и возвращает JSON с данными медиафайла.  
  **DoD:** Эндпоинт принимает файл, сохраняет, возвращает 201 и JSON с ID, статусом.

- [ ] **T3.7: Валидация magic bytes**  
  В составе пакета загрузки реализовать функцию `detectContentType(reader io.Reader) (MediaType, error)`, использующую `http.DetectContentType` и сигнатуры JPEG/PNG/MP4.  
  **DoD:** Файл с неверной сигнатурой отклоняется с ошибкой `ErrInvalidInput`.

- [ ] **T3.8: Юнит-тесты для валидации и usecase**  
  Написать тесты в `internal/usecase/media/upload_test.go`, используя моки FileStorage и MediaRepository. Проверить сценарии: валидный файл, неверный тип, превышение размера, ошибка сохранения.  
  **DoD:** Покрытие ключевых ветвей > 80%.

- [ ] **T3.9: Интеграционные тесты для MediaRepository**  
  В `internal/infrastructure/persistence/postgres/media_repository_pg_test.go` написать тесты, использующие тестовую БД (или Docker). Проверить Create, GetByID, UpdateStatus, List с фильтрами.  
  **DoD:** Тесты проходят при запущенном PostgreSQL.

---

## Спринт 4: Асинхронный воркер и интеграция ffmpeg

- [ ] **T4.1: Domain-сущности ProcessingJob, JobStatus**  
  Определить `processing_job.go` с полями (ID, AssetID, Status, Attempt, LeaseOwner, LeaseExpiresAt, Error, таймстемпы) и enum `JobStatus` (queued, processing, success, failed).  
  **DoD:** Сущность готова, статусы валидируются.

- [ ] **T4.2: Реализация JobQueue с SKIP LOCKED**  
  Создать `internal/infrastructure/persistence/postgres/job_queue_pg.go`, реализующий интерфейс `JobQueue`. Метод `ClaimNext` должен выполнять атомарный `UPDATE ... WHERE id = (SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING *`.  
  **DoD:** Метод возвращает задачу или nil, тесты подтверждают конкурентную безопасность.

- [ ] **T4.3: Реализация VideoProcessor (ffmpeg)**  
  Создать `internal/infrastructure/video/ffmpeg_processor.go` с методом `ProcessToHLS(ctx, inputPath, outputDir)`, использующим `exec.CommandContext` с аргументами ffmpeg для генерации HLS. Добавить ограничение `-threads`, таймаут.  
  **DoD:** Команда корректно запускается, при отмене контекста процесс завершается.

- [ ] **T4.4: Реализация ffprobe для валидации**  
  В `internal/infrastructure/video/ffprobe.go` реализовать функцию `ProbeMetadata` и `Validate`, вызывающие ffprobe и парсящие JSON-вывод.  
  **DoD:** Извлекаются длительность, разрешение, кодеки; невалидные файлы отвергаются.

- [ ] **T4.5: Создание воркера cmd/worker/main.go**  
  Реализовать запуск фонового воркера: инициализация логгера, конфигурации, БД, JobQueue, VideoProcessor; запуск пула горутин (количество из конфига), каждая из которых в цикле вызывает `ClaimNext` и обрабатывает задачу.  
  **DoD:** Worker стартует, читает очередь, запускает обработку.

- [ ] **T4.6: Реализация Heartbeat**  
  В воркере добавить горутину для периодического обновления `lease_expires_at` активных задач (каждые 30 секунд).  
  **DoD:** Lease продлевается, задачи не перехватываются другими воркерами.

- [ ] **T4.7: Обработка завершения задачи (Complete)**  
  Реализовать вызов `Complete` после завершения ffmpeg: при успехе — обновление `media_assets.status` на `ready`, `processing_jobs.status` на `success`; при ошибке — увеличение `attempt`, если < max_attempts, возврат в `queued`, иначе `failed`.  
  **DoD:** Логика ретраев работает согласно дизайну.

- [ ] **T4.8: Атомарное перемещение HLS**  
  После успешной генерации HLS во временную директорию выполнить `FileStorage.Move` в финальную `hls/{assetID}/`. Проверить целостность плейлиста.  
  **DoD:** Готовый HLS появляется только после полного копирования.

- [ ] **T4.9: Тесты JobQueue (race conditions)**  
  Написать интеграционный тест, запускающий несколько горутин, вызывающих `ClaimNext` параллельно, чтобы убедиться, что одна задача не выдаётся дважды.  
  **DoD:** Тест проходит многократно, не возникает дублирования.

- [ ] **T4.10: Тесты VideoProcessor (mock ffmpeg)**  
  Создать mock-скрипт ffmpeg, который создаёт фейковые HLS-файлы, и протестировать `ProcessToHLS` с ним. Проверить обработку ошибок и таймаута.  
  **DoD:** Процесс корректно завершается, ошибки обрабатываются.

---

## Спринт 5: CMS-модуль контентных блоков

- [ ] **T5.1: Domain-сущности ContentBlock, BlockMediaLink**  
  Определить `content_block.go` с полями (ID, PageKey, Type, Data, SortOrder) и `block_media_link.go`.  
  **DoD:** Типы блоков (`text`, `carousel`, `reviews`, `about`) зафиксированы.

- [ ] **T5.2: Реализация CMSRepository**  
  Создать `internal/infrastructure/persistence/postgres/cms_repository_pg.go` с методами CRUD для блоков, методами LinkMedia/UnlinkMedia, ListMediaByBlock, ListBlocksByMedia. Использовать транзакции для атомарности.  
  **DoD:** Все методы работают с БД, тесты интеграционные проходят.

- [ ] **T5.3: UseCase для CMS**  
  Реализовать `internal/usecase/cms/manage_blocks.go` с функциями CreateBlock, GetBlock, ListBlocks, UpdateBlock, DeleteBlock. Также `link_media.go` с проверкой статуса медиафайла.  
  **DoD:** UseCase вызывает репозиторий и выполняет бизнес-правила.

- [ ] **T5.4: HTTP-хендлеры для CMS**  
  Создать `internal/infrastructure/http/handler/cms.go` с эндпоинтами: `POST /api/v1/cms/blocks`, `GET /api/v1/cms/blocks/{id}`, `GET /api/v1/cms/pages/{pageKey}/blocks`, `PUT /api/v1/cms/blocks/{id}`, `DELETE /api/v1/cms/blocks/{id}`. Использовать `chi` для роутинга с path values (Go 1.22+).  
  **DoD:** CRUD работает, возвращаются корректные JSON.

- [ ] **T5.5: Эндпоинты привязки медиа**  
  Реализовать `POST /api/v1/cms/blocks/{blockID}/media` и `DELETE /api/v1/cms/blocks/{blockID}/media/{assetID}`.  
  **DoD:** Привязка и отвязка работают, проверяется существование блока и медиа.

- [ ] **T5.6: Валидация статуса ready перед привязкой**  
  В UseCase `LinkMedia` добавить проверку: если статус медиафайла не `ready`, возвращать ошибку `ErrConflict`.  
  **DoD:** Попытка привязать неготовое видео отклоняется.

- [ ] **T5.7: Тесты CMSRepository**  
  Написать интеграционные тесты для всех методов репозитория, включая транзакционные случаи.  
  **DoD:** Тесты проходят, включая сценарий удаления блока с каскадным удалением связей.

- [ ] **T5.8: Тесты UseCase CMS**  
  Написать юнит-тесты с моками репозитория, проверяющие бизнес-логику (например, запрет привязки неготового медиа).  
  **DoD:** Покрытие ключевых сценариев.

---

## Спринт 6: Финальная оркестрация и верификация

- [ ] **T6.1: Dockerfile.api**  
  Создать `Dockerfile.api` на основе `golang:1.26-alpine` (multi-stage): сборка бинарника `cmd/api`, установка `ffmpeg` (через `apk add ffmpeg`), финальный образ на `alpine`.  
  **DoD:** Образ собирается, контейнер запускается и слушает порт 8080.

- [ ] **T6.2: Dockerfile.worker**  
  Создать `Dockerfile.worker`, аналогичный Dockerfile.api, но с `cmd/worker`.  
  **DoD:** Контейнер worker запускается и подключается к БД.

- [ ] **T6.3: Финальный docker-compose.yml**  
  Обновить `docker-compose.yml`: добавить сервисы `api` и `worker`, использующие соответствующие Dockerfile, общий volume для файлового хранилища, зависимости от `postgres` (healthcheck), переменные окружения.  
  **DoD:** `docker compose up` поднимает все три сервиса, система работает.

- [ ] **T6.4: Postman-коллекция**  
  Создать `docs/postman/media-gallery.postman_collection.json` с примерами запросов: загрузка изображения/видео, получение статуса, получение HLS, CRUD CMS.  
  **DoD:** Коллекция импортируется в Postman, все запросы выполняются.

- [ ] **T6.5: README.md**  
  Написать инструкцию по локальному запуску через `docker compose`, описанием переменных окружения, шагами миграции, примерами API.  
  **DoD:** Новый разработчик может запустить проект по README за 10 минут.

- [ ] **T6.6: E2E-тестирование**  
  Выполнить ручной сценарий: загрузить MP4, дождаться статуса `ready`, запросить HLS, убедиться в корректном воспроизведении (например, через VLC).  
  **DoD:** Сценарий проходит без ошибок, логи не содержат критических ошибок.

- [ ] **T6.7: Добавление healthcheck endpoints**  
  Реализовать `GET /healthz` в API и воркере, возвращающий статус соединения с БД и хранилищем.  
  **DoD:** Healthcheck используется в docker-compose для проверки готовности.

- [ ] **T6.8: Финальная проверка линтера и тестов**  
  Запустить `golangci-lint run` и `go test ./...`, устранить все предупреждения.  
  **DoD:** Все проверки CI проходят, код готов к релизу.

---

## Итог

Общее количество задач: 49. Распределение по спринтам:

| Спринт | Количество задач |
|--------|------------------|
| 1      | 8                |
| 2      | 6                |
| 3      | 9                |
| 4      | 10               |
| 5      | 8                |
| 6      | 8                |

Каждая задача является атомарным шагом, что позволяет вести разработку инкрементально с частыми коммитами и контролем качества на каждом этапе.