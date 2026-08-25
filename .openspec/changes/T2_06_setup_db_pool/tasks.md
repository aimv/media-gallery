# OpenSpec Implementation Tasks: T2.6 PostgreSQL Connection Pool Initializer

- [x] Создать файл `internal/infrastructure/persistence/postgres/db.go` с документирующим комментарием пакета.
- [x] Реализовать инициализацию `pgxpool.NewWithConfig` с парсингом строки подключения DSN.
- [x] Добавить вызов блокирующего метода `pool.Ping(ctx)` для верификации физической доступности СУБД при инициализации.
- [x] Проверить успешное прохождение компиляции и статического анализа через команду `make lint`.
