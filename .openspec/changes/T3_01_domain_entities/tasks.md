# OpenSpec Implementation Tasks: T3.1 Implement Media Domain Entities and Status Validation

- [x] Описать доменные типы `MediaType` и `MediaStatus` с методами строковой конвертации.
- [x] Реализовать структуру `MediaAsset` со всеми метаданными из системного технического дизайна.
- [x] Написать метод валидации смены состояний `CanTransitionTo(next MediaStatus) bool` для защиты от race conditions.
- [x] Покрыть логику валидации переходов состояний изолированными Юнит-тестами в файле `media_asset_test.go`.
- [x] Обеспечить успешное выполнение тестов через системную команду `make test`.
