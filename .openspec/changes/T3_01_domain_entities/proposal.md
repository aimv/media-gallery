# OpenSpec Change Proposal: T3.1 Implement Media Domain Entities and Status Validation

## Objective
Реализация базовых сущностей ядра бизнес-логики (Domain Layer) для управления медиа-ассетами, описание типизированных перечислений (ENUM) и логики валидации переходов состояний медиафайлов.

## Scope
- Создание файлов `internal/domain/entity/media_asset.go` и `internal/domain/entity/enums.go`.
- Написание юнит-тестов для верификации конечного автомата статусов в `internal/domain/entity/media_asset_test.go`.
