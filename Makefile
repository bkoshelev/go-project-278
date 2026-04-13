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
