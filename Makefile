.PHONY: help run build test lint docker-up docker-down migrate clean deps

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

run: ## Run application
	go run cmd/main.go

build: ## Build binary
	go build -o bin/app cmd/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

lint: ## Run linter
	golangci-lint run

docker-up: ## Start infrastructure
	docker-compose up -d

docker-down: ## Stop infrastructure
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

migrate: ## Run migrations
	RUN_MIGRATIONS=true go run cmd/main.go

deps: ## Download dependencies
	go mod download
	go mod tidy

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out

dev: docker-up migrate run ## Start development environment

.DEFAULT_GOAL := help
