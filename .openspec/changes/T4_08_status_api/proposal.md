# OpenSpec Change Proposal: T4.8 Media Asset Processing Status Query HTTP API End-point

## Objective
Разработка сервисного слоя бизнес-логики и HTTP-интерфейса доставки запросов получения статуса асинхронной обработки медиаресурсов (`GET /api/media/:id`). Эндлер обеспечивает клиентов информацией о текущем состоянии задачи (`uploaded`, `processing`, `ready`, `failed`), техническими метаданными контейнера и динамической ссылкой на манифест HLS потока.

## Scope
- Создание Use Case компонента `internal/usecase/media/get_status.go`.
- Расширение методов контроллера `MediaHandler` функцией `GetStatus` с разбором Path-параметров роутинга Go 1.22+.
- Реализация JSON-сериализации ответов и типизированного маппинга ошибок через `apperror` без plain-text утечек.
- Написание покрывающих юнит-тестов HTTP-ответов в `media_test.go`.
