# Zencial VOD API

Go-based Video-on-Demand API following Clean Architecture, Clean Code, and SOLID principles.

**Stack:** Go 1.25, chi/v5, PostgreSQL 16 (pgx), Redis 7, MinIO (S3), JWT, goose migrations, slog logging

## Architecture

This project follows **Hexagonal/Clean Architecture**. Dependencies always point inward — outer layers depend on inner layers, never the reverse.

```
Domain (entities, value objects, repository interfaces, events)
  ↑
Use Cases (application services orchestrating domain logic)
  ↑
Adapters (HTTP handlers, DTOs, mappers, event dispatcher)
  ↑
Infrastructure (postgres repos, redis, minio, middleware, config)
```

### Layer Rules

- **Domain** (`internal/domain/`) — Zero external dependencies. Contains entities, value objects, repository interfaces, domain events, and domain errors. Never import from usecase, adapter, or infrastructure.
- **Use Cases** (`internal/usecase/`) — Orchestrate domain logic. Depend only on domain interfaces. Never import infrastructure or adapter packages.
- **Adapters** (`internal/adapter/`) — Convert between external world and use cases. HTTP handlers parse requests, call use case services, and return DTOs. Never access repositories directly.
- **Infrastructure** (`internal/infrastructure/`) — Implements domain interfaces (repositories, storage, auth). Provides middleware, config, and server setup.

### Dependency Injection

Manual constructor injection in `cmd/api/main.go`. No DI framework. All dependencies are explicit and passed via constructors. Use case services receive repository interfaces, not concrete implementations.

## Directory Structure

```
cmd/api/main.go              — API entry point, wires all dependencies
cmd/migrate/main.go           — Migration CLI
internal/domain/entity/       — Domain entities (User, Video, Genre)
internal/domain/valueobject/  — Value objects (Email, Slug, HashedPassword, Duration)
internal/domain/repository/   — Repository interfaces (contracts)
internal/domain/event/        — Domain events and dispatcher interface
internal/domain/service/      — Domain services
internal/usecase/{auth,video,genre}/  — Use case services with Input/Output types
internal/adapter/handler/v1/  — HTTP handlers, router, DTOs, mappers
internal/adapter/messaging/   — Event dispatcher implementation
internal/infrastructure/persistence/postgres/  — PostgreSQL repository implementations
internal/infrastructure/persistence/redis/     — Redis session store
internal/infrastructure/auth/      — JWT and bcrypt implementations
internal/infrastructure/storage/   — MinIO storage implementation
internal/infrastructure/middleware/ — HTTP middleware (auth, RBAC, CORS, logging, rate limit)
internal/infrastructure/config/    — Environment-based configuration
internal/infrastructure/server/    — HTTP server with graceful shutdown
internal/pkg/apperror/        — AppError type with HTTP status mapping
internal/pkg/filter/           — Query filtering and sorting
internal/pkg/httputil/         — HTTP response helpers
internal/pkg/validator/        — Request validation
internal/pkg/thumbnail/        — FFmpeg thumbnail extraction
internal/pkg/clock/            — Time abstraction for testing
internal/pkg/pagination/       — Pagination helpers
migrations/                    — SQL migrations (goose format)
deployments/docker/            — Dockerfile, docker-compose
```

## SOLID Principles

### Single Responsibility

- Each use case method handles one business operation (e.g., `Upload()`, `Publish()`, `Delete()`)
- Handlers only parse HTTP requests and delegate to use case services
- Mappers are separate from handlers — DTOs map to/from domain entities in `mapper/`
- Repository implementations only handle data access

### Open/Closed

- New features extend the system by adding new use case methods, handlers, and events
- The event dispatcher allows adding behavior (logging, analytics) without modifying existing use cases
- Filter system is configurable per-repository without modifying the filter engine

### Liskov Substitution

- Repository interfaces are implemented by concrete types that fulfill the full contract
- Value objects use separate constructors for trusted (`*FromTrusted()`) vs untrusted (`New*()`) sources — both return the same type
- `StorageService` interface is implemented by MinIO but could be swapped for any S3-compatible service

### Interface Segregation

- Repository interfaces are scoped per aggregate (`VideoRepository`, `UserRepository`, `GenreRepository`)
- `StorageService` exposes only the operations needed by use cases
- `TokenService` is separate from `Hasher` — authentication concerns are split

### Dependency Inversion

- Use cases depend on repository **interfaces** defined in the domain layer, not on PostgreSQL implementations
- `StorageService` interface lives in infrastructure/storage but is consumed by use cases via interface
- `EventDispatcher` interface is in domain/event, implementation is in adapter/messaging

## Coding Conventions

### Naming

- Packages: lowercase, descriptive (`auth`, `storage`, `middleware`, `persistence`)
- Interfaces: CamelCase, descriptive (`TokenService`, `StorageService`, `UserRepository`)
- Error variables: `Err` prefix (`ErrUserNotFound`, `ErrInvalidToken`)
- Constructors: `New*()` returns `*Type` or `(*Type, error)`

### Error Handling

- Use case methods return `(*Output, *apperror.AppError)` — never raw errors
- Use `apperror` factory functions: `NotFound()`, `BadRequest()`, `Unauthorized()`, `Forbidden()`, `Conflict()`, `Internal()`
- Error codes are centralized in `internal/pkg/apperror/codes.go`
- Domain errors (`internal/domain/errors.go`) are wrapped by `apperror` with appropriate HTTP status

### Value Objects

- Immutable types that enforce constraints at creation
- `New*()` validates input from external sources (API requests, user input)
- `*FromTrusted()` skips validation for data from trusted sources (database reads)
- Examples: `Email`, `Slug`, `HashedPassword`, `Duration`, `Pagination`

### Context

- `context.Context` is always the first parameter
- Transaction context is propagated via context for repository composability
- Middleware stores userID and role in context, extracted via helper functions

### Cleanup

- Use `defer` for resource cleanup (temp files, database connections)
- On failure in multi-step operations, clean up already-completed steps (e.g., delete uploaded files if DB insert fails)

### Soft Delete

- Videos are archived (status change + files moved to `deleted/` prefix in storage), not hard deleted
- Unarchive restores files and sets status to draft

## Testing

### Patterns

- **Table-driven tests** with `t.Run("scenario_name", ...)`
- **testify** for assertions: `assert.Equal()`, `require.NotNil()`, `assert.ErrorIs()`
- **Mock repositories** with closure fields (e.g., `getUserFn func(...) ...`) for flexible test doubles
- **Test helpers**: `newTestService()`, `newActiveUser()` for common setup
- **Event verification**: check `dispatcher.dispatched` slice for dispatched events

### What to Test

- Value object validation and edge cases
- Use case business logic with mocked repositories
- Domain entity state transitions
- Error scenarios and error code correctness

### Running Tests

```bash
make test          # Run all tests with race detector
make test-cover    # Run with coverage report
make lint          # Run golangci-lint
```

## Common Commands

```bash
make build              # Build API binary
make run                # Run API with go run
make docker-dev         # Start full dev environment (API + Postgres + Redis + MinIO)
make docker-up          # Start production stack
make docker-down        # Stop all containers
make migrate-up         # Apply pending migrations
make migrate-down       # Rollback one migration
make migrate-status     # Check migration status
make swagger            # Generate Swagger docs
make fmt                # Format code
make clean              # Remove build artifacts
```
