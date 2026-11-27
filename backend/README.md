# E-Wardrobe Backend

RESTful API server for the E-Wardrobe wardrobe management application, built with Go.

## Tech Stack

- **Language:** Go 1.27
- **Web Framework:** [chi](https://github.com/go-chi/chi) — lightweight HTTP router
- **Database:** PostgreSQL 16
- **Database Driver:** [pgx/v5](https://github.com/jackc/pgx)
- **Migrations:** [goose](https://github.com/pressly/goose)
- **Query Generation:** [sqlc](https://github.com/sqlc-dev/sqlc) — type-safe SQL from plain SQL
- **Image Processing:** `golang.org/x/image` — JPEG decoding, thumbnail generation
- **Storage:** Local filesystem (extensible via `PhotoStorage` interface)
- **Testing:** [testcontainers-go](https://github.com/testcontainers) — real PostgreSQL in Docker

## Architecture

```
cmd/server/          — application entry point, wiring
internal/
  auth/              — user registration, login, sessions, CSRF, recovery
  categories/        — hierarchical category management
  changejournal/     — outbox pattern for sync changes
  config/            — environment-based configuration
  httpapi/           — HTTP handlers, middleware, router
  items/             — wardrobe item CRUD
  photos/            — photo upload pipeline (decode → thumbnail → store)
  platform/
    database/        — pgx pool, sqlc generated queries, transactions
    storage/         — photo storage abstraction (local filesystem)
    tokens/          — secure token generation
  sync/              — bidirectional sync (push/pull with conflict resolution)
  webui/             — embedded SPA static files
```

### Key Design Patterns

- **Clean Architecture** — handlers → services → repositories → database
- **Outbox Pattern** — `changejournal` writes sync changes atomically with business transactions
- **Idempotent Sync** — `processed_mutations` table deduplicates client pushes
- **Canonical Ordering** — sync uses monotonically increasing sequence numbers
- **Interface-driven Storage** — `PhotoStorage` abstraction allows swapping local disk for S3
- **Optimistic Concurrency** — `version` column on all entities prevents lost updates

## Database Schema

Three migration files define the schema:

| Migration | Contents |
|-----------|----------|
| `00001_identity.sql` | `users`, `sessions`, `invitations`, `account_recovery_tokens` |
| `00002_wardrobe.sql` | `categories`, `items`, `item_photos` |
| `00003_sync.sql` | `user_sync_state`, `sync_changes`, `processed_mutations` |

All tables use `UUID` primary keys, `TIMESTAMPTZ` timestamps, and soft deletes via `deleted_at`.

## API Endpoints

### Public (no auth)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health/live` | Liveness probe |
| `GET` | `/health/ready` | Readiness probe (checks DB) |
| `POST` | `/api/v1/auth/register` | Register new user |
| `POST` | `/api/v1/auth/login` | Login, get session cookie |
| `POST` | `/api/v1/auth/logout` | Revoke session |
| `GET` | `/api/v1/auth/csrf` | Get CSRF token (web clients) |
| `POST` | `/api/v1/auth/invitations/validate` | Validate invitation token |
| `POST` | `/api/v1/auth/recovery/validate` | Start account recovery |
| `POST` | `/api/v1/auth/recovery/complete` | Complete account recovery |

### Protected (session cookie required)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/categories` | List categories |
| `POST` | `/api/v1/categories` | Create category |
| `PATCH` | `/api/v1/categories/:id` | Update category |
| `DELETE` | `/api/v1/categories/:id` | Archive category |
| `GET` | `/api/v1/items` | List items (paginated, filterable) |
| `POST` | `/api/v1/items` | Create item |
| `GET` | `/api/v1/items/:id` | Get item details |
| `PATCH` | `/api/v1/items/:id` | Update item |
| `DELETE` | `/api/v1/items/:id` | Delete item |
| `PATCH` | `/api/v1/items/:id/archive` | Archive/unarchive item |
| `GET` | `/api/v1/items/storage-locations` | Storage location suggestions |
| `POST` | `/api/v1/photos` | Create photo placeholder |
| `PUT` | `/api/v1/photos/:id/content` | Upload photo content |
| `GET` | `/api/v1/photos/:id/content` | Get photo content |
| `DELETE` | `/api/v1/photos/:id` | Delete photo |
| `PATCH` | `/api/v1/photos/order` | Reorder photos |
| `POST` | `/api/v1/sync/push` | Push local mutations to server |
| `GET` | `/api/v1/sync/pull?since=&limit=` | Pull server changes |

### SPA Fallback

All unmatched routes serve the embedded SPA from `internal/webui/dist/`.

## Configuration

All configuration is loaded from environment variables. See `.env.example`:

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | _(required)_ | PostgreSQL DSN |
| `HTTP_ADDR` | `:8080` | Listen address |
| `PUBLIC_BASE_URL` | `http://localhost:8080` | Public base URL |
| `PHOTO_STORAGE_DIR` | `./photos` | Local photo storage directory |
| `SESSION_COOKIE_SECURE` | `false` | HTTPS-only session cookies |
| `SESSION_TTL` | `24h` | Session expiration |
| `INVITATION_DEFAULT_TTL` | `168h` | Invitation token lifetime |
| `MAX_UPLOAD_BYTES` | `12582912` | Max upload size (12 MiB) |
| `LOG_LEVEL` | `info` | Log level: debug, info, warn, error |

## Development

### Prerequisites

- Go 1.27+
- PostgreSQL 16+
- Docker (for tests and `docker-up` make targets)
- sqlc CLI (for query generation)
- golangci-lint (optional, for `make lint`)

### Setup

```bash
# Copy environment
cp .env.example .env

# Run migrations
make migrate-up

# Start PostgreSQL via Docker (if using dev compose)
make docker-up
```

### Build & Run

```bash
# Build both server and CLI
make build

# Run the HTTP server
make run
# or: go run ./cmd/server
```

### Generate Queries

```bash
make sqlc-gen
```

This reads `sqlc.yaml`, processes SQL files from `sql/queries/` against the schema in `migrations/`, and generates typed Go code in `internal/platform/database/sqlc/`.

### Tests

```bash
# Run all tests with race detector
make test

# Run specific race test
make test-race
```

Tests use real PostgreSQL via testcontainers — no mocks for database access.

### Lint

```bash
make lint
```

Requires `golangci-lint` installed.

### Docker

```bash
# Build and run via Docker
docker build -t e-wardrobe-backend .
docker run -p 8080:8080 --env-file .env e-wardrobe-backend
```

The Dockerfile uses a multi-stage build: compiles in a `golang:1.27-alpine` stage, then runs from a minimal `alpine:3.20` image as a non-root user.

## Project Structure

```
backend/
├── cmd/
│   ├── server/main.go          — HTTP server entry point
│   └── wardrobe/main.go        — CLI entry point
├── internal/
│   ├── auth/                   — Authentication & session management
│   ├── categories/             — Category domain logic
│   ├── changejournal/          — Outbox for sync changes
│   ├── config/                 — Configuration loading
│   ├── httpapi/                — HTTP layer (handlers, middleware)
│   ├── items/                  — Items domain logic
│   ├── photos/                 — Photo upload & processing
│   ├── platform/
│   │   ├── database/           — DB pool, sqlc queries, transactions
│   │   ├── storage/            — Photo storage abstraction
│   │   └── tokens/             — Secure token utilities
│   ├── sync/                   — Bidirectional sync engine
│   └── webui/                  — Embedded SPA
├── sql/
│   └── queries/                — Raw SQL queries for sqlc
├── migrations/                 — goose migration files (parent dir)
├── Dockerfile
├── Makefile
├── sqlc.yaml
└── .env.example
```

## Sync Protocol

The sync system enables offline-first clients:

1. **Push** — Client sends local mutations with unique `mutation_id`. Server applies them, records results in `processed_mutations` for deduplication.
2. **Pull** — Client sends `since` (last known sequence). Server returns all changes from `sync_changes` since that sequence.
3. **Conflict Resolution** — Each mutation includes `base_version`. If the entity version has changed, the server returns a `conflict` status with the current entity state.
4. **Canonical Ordering** — Changes are assigned monotonically increasing `seq` numbers per user, providing a total order for pull operations.
