# OpenSpec Change Proposal: T4.3 VideoProcessor Configurable Contract and FFmpeg Implementation with Context Testing

## Objective
Внедрение конфигурируемого доменного интерфейса `VideoProcessor` и его инфраструктурного адаптера `FFmpegProcessor`. Расширение конфигурации приложения для контроля потоков кодирования и написание модульных тестов для верификации отмены процессов по сигналам контекста (DoD).

## Scope
- Расширение структуры `config.Config` полем `FFMPEGThreads` (`FFMPEG_THREADS`).
- Создание доменного интерфейса `internal/domain/service/video_processor.go`.
- Создание реализации `internal/infrastructure/video/ffmpeg_processor.go`.
- Создание модульных тестов контроля жизненного цикла процессов `internal/infrastructure/video/ffmpeg_processor_test.go`.
