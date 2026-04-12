.PHONY: build run test lint fmt clean migrate-up migrate-down swagger docker-up docker-down help

# Load .env file if it exists
-include .env
export

# Variables
APP_NAME := zencial-api
MIGRATE_NAME := zencial-migrate
BUILD_DIR := bin
MAIN_API := cmd/api/main.go
MAIN_MIGRATE := cmd/migrate/main.go

# Go commands
GOTEST := go test
GOBUILD := go build
GORUN := go run
GOFMT := gofmt

# Build info
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
VERSION    := $(shell git describe --tags --exact-match 2>/dev/null || echo dev)
PKG        := github.com/zenfulcode/zencial/internal/infrastructure/buildinfo
LDFLAGS    := -s -w -X '$(PKG).Version=$(VERSION)' -X '$(PKG).Commit=$(GIT_COMMIT)' -X '$(PKG).BuildTime=$(BUILD_TIME)'

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/  /'

## build: Build the API binary
build:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_API)

## build-migrate: Build the migration binary
build-migrate:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(MIGRATE_NAME) $(MAIN_MIGRATE)

## run: Run the API server
run:
	$(GORUN) $(MAIN_API)

## test: Run all tests
test:
	$(GOTEST) -v -race -count=1 ./...

## test-cover: Run tests with coverage
test-cover:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	$(GOFMT) -s -w .

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## migrate-up: Run database migrations up
migrate-up:
	$(GORUN) $(MAIN_MIGRATE) up

## migrate-down: Run database migrations down one step
migrate-down:
	$(GORUN) $(MAIN_MIGRATE) down

## migrate-status: Show migration status
migrate-status:
	$(GORUN) $(MAIN_MIGRATE) status

## swagger: Generate Swagger documentation
swagger:
	swag init -g $(MAIN_API) -o docs --parseInternal --parseDependency

## docker-up: Start all services with Docker Compose
docker-up:
	docker compose -f deployments/docker/docker-compose.yml --env-file .env up -d --build --remove-orphans

## docker-down: Stop all Docker Compose services
docker-down:
	docker compose -f deployments/docker/docker-compose.yml down

## docker-dev: Start development environment
docker-dev:
	docker compose -f deployments/docker/docker-compose.yml -f deployments/docker/docker-compose.dev.yml up -d --build --remove-orphans

## docker-logs: View Docker Compose logs
docker-logs:
	docker compose -f deployments/docker/docker-compose.yml logs -f
