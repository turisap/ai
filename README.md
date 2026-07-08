# mcp-task-server

A production-grade MCP (Model Context Protocol) server in Go that exposes your task domain to LLM agents.

## Quick start

```bash
# Run dependencies
docker compose -f docker/docker-compose.yml up postgres redis -d

# Run server
go run ./cmd/server

# Verify — should return tool list
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl","version":"0.1"},"capabilities":{}}}' | jq .

curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | jq .

# Call a tool
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_tasks","arguments":{"store_id":"some-uuid","limit":5}}}' | jq .
```

## Connect to Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "task-server": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

Restart Claude Desktop. Your tools will appear in the tool picker.

## Connect to OpenCode

Add to your OpenCode config:

```json
{
  "mcp": {
    "servers": {
      "task-server": {
        "url": "http://localhost:8080/mcp"
      }
    }
  }
}
```

## Project structure

```
cmd/server/main.go          — entrypoint, wires everything
internal/mcp/
  types.go                  — JSON-RPC + MCP protocol types
  session.go                — per-client SSE session
  registry.go               — tool registry + dispatcher
  server.go                 — HTTP server, initialize, tools/list, tools/call
internal/tools/
  tools.go                  — tool definitions + stub handlers (wire real DB here)
docker/docker-compose.yml
```

## Wiring real data (week 2 task)

Each tool handler in `internal/tools/tools.go` has a `// TODO` comment showing
exactly what query to run. Steps:

1. Add `pgx/v5` pool to `main.go`, pass it into `tools.RegisterAll`
2. Add `go-redis/v9` client similarly
3. Replace stub bodies with the real queries shown in the TODO comments
4. For `create_task`: wrap the insert in a transaction with an outbox row

## Adding a new tool

```go
// 1. Define it
var toolAssignTask = mcp.Tool{
    Name:        "assign_task",
    Description: "Reassign a task to a different user.",
    InputSchema: mcp.InputSchema{
        Type: "object",
        Properties: map[string]mcp.Property{
            "task_id":     {Type: "string", Description: "UUID of the task"},
            "assignee_id": {Type: "string", Description: "UUID of the new assignee"},
        },
        Required: []string{"task_id", "assignee_id"},
    },
}

// 2. Implement the handler
func handleAssignTask(ctx context.Context, raw json.RawMessage) (mcp.ToolCallResult, error) { ... }

// 3. Register in RegisterAll()
registry.Register(toolAssignTask, handleAssignTask)
```

## Roadmap

- [ ] Wire real Postgres queries (get_tasks, create_task)
- [ ] Wire real Redis counters (get_task_counters)  
- [ ] Outbox transaction for create_task / assign_task
- [ ] API key auth middleware
- [ ] Per-session rate limiting (sliding window in Redis)
- [ ] Tool call audit log (agent_activity table)


### PLAN
1. Read the transport page (~20 min)
   Go to modelcontextprotocol.io/docs/concepts/transports. The key things to understand:

The server must provide a single HTTP endpoint that supports both POST and GET methods. The client uses HTTP POST to send JSON-RPC messages, and must include an Accept header listing both application/json and text/event-stream. That's the whole transport in one sentence. Model Context Protocol
HTTP+SSE requires two separate endpoints (GET /mcp for SSE and POST /messages for requests), while Streamable HTTP uses a single endpoint (POST /mcp) but with more complex request/response patterns. Build Streamable HTTP — it's the future. Simplescraper

2. Read the tools/call spec (~15 min)
   At modelcontextprotocol.io/docs/concepts/tools. It's just JSON-RPC: client sends tools/call with a tool name + arguments, server returns a result. That's 90% of what you'll implement.
3. Look at one real Go implementation (~20 min)
   Search GitHub for mcp-go or mark3labs/mcp-go. Reading ~200 lines of a real Go MCP server will teach you more than the spec alone. You'll see the session management, the JSON-RPC dispatch loop, the SSE event loop.
4. Build initialize + tools/list yourself first
   Don't reach for a library yet. Wire up a bare net/http server that responds to those two methods with hardcoded JSON. Once it works with a real client (Claude Desktop or OpenCode), you'll have internalized the protocol. Then optionally switch to a library.

### Tools
`get_tasks (already scaffolded)`
Query tasks by store, status, assignee. The thing you currently do with a raw psql query a dozen times a day.
`get_task_counters (already scaffolded)`
Pull counter state from Redis. Instantly answer "does my CDC projection match reality?"
`get_store_state`
Given a store ID, return its manager, assistant, active task count, any open autotasks. Replaces the 3-join query you write when debugging a store-specific bug.
`explain_task`
Given a task ID, return full context: title, status, assignee, store, history of status changes, related outbox events. The "what happened to this task" query you run during incidents.
Tier 2 — useful several times a week
`check_outbox_lag`
Query the outbox table for unprocessed events older than N seconds. Answers "is my relay worker stuck?" in one call instead of you opening psql.
`get_recent_errors`
Pull last N failed outbox events or dead-letter entries. Incident debugging without opening Kibana/logs manually.
find_tasks_by_assignee
Given a user ID, return all their open tasks across stores. You currently do this with a JOIN — the agent can do it and summarize patterns ("this user has 40 open tasks, mostly autotasks").
`get_store_managers`
List stores missing a manager or assistant — the exact validation you check before task creation fails with your "performer not found" logic (conv 21).

Tier 3 — high value once you hit week 3+
`simulate_task_creation`
Dry-run task creation for a given store — validate performer resolution, check constraints, return what would happen without writing anything. Great for testing edge cases without polluting the DB.
`check_index_usage`
Wrap your pg_stat_user_indexes query (conv 36) — the agent can proactively flag unused indexes when you're about to add a new one.
`get_cdc_consumer_lag`
Query Redis for your CDC projection version vs the latest DB transaction ID. Answers "is my counter cache stale?" which you currently have to piece together manually.
`search_tasks`
Full-text search over task titles/descriptions. Useful when debugging "find me all tasks that mention X" without writing a LIKE query.


### Setup
* `brew install python@3.12 uv`
* `uv tool install graphifyy`
* `graphify install`
* `graphify install --platform opencode`
* `graphify opencode install` - for opencode always use the knoledge graph

### Commands
[List](https://github.com/safishamsi/graphify#full-command-reference)
* `/graphify ./raw --mode deep`        # more aggressive relationship extraction
* ```
  graphify hook install              # post-commit + post-checkout hooks
  graphify hook uninstall
  graphify hook status
  ```


### Codegraph
* `brew install node`
* `npm i -g @colbymchenry/codegraph`
* `codegraph install`

do not commit to git, rebuild locally