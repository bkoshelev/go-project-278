include .env

GOPATH := $(shell go env GOPATH)

dev:
	~/go/bin/air

build:
	go build -o bin/shorturl ./main.go

tidy: ## Tidy up dependencies, format code, and run vet
	go mod tidy
	go fmt ./...
	go vet ./...

lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix

test:
	go test ./...

generate:
	$(GOPATH)/bin/sqlc generate

migrate:
	$(GOPATH)/bin/goose --dir ./internal/db/migrations postgres $(DATABASE_URL) up
