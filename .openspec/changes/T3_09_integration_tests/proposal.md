# OpenSpec Change Proposal: T3.9 Implement PostgreSQL Integration Tests for MediaRepository

## Objective
Реализация сквозных интеграционных тестов для инфраструктурного адаптера `MediaRepository` для верификации консистентности сырых SQL-запросов с использованием механизма автоматического отката транзакций.

## Scope
- Создание тестового файла `internal/infrastructure/persistence/postgres/media_repository_integration_test.go`.
