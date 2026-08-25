# OpenSpec Change Proposal: T2.1 Setup PostgreSQL 16 Container

## Objective
Развертывание изолированного контейнера СУБД PostgreSQL 16 через Docker Compose для локального хранения метаданных и организации конкурентной очереди задач SKIP LOCKED.

## Scope
- Создание базового файла `docker-compose.yml` в корневом каталоге.
- Настройка переменных окружения СУБД и постоянного тома (volume) для данных.
- Конфигурация механизма healthcheck с использованием утилиты `pg_isready`.
