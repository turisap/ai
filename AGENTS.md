# AGENTS.md — `mcp-memory-service` launcher repo

## What this repo is

A minimal wrapper that provides a local Python venv with the
[`mcp-memory-service`](https://github.com/doobidoo/mcp-memory-service) PyPI
package (v10.69.0). **No source code lives here.** The actual package is
installed at `.venv/lib/python3.14/site-packages/mcp_memory_service/`.

## CLI

```bash
# Activate
source .venv/bin/activate

# Default MCP stdio server (for Claude Code / Claude Desktop stdio transport)
MCP_ALLOW_ANONYMOUS_ACCESS=true memory server

# HTTP REST API + dashboard (http://localhost:8000)
MCP_ALLOW_ANONYMOUS_ACCESS=true memory server --http

# Lifecycle helpers (background process management)
memory launch
memory stop
memory info
memory health
# ...
```

## Key configuration (100% env-var driven)

Config module: `.venv/.../mcp_memory_service/config.py`

| Env var | Default | Notes |
|---|---|---|
| `MCP_ALLOW_ANONYMOUS_ACCESS` | `false` | Set to `true` to skip auth |
| `MCP_MEMORY_STORAGE_BACKEND` | `sqlite_vec` | Also: `cloudflare`, `hybrid`, `milvus` |
| `MCP_MEMORY_SQLITE_PATH` | `~/Library/Application Support/mcp-memory/sqlite_vec.db` | |
| `MCP_HTTP_HOST` / `MCP_HTTP_PORT` | `127.0.0.1` / `8000` | |
| `MCP_SSE_HOST` / `MCP_SSE_PORT` | `127.0.0.1` / `8765` | |
| `MCP_CONSOLIDATION_ENABLED` | `false` | Dream-inspired memory consolidation |
| `MCP_QUALITY_SYSTEM_ENABLED` | `true` | Quality scoring for memories |
| `MCP_OAUTH_ENABLED` | `false` | OAuth 2.1 (RS256/HS256) |

`.env` files are auto-loaded from CWD or standard paths (config.py:36-67).

## No dev tooling

No tests, linting, typechecking, formatters, CI, or build steps. The package
is consumed from PyPI, not developed here. To inspect behavior, read the
installed source at `.venv/lib/python3.14/site-packages/mcp_memory_service/`.

## Heavy ML deps

`torch` / `transformers` are lazy-loaded at first use (not at import time) to
keep CLI startup fast. Lifecycle commands (`memory launch` / `stop` / `info` /
`health`) import only stdlib + click.
