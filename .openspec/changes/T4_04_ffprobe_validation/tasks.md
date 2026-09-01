# OpenSpec Implementation Tasks: T4.4 Video Validation Verification via ffprobe Mock Testing

- [x] Реализовать модульный тест `TestFFmpegProcessor_ProbeMetadata_Success` с имитацией JSON-вывода потоков.
- [x] Реализовать модульный тест `TestFFmpegProcessor_Validate_InvalidData` с контролем обработки поврежденных контейнеров.
- [x] Обеспечить строгое соответствие возвращаемых кодов ошибок спецификации `apperror.ErrInvalidInput`.
- [x] Успешно верифицировать сборку всего тестового покрытия через команду `make test`.
