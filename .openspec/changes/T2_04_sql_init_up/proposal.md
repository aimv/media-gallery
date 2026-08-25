# OpenSpec Change Proposal: T2.4 Database Schema Initialization (Up)

## Objective
Создание стартового SQL-скрипта миграции `0001_init.up.sql` для развертывания реляционной структуры таблиц, ENUM-типов и специализированных индексов СУБД PostgreSQL.

## Scope
- Создание файла миграции `internal/infrastructure/persistence/postgres/migrations/0001_init.up.sql`.
- Описание DDL для таблиц: `media_assets`, `processing_jobs`, `content_blocks`, `block_media_links`.

## Non-Goals
- Написание логики отката схемы (перенесено на шаг T2.5).
- Перевод миграций на технологию `go:embed` (зафиксировано в коде как техдолг для будущих итераций системы).