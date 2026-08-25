# OpenSpec Change Proposal: T2.5 Database Schema Rollback (Down)

## Objective
Создание зеркального SQL-скрипта миграции `0001_init.down.sql` для корректного деструктивного удаления реляционных таблиц, кастомных ENUM-типов и индексов СУБД PostgreSQL в обратной последовательности.

## Scope
- Создание файла миграции `internal/infrastructure/persistence/postgres/migrations/0001_init.down.sql`.
- Описание каскадного безопасного удаления структуры таблиц и типов.
