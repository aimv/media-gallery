# OpenSpec Implementation Tasks: T2.5 Database Schema Rollback (Down)

- [x] Создать файл DDL-миграции `0001_init.down.sql` в каталоге миграций.
- [x] Прописать операторы `DROP TABLE IF EXISTS` в строгой обратной последовательности для сохранения целостности внешних ключей.
- [x] Описать деструктивное удаление кастомных ENUM-типов данных.
