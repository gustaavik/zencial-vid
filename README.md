# Zencial

Zencial is a community-driven video streaming platform where users upload and share homemade films and series. Submissions are reviewed by moderators before they go live. Once a creator's first video is approved, they gain access to a **Publisher Dashboard** where they can manage their content, casts, and analytics.

The platform supports both **ad-supported** and **premium** content. By default, videos include ads; publishers with an active subscription can release paid content. Viewers with a subscription plan get benefits such as ad-free viewing, 4K streaming, access to premium content, and discounts.

This repository (`zencial-vid`) hosts the **Zencial VoD API** — the Go backend that powers the platform. Sibling repositories provide the rest of the system:

- `frontend/` — Expo / React Native client (web + mobile)
- `zencial-cdn/` — HLS transcoder and signed-URL CDN service

---

## Table of Contents

- [Zencial](#zencial)
  - [Platform Overview](#platform-overview)
  - [Features](#features)
    - [User Features](#user-features)
    - [Publisher Dashboard (planned)](#publisher-dashboard-planned)
    - [Admin Dashboard](#admin-dashboard)
  - [Roles \& Permissions](#roles--permissions)
  - [About this repo (Zencial VoD API)](#about-this-repo-zencial-vod-api)
    - [Architecture](#architecture)
    - [Directory Structure](#directory-structure)
    - [Setup \& Usage](#setup--usage)
    - [API \& Documentation](#api--documentation)
    - [Testing](#testing)
    - [Contributing](#contributing)
    - [License](#license)

---

## Platform Overview

- **Community submissions.** Any registered user can upload homemade films and series.
- **Moderator approval.** Every new submission enters a review queue and only becomes public after a moderator approves it.
- **Publisher unlock.** A user becomes a Publisher when their first video is approved. Publishers get the Publisher Dashboard for managing their catalog, casts, and analytics.
- **Ad-supported by default.** Videos play with ads unless the publisher has a subscription that unlocks paid content.
- **Premium content.** Publishers on a subscription can release videos behind a minimum plan tier; viewers on that tier (or higher) can watch them.
- **Subscription benefits.** Viewer plans unlock ad-free playback, 4K streaming, premium content, and discounts.

---

## Features

Items marked **(planned)** are part of the roadmap and not yet implemented in the API.

### User Features

- Watchlists — implemented
- Watch history — implemented
- Continue watching — implemented (resume from last position via watch-progress)
- Active sessions management — implemented (list and revoke device sessions)

### Publisher Dashboard (planned)

Available to users with the **Publisher** role (unlocked after a user's first approved upload):

- Manage videos, films, and series (planned)
- Manage casts and contributors (planned)
- Analytics — views, reviews, paid-content performance (planned)
- Ad revenue tracking (planned)

> Today, publishers manage uploads through the existing video endpoints (`POST /api/v1/videos`, `PATCH /api/v1/videos/{id}`, etc.). The dedicated dashboard surface and the cast / analytics models are not yet built.

### Admin Dashboard

Available to **Admins** and **Moderators**:

- Manage users — sessions, roles, subscriptions — implemented
- Manage all platform content — implemented
- Manage categories (genres) — implemented
- Manage subscriptions and plans — implemented (Stripe-backed billing)
- Audit logs — implemented
- Review and approve submissions (planned) — the `VideoStatus` lifecycle exists, but the moderator review queue and approval endpoints are not yet exposed
- Ads management overview (planned)

---

## Roles & Permissions

| Role          | Granted by                               | Capabilities                                                                                |
| ------------- | ---------------------------------------- | ------------------------------------------------------------------------------------------- |
| **Viewer**    | Registration                             | Watch ad-supported content, build watchlists, track history.                                |
| **Subscriber**| Active subscription plan                 | All Viewer rights plus ad-free playback, 4K streaming, premium content, plan-tier perks.    |
| **Publisher** | First video approved by a moderator      | All Subscriber rights plus upload, manage own catalog, and (planned) the Publisher Dashboard. Releasing paid content requires an active subscription. |
| **Moderator** | Granted by Admin                         | Review and approve/reject submissions (planned), moderate content and users.                |
| **Admin**     | Set directly in the database             | Full access to users, content, plans, billing, categories, and audit logs.                  |

---

## About this repo (Zencial VoD API)

The VoD API is a Go service built on Clean / Hexagonal Architecture with PostgreSQL, Redis, and S3-compatible storage. It exposes a versioned REST API under `/api/v1` with JWT authentication.

### Architecture

```
Domain (entities, value objects, repository interfaces, events)
  ↑
Use Cases (application services orchestrating domain logic)
  ↑
Adapters (HTTP handlers, DTOs, mappers, event dispatcher)
  ↑
Infrastructure (Postgres repos, Redis, S3 storage, middleware, config)
```

- **Domain** — pure business logic, no external dependencies.
- **Use Cases** — application services that orchestrate domain logic.
- **Adapters** — HTTP handlers, DTOs, mappers, event dispatcher.
- **Infrastructure** — database, storage, middleware, config, server setup.

Dependency injection is manual and explicit (see [cmd/api/main.go](cmd/api/main.go)).

### Directory Structure

```
cmd/api/main.go              — API entry point, wires dependencies
cmd/migrate/main.go          — Migration CLI
internal/domain/             — Entities, value objects, repository interfaces, events
internal/usecase/            — Use case services (auth, video, watchlist, watchprogress, session, audit, ...)
internal/adapter/            — HTTP handlers, DTOs, mappers, event dispatcher
internal/infrastructure/     — Postgres, Redis, S3, middleware, config, server
internal/pkg/                — apperror, filter, httputil, validator, thumbnail, clock, pagination
migrations/                  — SQL migrations (goose format)
deployments/docker/          — Dockerfile, docker-compose
docs/                        — Swagger / OpenAPI docs
```

### Setup & Usage

**Prerequisites**

- Go 1.26+
- PostgreSQL 16+
- Redis 7+
- Minio or S3-compatible storage

**Quick start (development)**

```bash
# Clone and enter the repo
git clone https://github.com/zenfulcode/zencial.git
cd zencial-vid

# Start full dev environment (API + Postgres + Redis + Minio)
make docker-dev

# Apply DB migrations
make migrate-up

# Or run the API directly against running services
make run

# Run tests
make test
```

**Environment variables**

Copy `.env.example` to `.env` and configure DB, Redis, S3, JWT, and Stripe credentials.

### API & Documentation

- RESTful API versioned under `/api/v1` (see [docs/swagger.yaml](docs/swagger.yaml)).
- JWT-based authentication with session tracking in Redis.
- Error responses use a consistent envelope; see [internal/pkg/apperror](internal/pkg/apperror).

### Testing

- Table-driven tests with `testify`.
- Mock repositories use closure fields for flexible test doubles.
- Run all tests: `make test` (race detector enabled).
- Coverage report: `make test-cover`.
- Lint: `make lint` (golangci-lint).

### Contributing

1. Fork and clone the repo.
2. Create a feature branch.
3. Write tests first (TDD).
4. Run `make lint test` before opening a PR.
5. Submit a pull request following conventional-commit messages (`feat:`, `fix:`, `refactor:`, ...).

### License

No license file is published yet — Zencial is proprietary until a `LICENSE` is added at the repo root.
