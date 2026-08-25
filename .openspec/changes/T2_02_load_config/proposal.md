# OpenSpec Change Proposal: T2.2 Environment Configuration Loader

## Objective
Реализация инфраструктурного компонента загрузки конфигурации из файлов среды `.env` для обеспечения принципов методологии 12-Factor App и изоляции секретов.

## Scope
- Создание пакета `internal/infrastructure/config/`.
- Реализация структуры `Config` для хранения параметров СУБД, HTTP-сервера, API-ключей и лимитов хранилища.
- Интеграция библиотеки `godotenv`.
