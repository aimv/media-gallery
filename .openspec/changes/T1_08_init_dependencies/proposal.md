# OpenSpec Change Proposal: T1.8 Initialize External Project Dependencies

## Objective
Декларативное добавление и фиксация версий внешних библиотек общего назначения в конфигурационном файле `go.mod` для последующей реализации слоев UseCase и Infrastructure.

## Scope
- Интеграция библиотек: `pgx/v5` (PostgreSQL), `chi/v5` (HTTP REST routing), `golang-migrate`, `godotenv`, `uuid`.
- Выполнение очистки и верификации графа зависимостей через `go mod tidy`.
