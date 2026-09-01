# OpenSpec Implementation Tasks: T4.7 Atomic HLS Stream Directory Relocation Mechanics

- [x] Добавить операцию `MoveDir` в интерфейсный контракт хранилища `FileStorage`.
- [x] Реализовать атомарный перенос директорий на базе вызова `os.Rename` в драйвере `LocalStorage`.
- [x] Переработать `ProcessVideoUseCase` для изоляции процесса генерации потоков во временном каталоге `tmp/hls/`.
- [x] Интегрировать автоматическую очистку мусорных временных сегментов в сценарии компенсации сбоев (`fail`).
- [x] Добавить модульный тест `TestLocalStorage_MoveDir_Success` для проверки атомарного перемещения папок в тестовом окружении.
