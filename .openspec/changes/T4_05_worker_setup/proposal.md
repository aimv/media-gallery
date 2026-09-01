# OpenSpec Change Proposal: T4.5 Async Worker Entrypoint Core Infrastructure Setup

## Objective
Развертывание масштабируемой инфраструктуры асинхронного фонового процесса в `cmd/worker/main.go`. Задача включает внедрение пула параллельных горутин-воркеров, конфигурируемых из переменных среды, интеграцию механизмов Graceful Shutdown и атомарный опрос распределенной очереди задач PostgreSQL.

## Scope
- Модификация `internal/infrastructure/config/config.go` для поддержки параметра `WorkerPoolSize`.
- Переработка точки входа `cmd/worker/main.go` с интеграцией `sync.WaitGroup`, сигнальной обработки `os/signal` и конкурентных вызовов `ClaimNext`.
