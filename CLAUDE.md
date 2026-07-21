# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Bookstore Backend (博学书城后端) — a RESTful API for an online bookstore built with Go, Gin, GORM, and Redis.

## Commands

```bash
# Build
make bookstore-manager          # compile to bin/bookstore-manager
go build -o bin/bookstore-manager ./cmd/bookstore-manager

# Run
make run-bookstore-manager      # run compiled binary
go run ./cmd/bookstore-manager  # run directly from source

# Dependencies
go mod download

# Test (no tests exist yet)
go test ./...

# Clean
make clean
```

## Architecture

4-layer pattern: **Controller → Service → Repository → Model**

All API routes are under `/api/v1`. JSON responses use `{"code": 0/-1, "message": "...", "data": ...}`.

| Layer | Directory | Responsibility |
|---|---|---|
| Entry point | `cmd/bookstore-manager/` | Init config → MySQL → Redis → router, then graceful shutdown |
| Config | `config/` + `conf/config.yaml` | YAML config loaded into `config.AppConfig` |
| Global clients | `global/` | `global.DBClient` (GORM) and `global.RedisClient` (go-redis) |
| Controllers | `web/controller/` | HTTP handlers, request binding, response formatting |
| Services | `service/` | Business logic, validation, orchestration |
| Repositories | `repository/` | Thin GORM query wrappers (DAOs) |
| Models | `model/` | GORM structs for `users`, `books`, `favorites`, `orders`, `order_items`, `carousel`, `categories` |
| Router | `web/router/` | Route registration + CORS middleware |
| Auth middleware | `web/middleware/` | JWT Bearer token validation, sets `userID`/`username` in context |
| JWT | `jwt/` | Token generation (access 2h / refresh 7d), parsing, revocation via Redis |
| Cache | `cache/` | Redis book caching with singleflight, TTL jitter, null-value caching |
| SQL | `sql/` | DDL (`bookstore.sql`) and seed data (`mock.sql`) |

## Key Patterns

- **Auth**: JWT access tokens in `Authorization: Bearer <token>` header. Tokens stored in Redis hashes for revocation. Middleware at `web/middleware/middleware.go`.
- **Caching**: Book data cached in Redis with singleflight to prevent cache breakdown. TTLs have random jitter (0-60s). Null values cached for 1min to prevent penetration. Search is never cached.
- **Dependencies wired in main**: controllers → services → repositories are manually instantiated in `cmd/bookstore-manager/bookstore-manager.go`.
- **GORM models** use `gorm:"column:..."` tags; table names are pluralized from struct names (e.g., `OrderItem` → `order_items`).

## Database

MySQL + GORM. Schema in `sql/bookstore.sql`, seed data in `sql/mock.sql`. Config in `conf/config.yaml`.

## Adding a New Endpoint

1. Add model in `model/` if needed
2. Add DAO methods in `repository/`
3. Add business logic in `service/`
4. Add handler in `web/controller/`
5. Register route in `web/router/router.go`
6. Wire dependencies in `cmd/bookstore-manager/bookstore-manager.go`