dev:
	~/go/bin/air

tidy: ## Tidy up dependencies, format code, and run vet
	go mod tidy
	go fmt ./...
	go vet ./...
