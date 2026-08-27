# OpenSpec Implementation Tasks: T3.6 Implement HTTP API Media Upload Handler

- [x] Реализовать хендлер `Upload` с парсингом Multipart Form данных (`FormFile`).
- [x] Интегрировать strict JSON вывод ошибок через кастомный маршалинг без использования `http.Error`.
- [x] Написать модульные тесты хендлера с эмуляцией HTTP-запросов загрузки файлов через `httptest.NewRecorder`.
- [x] Убедиться в успешном прохождении статического анализа и тестов через команду `make test`.
