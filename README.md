# Zencial VoD API

Zencial is a Danish-based Video-on-Demand (VoD) platform that empowers users to submit, stream, and monetize their own videos, movies, and series. The platform supports both free and paid content, enabling creators to reach audiences and earn revenue, while viewers can enjoy a wide range of user-generated content.

---

## Table of Contents

- [Zencial VoD API](#zencial-vod-api)
  - [Table of Contents](#table-of-contents)
  - [Project Overview](#project-overview)
  - [Architecture](#architecture)
  - [Directory Structure](#directory-structure)
  - [Features](#features)
    - [Core (Implemented)](#core-implemented)
    - [In Progress / Planned](#in-progress--planned)
  - [Roles \& Permissions](#roles--permissions)
    - [Basic User](#basic-user)
    - [Premium Plan User (e.g., $20/month)](#premium-plan-user-eg-20month)
    - [Publisher](#publisher)
    - [Admin](#admin)
  - [Setup \& Usage](#setup--usage)
    - [Prerequisites](#prerequisites)
    - [Quick Start (Development)](#quick-start-development)
    - [Environment Variables](#environment-variables)
  - [API \& Documentation](#api--documentation)
  - [Testing](#testing)
  - [Contributing](#contributing)
  - [License](#license)

---

## Project Overview

Zencial VoD is built with Go, following Clean Architecture and SOLID principles for maintainability and scalability. It leverages PostgreSQL, Redis, Minio (S3-compatible), JWT authentication, and a modular, testable codebase. The platform is designed for:

- **Creators**: Upload, manage, and monetize video content.
- **Viewers**: Watch free or paid content, subscribe to plans, and interact with creators.
- **Admins**: Oversee users, content, plans, analytics, and platform health.

---

## Architecture

Zencial follows a strict **Hexagonal/Clean Architecture**:

```
Domain (entities, value objects, repository interfaces, events)
	↑
Use Cases (application services orchestrating domain logic)
	↑
Adapters (HTTP handlers, DTOs, mappers, event dispatcher)
	↑
Infrastructure (Postgres repos, Redis, S3 storage, middleware, config)
```

- **Domain**: Pure business logic, no dependencies.
- **Use Cases**: Application services, orchestrate domain logic.
- **Adapters**: HTTP handlers, DTOs, mappers, event dispatcher.
- **Infrastructure**: Database, storage, middleware, config, server setup.

**Dependency injection** is manual and explicit (see `cmd/api/main.go`).

---

## Directory Structure

```
cmd/api/main.go              — API entry point, wires dependencies
cmd/migrate/main.go          — Migration CLI
internal/domain/             — Entities, value objects, repository interfaces, events
internal/usecase/            — Use case services
internal/adapter/            — HTTP handlers, DTOs, mappers, event dispatcher
internal/infrastructure/     — DB, Redis, S3, middleware, config, server
internal/pkg/                — App error, filter, HTTP util, validator, thumbnail, clock, pagination
migrations/                  — SQL migrations (goose format)
deployments/docker/          — Dockerfile, docker-compose
docs/                        — Swagger/OpenAPI docs
```

---

## Features

### Core (Implemented)

- Upload video content
- Stream video content
- Manage plans and genres

### In Progress / Planned

- Create and manage series
- User management (registration, login, profile)
- Subscription management
- Payment integration
- Watchlist and favorites
- Analytics and watch histories
- Admin dashboard and platform analytics

---

## Roles & Permissions

All roles and permissions are managed by admins. Key roles:

### Basic User

- Watch all free content
- Pay one-time fee for paid content

### Premium Plan User (e.g., $20/month)

- Watch all free and premium content
- Access 4K content (if available)
- Submit free and paid content (up to a price cap, e.g., $10; Zencial takes 5% cut)

### Publisher

- Must have an active Premium Plan
- Can submit and monetize content

### Admin

- Set only in DB
- Full access: manage all content, users, plans, analytics

---

## Setup & Usage

### Prerequisites

- Go 1.26+
- PostgreSQL 16+
- Redis 7+
- Minio or S3-compatible storage

### Quick Start (Development)

```bash
# Clone repo and enter directory
git clone https://github.com/zenfulcode/zencial.git
cd zencial-vid

# Build API binary
make build

# Run API server
make run

# Start full dev environment (API + Postgres + Redis + Minio)
make docker-dev

# Apply DB migrations
make migrate-up

# Run tests
make test
```

### Environment Variables

Copy `.env.example` to `.env` and configure DB, Redis, S3, JWT, etc.

---

## API & Documentation

- RESTful API (see `docs/swagger.yaml` for OpenAPI spec)
- Versioned under `/api/v1`
- JWT-based authentication
- Error responses use a consistent envelope (see `internal/pkg/apperror`)

---

## Testing

- Table-driven tests with `testify`
- Mock repositories for business logic
- Run all tests: `make test`
- Coverage report: `make test-cover`
- Lint: `make lint`

---

## Contributing

1. Fork and clone the repo
2. Create a feature branch
3. Write tests before code (TDD)
4. Run `make lint test`
5. Submit a pull request

---

## License

See [LICENSE](LICENSE) for details.
