# OpenSpec Change Proposal: Global Architecture Design & ADR-0001

## Objective
Разработка комплексного технического дизайна микросервиса медиагалереи на Go 1.26 по стандартам Clean Architecture / DDD и детальное проектирование реляционной схемы СУБД PostgreSQL. Фиксация ключевых проектных компромиссов в виде документов решений (ADR) и сквозного архитектурного обоснования.

## Scope
- Проектирование и визуализация дерева каталогов (`cmd/`, `internal/domain`, `internal/usecase`, `internal/infrastructure`).
- Спецификация контрактов (Go-интерфейсов) для портов слоев: `MediaRepository`, `JobQueue`, `VideoProcessor`, `FileStorage`, `CMSRepository`.
- Описание DDL-схемы таблиц базы данных (`media_assets`, `processing_jobs`, `content_blocks`, `block_media_links`) со всеми индексами, внешними ключами и ограничениями.
- Добавление раздела обоснований ключевых проектных решений (UUIDv4, `io.Reader` для потоковой обработки, Soft Delete).
- Составление документа архитектурного решения `docs/adr/0001-use-postgres-skip-locked-queue.md` по методологии MADR.

## Non-Goals
- Написание исполняемого Go-кода (реализации репозиториев, хендлеров и воркеров).
- Декомпозиция бэклога спринтов на атомарные подзадачи.
