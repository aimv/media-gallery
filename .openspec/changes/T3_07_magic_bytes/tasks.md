# OpenSpec Implementation Tasks: T3.7 Isolate Magic Bytes Content Type Detection

- [x] Реализовать функцию `DetectContentType(io.Reader) (MediaType, error)` в доменном слое.
- [x] Добавить явный возврат `apperror.ErrInvalidInput` для неподдерживаемых сигнатур.
- [x] Обновить код `LocalStorage`, удалив дублирующий захардкоженный switch.
- [x] Проверить успешность сборки, линтинга и прогона тестов через `make test`.
