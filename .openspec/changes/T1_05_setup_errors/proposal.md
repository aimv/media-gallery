# OpenSpec Change Proposal: T1.5 Setup Domain Error Package

## Objective
Реализация централизованного пакета обработки ошибок `apperror` для типизации доменных и прикладных сбоев, содержащего уникальные коды ошибок и соответствующие HTTP-статусы ответов.

## Scope
- Создание пакета `internal/pkg/apperror/`.
- Объявление структуры `AppError` и конструкторов для типов ошибок: `ErrNotFound`, `ErrInvalidInput`, `ErrConflict`, `ErrInternal`.
