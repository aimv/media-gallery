# OpenSpec Implementation Tasks: T4.3.1 Integration Test Database Environment Isolation via Dedicated Container

- [x] Добавить изолированный сервис `postgres_test` (порт 5433) в инфраструктурный манифест `docker-compose.yml`.
- [x] Расширить структуру конфигурации параметром `TestDBDSN` в слое `internal/infrastructure/config`.
- [x] Модифицировать интеграционный тест `media_repository_integration_test.go` на использование пула `cfg.TestDBDSN`.
- [x] Модифицировать интеграционный тест `queue_repository_integration_test.go` на использование пула `cfg.TestDBDSN`.
- [x] Проверить успешное прохождение пайплайна локальных тестов через `make test`.
