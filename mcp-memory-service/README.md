## Setup

* `brew install python`
* `python3 -m venv .venv  && source .venv/bin/activate`
* `pip3 install mcp-memory-service`
* `deactivate` when finish

## Usage

* `MCP_ALLOW_ANONYMOUS_ACCESS=true memory server --http`

## OpenCode plugin

```bash
git clone https://github.com/doobidoo/mcp-memory-service.git
cd mcp-memory-service
mkdir -p ~/.config/opencode/plugins
cp opencode/memory-plugin.js ~/.config/opencode/plugins/
```

## OpenCode `/memory` command

```
{
  "command": {
    "memory": {
      "description": "Show MCP Memory Service status. Usage: /memory, /memory search <query>, /memory health",
      "template": ""
    }
  }
}
```

https://github.com/doobidoo/mcp-memory-service/blob/main/opencode/README.md
