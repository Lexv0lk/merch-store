# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Run full stack (gateway + auth + store + postgres + migrator) with docker-compose
make build-app-up    # build and start
make app-up          # start existing images

# Run database only (postgres + migrator)
make infra-up

# Run all tests with race detection and coverage
make test

# Run a single test
go test -run TestName ./path/to/package/...

# Database migrations (requires goose CLI, default DSN uses localhost:5433)
make migrate-up
make migrate-down
make migrate-status
```

## Architecture

This is a Go microservices project with three services communicating via gRPC, fronted by an HTTP REST gateway.

**Services:**
- **Gateway** (`cmd/gateway`, `internal/gateway`) — HTTP REST API on port 8080 using Gin. Translates HTTP requests to gRPC calls. Handles Bearer token extraction and forwards JWT via gRPC metadata.
- **Auth** (`cmd/auth`, `internal/auth`) — gRPC service on port 9090. Handles registration and login (auto-creates users on first auth). Uses Argon2id for password hashing. Issues HS256 JWT tokens (1h expiry).
- **Store** (`cmd/store`, `internal/store`) — gRPC service on port 9091. Manages coin balances, merchandise purchases, and coin transfers. JWT validated via gRPC unary interceptor.

**Request flow:** HTTP → Gateway → gRPC → Auth/Store service → PostgreSQL

**Each service follows clean architecture internally:**
```
service/
├── bootstrap/        # Service initialization and dependency wiring
├── domain/           # Interfaces and domain error types
├── application/      # Use cases (business logic)
├── grpc/             # gRPC server handlers
└── infrastructure/   # PostgreSQL repository implementations
```

**Shared internal packages** (`internal/pkg/`): database utilities, env helpers, JWT issuing/parsing, logging interface.

**Generated code** (`gen/`): Protobuf/gRPC stubs from `api/merch/v1/*.proto` and golang/mock mocks.

## Database

PostgreSQL with Goose migrations in `migrations/`. Key tables: `users`, `balances`, `goods`, `transactions`, `purchases`. The `goods` table is seeded with 10 items via migration.

Coin transfers and purchases use `BeginTx` with `ReadCommitted` isolation and `FOR UPDATE` row locking for concurrency safety.

## Testing

- Unit tests use interface-based mocking (golang/mock, mocks in `gen/mocks/`)
- **When new mocks are needed** (e.g. for new tests), generate them via `mockgen` and store in `gen/mocks/` following the existing folder structure
- Integration tests in `tests/integration/` use testcontainers to spin up real PostgreSQL instances
- Integration tests run full scenarios: auth → transfer coins, auth → purchase items

## Configuration

Environment variables (see `.env.example`): `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `GRPC_AUTH_PORT`, `GRPC_STORE_PORT`, `GRPC_AUTH_HOST`, `GRPC_STORE_HOST`, `HTTP_PORT`, `JWT_SECRET`.

Docker builds use multi-stage alpine images. Docker Compose profiles: `app` (full stack), `infra` (database only).
