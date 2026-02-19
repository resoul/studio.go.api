.PHONY: help build run migrate test test-coverage lint fmt tidy clean dev install-tools

GOCACHE ?= /tmp/go-build

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build binary to ./bin/api
	go build -o ./bin/api .

run: ## Run API server
	go run main.go serve

migrate: ## Run database migrations
	go run main.go migrate

test: ## Run tests
	GOCACHE=$(GOCACHE) go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Open coverage report
	go tool cover -html=coverage.out

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format Go code
	go fmt ./...
	gofumpt -l -w .

tidy: ## Tidy go modules
	go mod tidy

clean: ## Remove build artifacts
	rm -rf ./bin
	rm -f coverage.out

docker-build: ## Build docker image
	docker build -t senderscore-api:latest .

docker-run: ## Start docker compose
	docker-compose up -d

docker-stop: ## Stop docker compose
	docker-compose down

dev: ## Run with air hot-reload
	air -c air.api.toml

install-tools: ## Install local dev tools
	go install github.com/cosmtrek/air@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.DEFAULT_GOAL := help
