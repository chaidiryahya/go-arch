# CLAUDE.md

This file provides a fast structural overview of the project for Claude Code agents. Agents should read this first before scanning the full codebase.

## Project Overview

- **Project**: go-arch
- **Purpose**: Go microservice skeleton — Clean Architecture / Repository Pattern with multi-database support, graceful shutdown, and clear layer separation
- **Stack**: Go 1.22+ · `database/sql` · `gorilla/mux` · `rs/zerolog` · `go-redis` · YAML config
- **Modes**: `--mode=http|cron|nsq`, selected by a flag at startup

> **Status**: skeleton. The layer structure and wiring are in place; the `example` entity is the reference vertical slice. Read the *Current State* section before trusting a layer to be complete.

## Architecture

```
cmd/app/main.go                  ← Entry point (--mode=http|cron|nsq, -t config check)
internal/
├── app/app.go                   ← Composition root: wiring + graceful shutdown
├── handler/
│   ├── http/                    ← Router, middleware, response envelope
│   ├── cron/                    ← Scheduled job handler
│   └── nsq/                     ← Message consumer handler
├── usecase/                     ← Orchestrates services (1 usecase → N services)
├── service/                     ← Single-responsibility business logic
├── repository/                  ← Data access interfaces + implementations
├── entity/                      ← Pure domain models
└── infra/
    ├── configuration/           ← YAML loader, defaults, validation
    ├── db/                      ← Multi-database connection registry
    ├── redis/                   ← Redis connection registry
    ├── log/                     ← Zerolog structured logger
    └── net/                     ← Reusable HTTP client
library/                         ← Stateless shared utilities
files/config/app.yaml.sample     ← Config template
```

### Layer Rules

| Layer | Responsibility | Rule |
|---|---|---|
| `handler` | Receive request, decode, return response | Calls exactly 1 usecase |
| `usecase` | Orchestrate business flow | Composes N services, never touches repository directly |
| `service` | Single-responsibility business logic | Reusable across usecases, calls repository interfaces |
| `repository` | Data access abstraction | Interface-driven; receives its connection via constructor injection |
| `entity` | Domain models | Plain Go structs, zero dependencies |
| `infra` | Infrastructure concerns | DB, Redis, config, logging |
| `library` | Shared utilities | Stateless helpers, importable by any layer |

### Dependency Flow

```
handler → usecase → service → repository → database
```

Dependencies point inward. Nothing below a layer imports a layer above it.

## Current State (read before building)

This is a skeleton, and the layers are not equally complete. Known gaps:

- **Repository → service → usecase are placeholders.** `internal/app/app.go` wires a nil `ExampleRepository` because no database connection is opened yet. The example route is commented out in `router.go` until that is fixed.
- **Multi-database is configured, not connected.** `resource.database` supports named MySQL + Postgres entries, but `cmd/app/main.go` passes `nil` for the DB and Redis registries.
- **Test coverage is thin.** `library/idgen`, `library/requestid`, `internal/infra/log`, `internal/handler/{http,cron,nsq}`, and `internal/usecase` (cross-layer request-id propagation) have tests; `entity`, `service`, `repository`, `app`, and the remaining infra packages still report "no test files".
- **No local environment.** There is no `docker-compose.yml` or migration tooling, so the multi-DB setup cannot be exercised locally yet.

When a task touches the DB layer, fix the affected gap rather than building on top of the placeholder.

## Key Conventions

- **No global state.** Connections and clients are created in `app.go` and injected. No package-level `var db *sql.DB`.
- **Nondeterminism goes through `library/`.** Time (`time.Now`), IDs (`uuid`), and randomness are to be injected via `library/` seams — today `library/clock` and `library/idgen` (constructed once in `app.go`), never called directly in handlers, services, or infra. `library/*` also ships exported fakes (`clock.NewFake`, `idgen.NewFake`) so other packages' tests can control them. Pure and stdlib-deterministic functions (`time.Parse`, `time.Duration`, `encoding/json`) are not wrapped.
- **Constructor injection.** Every type receives its dependencies through `NewX(...)`; nothing reaches out to a singleton.
- **Errors wrap with context** (`fmt.Errorf("...: %w", err)`) and are classified with `errors.Is`/`errors.As`. Handlers map errors to status codes; layers below return them.
- **Context first.** `ctx context.Context` is the first parameter of anything that does I/O or can block.
- **Graceful shutdown.** `main` selects on a signal channel, drains the HTTP server, then `Application.Close()` releases DB and Redis. In-flight work gets a timeout.
- **Structured logging.** `zerolog` key/value fields; include a request id where one exists.
- **Request ID follows every layer, in every mode.** An ID is generated per unit of work: a UUID v7 rendered as 32 dashless hex chars via `library/idgen`. HTTP gets one from `RequestIDMiddleware` (a client `X-Request-ID` is honoured only if it is 8–64 alphanumerics, else replaced); cron and NSQ derive one per job run / message via `infralog.NewRequestID`. The ID is **never returned to a caller** — no response header, no response body field — it exists only for logs. Layers log through `infralog.ContextLogger(ctx, logger)`, which returns a logger already carrying `request_id`; use it instead of the bare injected logger so the field is always present. `library/requestid` holds the context carrier so no layer imports `handler`.

### Config

- File: `files/config/app.yaml` (git-ignored) · template: `files/config/app.yaml.sample`
- Multi-database: `resource.database.<name>` · multi-redis: `resource.redis.<name>`
- Missing pool settings must default to sane values in code — a zero `max_open_conns` means *unlimited* to `database/sql`.

## Development Workflow

This repo carries the local Claude Code workflow. The single entry point is:

```
/dev <request | ticket ref | path to a spec PDF>
```

`/dev` is an **orchestrator**. It spawns subagents, relays structured reports, and is the only participant that talks to the user:

```
planner-agent → implementer-agent → reviewer-agent (max 3 rounds) → finalizer-agent
```

Do not invoke the agents ad hoc to run a full cycle — that loses the review loop and the human checkpoints.

| Agent | Model | Role | Writes |
|---|---|---|---|
| `planner-agent` | Opus | Phased plan | `.plans/` only |
| `implementer-agent` | Sonnet | Execute a phase + unit tests | source + tests |
| `reviewer-agent` | Opus | Read-only quality gate | nothing |
| `finalizer-agent` | Haiku | Summary, branch, commit, push | `.ai-logs/` + git |

Two standalone verification agents exist outside the `/dev` pipeline and are spawned when a prompt asks for end-to-end verification: `e2e-static-agent` (zero network, traces the code path) and `e2e-runtime-agent` (executes real read-only requests). Neither has a slash command; the main thread does intake and normalizes the test case to `.test-reports/`.

### Artifact locations

| Path | Contents | Committed |
|---|---|---|
| `.plans/YYYYMMDD-slug.md` | Development plans | No (git-ignored) |
| `.ai-logs/YYYYMMDD-slug.md` | Summaries + lessons learned | Yes |
| `.test-reports/` | E2E verification evidence | No (git-ignored) |

### Branch naming

`<prefix>/<short-description>` — `feat/`, `fix/`, `docs/`, `chore/`, `refactor/`, `test/`, `perf/`. The finalizer derives the branch from the plan slug.

### MR/PR

The finalizer detects `gh` → `GITHUB_TOKEN`/`GH_TOKEN` → compare link. Missing setup is reported as `MR_SETUP: missing`, not a failure. Tokens are detected by presence only (`${VAR:+set}`) and never echoed.

## Key Commands

```bash
make run          # go run ./cmd/app/ --mode=http
make run-cron     # go run ./cmd/app/ --mode=cron
make run-nsq      # go run ./cmd/app/ --mode=nsq
make config-test  # validate config and exit
make build        # go build -o bin/app ./cmd/app/
make test         # go test ./...
make vet          # go vet ./...
```

> **macOS note:** build with `CGO_ENABLED=0` to avoid the Go 1.22.6 + recent Darwin dyld incompatibility.

## Adding a New Feature

1. **Entity** — add the struct in `internal/entity/`
2. **Repository** — define the interface + SQL implementation in `internal/repository/`
3. **Service** — business logic in `internal/service/`, inject the repository interface
4. **Usecase** — compose services in `internal/usecase/`, inject service interfaces
5. **Handler** — add the HTTP handler in `internal/handler/http/`, register the route in `router.go`
6. **Wire** — connect the dependencies in `internal/app/app.go`

## Dependencies

| Package | Purpose |
|---|---|
| `gorilla/mux` | HTTP routing (archived upstream — evaluate `chi` before relying on it long-term) |
| `rs/zerolog` | Structured logging |
| `redis/go-redis/v9` | Redis client |
| `go-sql-driver/mysql` | MySQL driver |
| `lib/pq` | PostgreSQL driver (maintenance mode — `pgx` is the successor) |
| `gopkg.in/yaml.v3` | Config parsing |
| `google/uuid` | Request ID generation |

## Last Updated

- **Date**: 2026-11-26
- **By**: Request ID now follows handler → usecase → service → repository in HTTP, cron, and NSQ via `infralog.NewRequestID`/`ContextLogger`; `idgen` switched to UUID v7 as 32 dashless hex chars; ID is internal only (removed from response header and body); fixed `server_process_time` always reporting ~0.
