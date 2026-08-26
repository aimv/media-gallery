# OpenSpec Change Proposal: T3.3 Implement Local Disk FileStorage Adapter

## Objective
Реализация инфраструктурного адаптера `LocalStorage`, соответствующего доменному интерфейсу `FileStorage`, для потокового сохранения бинарных данных на диск с обязательной валидацией контента по сигнатурам (Magic Bytes).

## Scope
- Создание адаптера `internal/infrastructure/storage/local.go`.
- Написание модульных тестов `internal/infrastructure/storage/local_test.go` с использованием временных каталогов `t.TempDir()`.
