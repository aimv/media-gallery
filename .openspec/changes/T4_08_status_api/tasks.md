# OpenSpec Implementation Tasks: T4.8 Media Asset Processing Status Query HTTP API End-point

- [x] Создать Use Case `GetStatusUseCase` в бизнес-пакете `internal/usecase/media`.
- [x] Реализовать метод `GetStatus` в контроллере `MediaHandler` с валидацией UUID в Path-параметрах.
- [x] Добавить логику динамической сборки относительного URL-адреса HLS-плейлиста для ресурсов в статусе `ready`.
- [x] Обеспечить унифицированную JSON-сериализацию доменных ошибок и исключить plain-text ответы.
- [x] Разработать модульные тесты `TestMediaHandler_GetStatus_Success` и `TestMediaHandler_GetStatus_NotFound` для проверки HTTP-слоя.
