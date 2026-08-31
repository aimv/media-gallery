# OpenSpec Implementation Tasks: T4.3 VideoProcessor Configurable Contract and FFmpeg Implementation with Context Testing

- [x] Добавить параметр `FFMPEG_THREADS` в конфигурационный слой `internal/infrastructure/config`.
- [x] Создать интерфейс `VideoProcessor` и DTO `VideoMetadata` в пакете `internal/domain/service`.
- [x] Реализовать методы `ProbeMetadata` и `Validate` с разбором вывода утилиты `ffprobe`.
- [x] Реализовать метод `ProcessToHLS` с динамической передачей флага `-threads`.
- [x] Написать покрывающие тесты для проверки корректного завершения внешних процессов при принудительной отмене `context.Context`.
