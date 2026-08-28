# OpenSpec Implementation Tasks: T3.9 Implement PostgreSQL Integration Tests for MediaRepository

- [x] Реализовать инициализацию тестового пула соединений на основе параметров среды `.env`.
- [x] Написать интеграционные тесты для методов `Save`, `FindByID`, `UpdateStatus`, `UpdateMetadata` с использованием `ROLLBACK`.
- [x] Обеспечить успешный зеленый прогон всего тестового пакета через `make test`.
