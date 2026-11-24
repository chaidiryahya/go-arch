# Nondeterminism Seams (clock + idgen) — Summary

**Date**: 2026-11-25
**Plan**: .plans/20261125-nondeterminism-seams.md
**Review rounds**: 2

## TL;DR
Added two leaf packages under `library/` — `clock` (time) and `idgen` (IDs) — each shipping a stateless real implementation and an exported, concurrency-safe fake for use from any package's tests. Every direct `time.Now`, `time.Since`, and `uuid.New` call site in the HTTP handler package and the reusable HTTP client now goes through an injected seam. Behavior is unchanged: same `X-Request-ID` format, same `server_process_time` values, same logged durations. No tests were added (deferred by user decision) and no new dependencies were introduced.

## What Was Done
- **New `library/clock`**: `Real` (zero-value usable, `Now()` returns `time.Now()` untouched so the monotonic reading — and thus `Now().Sub(start)` ≡ `time.Since(start)` — is preserved) and `Fake` (`NewFake`, `Set`, `Advance`, mutex-guarded for concurrent use).
- **New `library/idgen`**: `Real` (`New() string` over `uuid.NewString()`, UUID v4) and `Fake` (`NewFake`, mutex-guarded sequence, emits v4-shaped values starting at `...000000000001`) so tests that assert UUID format or echo the header behave realistically.
- **Ratified Decision 1 — consumer-side, package-private interfaces**: `internal/handler/http` declares `clock` and `idGenerator`; `internal/infra/net` declares its own local `clock`. The provider packages export concrete types only; consumers are structurally satisfied without importing `library/*`.
- **Ratified Decision 2 — `idgen.New() string`, panic on entropy failure**: matches the prior `uuid.New().String()` behavior; a dead CSPRNG is not recoverable in middleware. Documented in godoc.
- **Ratified Decision 3 — UUID v4, not v7**: preserves byte-for-byte request-ID behavior and avoids embedding a timestamp in a client-visible `X-Request-ID`. v7 remains available later behind a distinct method if a DB-primary-key use case appears.
- **Call sites converted**:
  - `middleware.go` — `RequestIDMiddleware(ids)` is now curried (`func(http.Handler) http.Handler`) so `r.Use` still works; `LoggingMiddleware(logger, clk)` and `RecoveryMiddleware(logger, clk)` take the clock; `GetStartTime(ctx, clk)` fallback converted.
  - `response.go` — `writeJSON`/`WriteSuccess`/`WriteError` thread `clk` and use `clk.Now().Sub(startTime)`.
  - `health_handler.go` — `NewHealthHandler(clk)` stores and uses the clock.
  - `router.go` — `NewRouter(logger, clk, ids, handlers)` wires all three middleware with seams.
  - `httpclient.go` — local `clock` interface, `NewHTTPClient(..., clk)`; both timing calls converted.
  - `app.go` — constructs `clock.Real{}` and `idgen.Real{}` exactly once and injects them; no package-level mutable state.
- **`CLAUDE.md`** — added the "Nondeterminism goes through `library/`" convention bullet and refreshed the Last Updated stanza.
- **Verification**: `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go vet ./...` pass; `git diff go.mod go.sum` is empty; a repo-wide grep confirms the only `time.Now` / `time.Since` / `uuid.New` hits are inside `library/clock` and `library/idgen`.

## Files Changed
| File | Change |
|---|---|
| `library/clock/clock.go` | New — `Real` + `Fake` (`NewFake`/`Set`/`Advance`), monotonic-preserving `Now`, compile-time assertions |
| `library/idgen/idgen.go` | New — `Real` (UUID v4) + deterministic `Fake`, compile-time assertions |
| `internal/handler/http/middleware.go` | Package-private `clock`/`idGenerator` interfaces; curried `RequestIDMiddleware`; clock threaded into logging, recovery, `GetStartTime` |
| `internal/handler/http/response.go` | `writeJSON`/`WriteSuccess`/`WriteError` take `clk`; `server_process_time` via `clk.Now().Sub(startTime)` |
| `internal/handler/http/health_handler.go` | `NewHealthHandler(clk)` stores the clock |
| `internal/handler/http/router.go` | `NewRouter(logger, clk, ids, handlers)` wires the seams into middleware |
| `internal/infra/net/httpclient.go` | Local `clock` interface; `NewHTTPClient` takes `clk`; both timing calls converted |
| `internal/app/app.go` | Constructs `clock.Real{}` / `idgen.Real{}` once and injects them |
| `CLAUDE.md` | New nondeterminism convention bullet; Last Updated refreshed |
| `.ai-logs/20261125-nondeterminism-seams.md` | This summary |

## Lessons Learned
- **Review / documentation accuracy**: the review caught a false claim in `CLAUDE.md` about agent model pins. Documentation that asserts configuration facts must be checked against the real configuration file before being committed — a plausible-sounding bullet is still a bug if it is wrong.
- **Subagent configuration**: subagent model pinning must match what the gateway key actually permits. A pin the key cannot serve fails at spawn time, not at review time, so verify the pin against the provider's allowance before relying on it.

## Review Notes
- No tests were added; this was explicitly deferred by user decision. The fakes exist so later test work can be deterministic without another seam change.
- Behavior is unchanged by design; the diff is a pure dependency-injection refactor (~130–150 LOC plus godoc).
- Trade-off accepted (Decision 1): `clock`/`idGenerator` are unexported yet appear in exported constructor signatures (`NewRouter`, `NewHealthHandler`, `WriteSuccess`, `WriteError`). Legal Go, and cross-package callers can still inject structurally; the interface type just cannot be named. Revisit only if the team prefers exported `Clock`/`IDGenerator`.
- `NewHTTPClient` has no call sites today, so its signature change is invisible to the compiler-checked graph — wiring it will be a deliberate later step.
- Switching to UUID v7 later is a one-file change in `library/idgen`, behind a distinct method so request-ID semantics are not silently altered.
