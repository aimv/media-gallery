# OpenSpec Change Proposal: T2.3 Integrate Automated Migrations Engine

## Objective
Интеграция встроенного механизма автоматического применения SQL-миграций СУБД PostgreSQL на старте сервисов с использованием библиотеки `golang-migrate/v4`.

## Scope
- Реализация пакета мигратора `internal/infrastructure/persistence/postgres/migrator.go`.
- Создание заглушек точек входа `cmd/api/main.go` и `cmd/worker/main.go` для верификации автомиграций.
