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