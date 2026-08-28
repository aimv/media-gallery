# OpenSpec Change Proposal: T3.8 Expand UseCase Unit Tests and Verify Coverage

## Objective
Расширение тестовых сценариев `UploadUseCase` для полной верификации лимитов размера файлов и физическое обеспечение метрики покрытия кода (Test Coverage) выше целевого порога в 80%.

## Scope
- Модификация тестового файла `internal/usecase/media/upload_test.go`.
- Профайлинг покрытия через системный вызов `go tool cover`.
