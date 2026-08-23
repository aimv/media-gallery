# OpenSpec Implementation Tasks: Architecture Design & Database Schema

- [x] Спроектировать целевую структуру папок Go-монолита по стандартам Clean Architecture.
- [x] Описать чистые Go-интерфейсы портов инфраструктуры для репозиториев и внешних сервисов.
- [x] Разработать реляционную схему таблиц PostgreSQL с учетом связей многие-ко-многим и индексов под SKIP LOCKED.
- [x] Подготовить раздел обоснований проектных компромиссов (UUIDv4 vs Serial, io.Reader stream, Soft Delete).
- [x] Сформировать архитектурный документ решений ADR-0001 в формате Context-Decision-Consequences для нативной очереди СУБД.
- [x] Создать результирующие файлы проектной документации `docs/2_DESIGN.md` и `docs/adr/0001-use-postgres-skip-locked-queue.md`.
