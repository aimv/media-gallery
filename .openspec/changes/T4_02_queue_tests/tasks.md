# OpenSpec Implementation Tasks: T4.2 PostgreSQL Queue Integration Testing

- [x] Реализовать интеграционный тест `TestIntegration_QueueRepository_Lifecycle` для проверки методов захвата задач.
- [x] Написать сценарий симуляции зависшей задачи с истекшим `lease_expires_at` и верифицировать её перехват.
- [x] Обеспечить успешное выполнение всего тестового пакета через `make test`.
