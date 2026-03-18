# go-arch — Claude Code Instructions

## Project Overview

Go Clean Architecture skeleton for microservices with multi-database support, graceful shutdown, and clear layer separation.

## Architecture

```
cmd/app/main.go                      ← Entry point (--mode=http|cron|nsq, -t)
internal/app/app.go                  ← Dependency wiring
internal/handler/http|cron|nsq/      ← Transport layer (1 handler : 1 usecase)
internal/usecase/                    ← Orchestrates services (1 usecase → N services)
internal/service/                    ← Single-responsibility business logic
internal/repository/                 ← Data access interfaces + implementations
internal/entity/                     ← Pure domain models
internal/infra/                      ← Config, DB, Redis, Logger, HTTP client
library/                             ← Stateless shared utilities
```

## Layer Rules

- `handler` → calls exactly 1 usecase
- `usecase` → composes N services, never touches repository directly
- `service` → reusable across usecases, calls repository interfaces
- `repository` → receives `*sql.DB` via constructor (DI), no global state
- `entity` → plain Go structs, zero dependencies
- `library` → stateless helpers, no business logic, importable by any layer

## Key Commands

```bash
make run          # go run ./cmd/app/ --mode=http
make run-cron     # go run ./cmd/app/ --mode=cron
make run-nsq      # go run ./cmd/app/ --mode=nsq
make config-test  # go run ./cmd/app/ --mode=http -t
make build        # go build -o bin/app ./cmd/app/
make vet          # go vet ./...
make test         # go test ./...
```

> **Note:** Build with `CGO_ENABLED=0` on macOS due to Go 1.22.6 + Darwin 25.3 dyld compatibility issue.

## Config

- Config file: `files/config/app.yaml` (gitignored)
- Template: `files/config/app.yaml.sample`
- Supports multi-database (`resource.database.<name>`) and multi-redis (`resource.redis.<name>`)

## Adding a New Feature

1. **Entity** — add struct in `internal/entity/`
2. **Repository** — define interface + implementation in `internal/repository/`
3. **Service** — implement business logic in `internal/service/`, inject repository interface
4. **Usecase** — compose services in `internal/usecase/`, inject service interfaces
5. **Handler** — add HTTP handler in `internal/handler/http/`, register route in `router.go`
6. **Wire** — connect dependencies in `internal/app/app.go`

## Dependencies

| Package | Purpose |
|---|---|
| `gorilla/mux` | HTTP routing |
| `rs/zerolog` | Structured logging |
| `redis/go-redis/v9` | Redis client |
| `go-sql-driver/mysql` | MySQL driver |
| `lib/pq` | PostgreSQL driver |
| `gopkg.in/yaml.v3` | Config parsing |
| `google/uuid` | Request ID generation |
