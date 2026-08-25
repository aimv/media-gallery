# OpenSpec Change Proposal: T2.6 PostgreSQL Connection Pool Initializer

## Objective
Реализация инфраструктурного компонента управления пулом соединений к СУБД PostgreSQL на базе промышленной библиотеки `pgx/v5/pgxpool` с поддержкой автоматической валидации связи (Ping) и конфигурации таймаутов рантайма.

## Scope
- Создание файла `internal/infrastructure/persistence/postgres/db.go`.
- Реализация функции инициализации пула `NewPool`.
