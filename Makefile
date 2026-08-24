.PHONY: build test lint run-api run-worker

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
