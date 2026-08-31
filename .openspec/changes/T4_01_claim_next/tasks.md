# OpenSpec Implementation Tasks: T4.1 PostgreSQL Queue ClaimNext Implementation

- [x] Создать доменную структуру `ProcessingJob` и типы состояний `JobStatus` в пакете сущностей.
- [x] Декларировать абстрактный порт `QueueRepository` в слое доменных контрактов.
- [x] Реализовать SQL-метод захвата задач `ClaimNext` с проставлением блокировки аренды и защитой от зависших тасков.
- [x] Убедиться в успешном прохождении статического анализа и компиляции через `make lint`.