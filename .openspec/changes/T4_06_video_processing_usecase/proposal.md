# OpenSpec Change Proposal: T4.6 Video Processing UseCase and Asynchronous Lease Heartbeat Integration

## Objective
Разработка координирующего Use Case компонента `ProcessVideoUseCase` для оркестрации шагов конвертации видеофайлов в формат HLS. Задача включает интеграцию конкурентного фонового механизма удержания блокировок (`Heartbeat`) через распределенное продление времени жизни аренды задачи в БД во время работы утилиты `ffmpeg`.

## Scope
- Расширение интерфейсного контракта доменного порта `QueueRepository` методами `UpdateStatus` и `ExtendLease`.
- Создание Use Case бизнес-логики в файле `internal/usecase/media/process_video.go`.
- Реализация фонового тикера в горутине для периодического обновления `lease_expires_at` активных задач.
- Покрытие бизнес-сценариев модульными тестами в `internal/usecase/media/process_video_test.go`.
