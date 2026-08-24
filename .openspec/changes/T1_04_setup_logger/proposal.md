# OpenSpec Change Proposal: T1.4 Setup Structured Logger

## Objective
Реализация пакета структурированного логирования общего назначения на базе стандартной библиотеки `log/slog` для вывода логов приложения в формате JSON в stdout с поддержкой уровней логирования.

## Scope
- Создание пакета `internal/pkg/logger/`.
- Инициализация глобального JSON-логгера.
