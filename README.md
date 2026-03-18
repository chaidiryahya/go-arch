# go-arch

A production-ready Go microservice skeleton implementing **Clean Architecture** / **Repository Pattern** with multi-database support, graceful shutdown, and clear layer separation.

> This repository was fully assisted and coded by [Claude](https://claude.ai) (Anthropic).

---

## Architecture

```
go-arch/
├── cmd/app/main.go                      ← Entry point (--mode=http|cron|nsq, -t)
├── internal/
│   ├── app/app.go                       ← Dependency wiring & graceful shutdown
│   ├── handler/
│   │   ├── http/                        ← HTTP transport (router, middleware, response)
│   │   ├── cron/                        ← Scheduled cron job handler
│   │   └── nsq/                         ← NSQ message consumer handler
│   ├── usecase/                         ← Orchestrates services (1 usecase → N services)
│   ├── service/                         ← Single-responsibility business logic
│   ├── repository/                      ← Data access interfaces + implementations
│   ├── entity/                          ← Pure domain models (no dependencies)
│   └── infra/
│       ├── configuration/               ← YAML config loader
│       ├── db/                          ← Multi-database connection registry
│       ├── redis/                       ← Redis connection registry
│       ├── log/                         ← Zerolog structured logger
│       └── net/                         ← Reusable HTTP client
├── library/                             ← Stateless shared utilities
└── files/config/
    └── app.yaml.sample                  ← Config template
```

### Layer Rules

| Layer | Responsibility | Rule |
|---|---|---|
| `handler` | Receive request, validate input, return response | Calls exactly 1 usecase |
| `usecase` | Orchestrate business flow | Composes N services, never touches repository directly |
| `service` | Single-responsibility business logic | Reusable across usecases, calls repository interfaces |
| `repository` | Data access abstraction | Interface-driven, receives `*sql.DB` via constructor injection |
| `entity` | Domain models | Plain Go structs, zero dependencies |
| `infra` | Infrastructure concerns | DB, Redis, config, logging |
| `library` | Shared utilities | Stateless helpers, importable by any layer |

### Dependency Flow

```
handler → usecase → service → repository → database
```

---

## Service Modes

| Flag | Mode | Description |
|---|---|---|
| `--mode=http` | HTTP Server | REST API on configurable port (default `:8080`) |
| `--mode=cron` | Cron Jobs | Scheduled background jobs |
| `--mode=nsq` | NSQ Consumer | Message queue consumer |
| `-t` | Config Test | Validate config file and exit |

---

## How to Run

### 1. Setup config

```bash
cp files/config/app.yaml.sample files/config/app.yaml
# Edit files/config/app.yaml with your database/redis credentials
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run

```bash
# HTTP server (default port :8080)
make run

# Cron jobs
make run-cron

# NSQ consumer
make run-nsq

# Validate config only
make config-test
```

### 4. Build

```bash
make build
# Binary output: bin/app
```

### 5. Health check

```bash
curl http://localhost:8080/health
# {"status":"ok","server_process_time":"...","data":{"status":"ok"}}
```

---

## Configuration

`files/config/app.yaml` supports multiple named database and redis connections:

```yaml
server:
  port: 8080

resource:
  database:
    main:
      driver: mysql
      host: localhost
      port: 3306
      db_name: maindb
      username: root
      password: secret
    analytics:
      driver: postgres
      host: localhost
      port: 5432
      db_name: analytics
      username: root
      password: secret
  redis:
    default:
      host: localhost
      port: 6379
      password: ""
      db: 0
```

Each repository receives its specific `*sql.DB` connection via constructor injection — no global state.

---

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

---

## Adding a New Feature

1. **Entity** — add struct in `internal/entity/`
2. **Repository** — define interface + SQL implementation in `internal/repository/`
3. **Service** — implement business logic in `internal/service/`
4. **Usecase** — compose services in `internal/usecase/`
5. **Handler** — add HTTP handler in `internal/handler/http/`, register route in `router.go`
6. **Wire** — connect all dependencies in `internal/app/app.go`

---

*This repository was fully assisted and coded by [Claude](https://claude.ai) (Anthropic).*
