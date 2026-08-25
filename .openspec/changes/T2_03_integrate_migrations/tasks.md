# OpenSpec Implementation Tasks: T2.3 Integrate Automated Migrations Engine

- [x] Реализовать служебную функцию `RunMigrations(dbDSN, migrationsDir string) error` в пакете инфраструктуры.
- [x] Создать минимальные точки входа `cmd/api/main.go` и `cmd/worker/main.go` с вызовом загрузки конфига, логгера и мигратора.
- [x] Проверить успешное прохождение компиляции и статического анализа `make lint`.
