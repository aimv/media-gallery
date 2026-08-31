# OpenSpec Change Proposal: T4.3.1 Integration Test Database Environment Isolation via Dedicated Container

## Objective
Изоляция окружения интеграционного тестирования от основной базы данных приложения. Решение предусматривает развертывание независимого контейнера СУБД `postgres_test` на выделенном порту `5433` для полной изоляции тестовых транзакций и исключения деструктивного влияния тестов на рабочую БД.

## Scope
- Расширение конфигурации `docker-compose.yml` новым изолированным сервисом `postgres_test`.
- Модификация `internal/infrastructure/config/config.go` для динамической сборки `TestDBDSN` с привязкой к порту `5433`.
- Перевод пулов соединений в интеграционных тестах пакета `postgres` на тестовый DSN.
