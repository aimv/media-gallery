# OpenSpec Change Proposal: T1.1 Init Layout & Go Module

## Objective
Инициализация базового Go-модуля проекта и физическое создание скелета каталогов по стандартам Clean Architecture на основе утвержденного системного дизайна.

## Scope
- Инициализация модуля `go mod init media-gallery`.
- Создание корневых директорий: `cmd/api`, `cmd/worker`, `internal/domain`, `internal/usecase`, `internal/infrastructure`, `internal/pkg`.
