# OpenSpec Implementation Tasks: T1.5 Setup Domain Error Package

- [x] Создать файл `internal/pkg/apperror/errors.go` с документирующим комментарием пакета.
- [x] Реализовать кастомный тип `AppError`, удовлетворяющий стандартному интерфейсу `error`.
- [x] Объявить предопределенные экземпляры доменных ошибок для маппинга (NotFound, InvalidInput, Conflict, Internal).
- [x] Проверить успешное прохождение статического анализа через команду `make lint`.
