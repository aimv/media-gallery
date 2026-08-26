# OpenSpec Implementation Tasks: T3.4 Implement PostgreSQL MediaRepository Adapter

- [x] Реализовать SQL-методы `Save`, `FindByID`, `UpdateStatus` и `UpdateMetadata`.
- [x] Интегрировать маппинг доменной ошибки `apperror.ErrNotFound` на основе системного флага `pgx.ErrNoRows`.
- [x] Убедиться в прохождении компиляции, линтинга и базовых тестов пакета через `make test`.
