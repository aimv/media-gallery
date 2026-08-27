# OpenSpec Change Proposal: T3.5 Implement Media Upload UseCase & Mocked Unit Tests

## Objective
Реализация прикладного уровня бизнес-логики (UseCase Layer) для координации и оркестрации процесса валидации, потокового сохранения файлов на диск и атомарной фиксации метаданных в СУБД.

## Scope
- Создание UseCase-компонента `internal/usecase/media/upload.go`.
- Написание изолированных табличных модульных тестов с рукописными моками интерфейсов в `internal/usecase/media/upload_test.go`.
