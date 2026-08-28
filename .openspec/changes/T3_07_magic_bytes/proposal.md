# OpenSpec Change Proposal: T3.7 Isolate Magic Bytes Content Type Detection

## Objective
Вынесение логики валидации и распознавания реального типа контента по сигнатурам (Magic Bytes) в изолированную служебную функцию общего назначения для переиспользования в слоях системы.

## Scope
- Создание утилиты `internal/domain/entity/validator.go` (или расширение пакета сущностей).
- Рефакторинг `LocalStorage` для использования новой функции.
