# OpenSpec Change Proposal: T3.4 Implement PostgreSQL MediaRepository Adapter

## Objective
Реализация инфраструктурного адаптера данных `MediaRepository` на базе драйвера `pgx/v5` для сохранения, поиска и изменения метаданных ассетов в СУБД PostgreSQL.

## Scope
- Создание адаптера `internal/infrastructure/persistence/postgres/media_repository.go`.
- Создание модульных тестов-заглушек верификации типов `internal/infrastructure/persistence/postgres/media_repository_test.go`.
