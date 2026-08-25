# OpenSpec Implementation Tasks: T2.1 Setup PostgreSQL 16 Container

- [x] Создать файл `docker-compose.yml` с описанием сервиса базы данных `postgres`.
- [x] Прописать параметры авторизации, порты `5432:5432` и монтирование именованного волюма `postgres_data`.
- [x] Реализовать декларативную проверку готовности БД (`healthcheck`) через интервалы в 5 секунд.
