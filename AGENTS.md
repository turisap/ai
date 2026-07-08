# AGENTS.md — `mcp-task-server`

## What this is

A Go MCP (Model Context Protocol) server implementing the Streamable HTTP
transport. Serves a single endpoint `/mcp` — `GET` for SSE streams, `POST` for
JSON-RPC tool calls.

## Key structure

```
cmd/server/main.go         — wiring: config, pgxpool, redis, HTTP server
internal/mcp/              — MCP protocol types, registry, session, server
internal/tools/tools.go    — tool definitions + real DB handlers (closures)
docker/docker-compose.yml  — postgres:16-alpine + redis:7-alpine
Taskfile.yml               — all dev commands (Task runner, not make)
```

## Developer commands (use `task`, not ad-hoc)

| `task` target | What it does |
|---|---|
| `up` | Start postgres + redis containers |
| `run` | `up` + `go run ./cmd/server` (reads .env) |
| `build` | Build binary to `./bin/server` |
| `lint` | `golangci-lint run ./...` |
| `test` | `go test ./...` (no tests exist yet) |
| `smoke` | `initialize` + `tools/list` via curl |
| `psql` | `psql $DATABASE_URL` |
| `migrate:up` | `goose -dir migrations postgres "$DATABASE_URL" up` |

No `migrations/` directory exists yet. `go test ./...` passes with 0 tests.

## Config (100% env-var driven via `cleanenv`)

`.env` file is auto-loaded from CWD (falls back to env vars silently).

| Env var | Default | Notes |
|---|---|---|
| `ADDR` | `:8080` | Server listen address |
| `DATABASE_URL` | `postgres://tasks:tasks@localhost:5432/tasks` | pgx/v5 pool |
| `REDIS_URL` | `redis://localhost:6379` | go-redis/v9 |
| `LOG_LEVEL` | `debug` | `debug` or `info` |

## HTTP transport quirk

Single endpoint `/mcp`:
- **GET /mcp** — opens an SSE stream. The first event (`endpoint`) tells the
  client where to POST (`/mcp?sessionId=<uuid>`). Required for MCP clients that
  use the HTTP+SSE transport.
- **POST /mcp** — synchronous JSON-RPC. Response goes in the HTTP body (not
  over SSE). This is what tool calls actually use.

For Claude Desktop / OpenCode config use `"url": "http://localhost:8080/mcp"`
(not `localhost:8080`).

## Tools are wired by injected DB handles

All three tools (`get_tasks`, `get_task_counters`, `create_task`) are registered
in `tools.RegisterAll()` as closures over the `*pgxpool.Pool` and `*redis.Client`.
`create_task` uses a transaction + outbox row atomically.

To add a tool: define the `mcp.Tool`, write the handler, call `registry.Register()`
in `RegisterAll()`.

## MCP protocol

Implements: `initialize`, `initialized`, `tools/list`, `tools/call`, `ping`.
Protocol version: `2024-11-05`. No auth, no rate limiting, no audit log yet.

## No dev infra

- No CI, no test fixtures, no snapshot tests
- No pre-commit hooks
- No gitignore beyond `.idea/` and `.env`
- `WriteTimeout` is intentionally omitted (SSE connections are long-lived)
