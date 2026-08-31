# OpenSpec Change Proposal: T4.2 PostgreSQL Queue Integration Testing

## Objective
Разработка сквозных интеграционных тестов для компонента `QueueRepository` для проверки атомарного захвата задач и автоматического восстановления зависших тасков при истечении времени аренды.

## Scope
- Создание тестового файла `internal/infrastructure/persistence/postgres/queue_repository_integration_test.go`.
