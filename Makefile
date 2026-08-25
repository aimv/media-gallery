.PHONY: build test lint run-api run-worker

# Автоматически подгружаем переменные из локального .env файла
ifneq ($(wildcard .env),)
    include .env
    export
endif

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test -v -race ./...

lint:
	$(shell go env GOPATH)/bin/golangci-lint run

run-api:
	go run ./cmd/api/main.go

run-worker:
	go run ./cmd/worker/main.go

migrate-up:
	$(shell go env GOPATH)/bin/migrate -path internal/infrastructure/persistence/postgres/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

migrate-down:
	$(shell go env GOPATH)/bin/migrate -path internal/infrastructure/persistence/postgres/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down 1

migrate-down-all:
	$(shell go env GOPATH)/bin/migrate -path internal/infrastructure/persistence/postgres/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down -all