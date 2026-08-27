# OpenSpec Change Proposal: T3.6 Implement HTTP API Media Upload Handler

## Objective
Реализация транспортного инфраструктурного компонента (HTTP Controller) для приёма медиафайлов через Multipart Form data, маппинга ошибок и возврата strict JSON-ответов.

## Scope
- Создание контроллера `internal/infrastructure/delivery/http/handlers/media.go`.
- Написание транспортных тестов через `httptest` в `internal/infrastructure/delivery/http/handlers/media_test.go`.
