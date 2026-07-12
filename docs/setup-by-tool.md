---
description: Configure the Google PageSpeed Insights MCP server in GitHub Copilot CLI, Claude Desktop, Cursor, VS Code, Visual Studio, or via a .env file.
---

# Setup by Tool

Configuration snippets for MCP clients. Enable PageSpeed Insights for PSI tools
and the Chrome UX Report API for direct CrUX tools.

---

## GitHub Copilot CLI / Claude Code

Add to `.mcp.json` in your project or home directory:

```json
{
  "mcpServers": {
    "psi": {
      "type": "stdio",
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "args": [],
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key"
      }
    }
  }
}
```

!!! note
    Some MCP clients (including GitHub Copilot CLI) require `"args": []` when `"type": "stdio"` is specified. Claude Desktop does not require it.

---

## Claude Desktop

Add to `claude_desktop_config.json` (`~/Library/Application Support/Claude/` on macOS, `%APPDATA%\Claude\` on Windows):

```json
{
  "mcpServers": {
    "psi": {
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key"
      }
    }
  }
}
```

---

## Cursor

Add to `.cursor/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "psi": {
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "args": [],
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key"
      }
    }
  }
}
```

---

## VS Code

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "psi": {
      "type": "stdio",
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "args": [],
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key"
      }
    }
  }
}
```

---

## Visual Studio

Add to the MCP configuration in Visual Studio's GitHub Copilot settings:

```json
{
  "mcpServers": {
    "psi": {
      "type": "stdio",
      "command": "C:\\path\\to\\psi-mcp-go-windows-amd64.exe",
      "args": [],
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key"
      }
    }
  }
}
```

---

## Using a .env File

Place a `.env` file in the same directory as the binary:

```env
GOOGLE_PSI_API_KEY=your-api-key
```

Then the `env` block in your tool config can be omitted. The binary reads `.env` automatically from its working directory.

---

## Using a CLI Argument

Pass the key directly on the command line:

```bash
./psi-mcp-go-linux-amd64 --api-key your-api-key
```

See [Configuration](configuration.md) for the full resolution order.

---

## Streamable HTTP

Start either binary with HTTP transport:

```bash
./psi-mcp-go-linux-amd64 \
  --transport http \
  --listen-address 127.0.0.1 \
  --port 8080 \
  --allowed-hosts localhost,127.0.0.1 \
  --api-key your-api-key
```

Configure an HTTP-capable MCP client with:

```json
{
  "mcpServers": {
    "pagespeed-insights": {
      "type": "http",
      "url": "http://127.0.0.1:8080/mcp",
      "tools": ["*"]
    }
  }
}
```

All agent sessions can share this one process. See
[Running One Shared Service](shared-service.md) for lifecycle management.

The transport is stateless but does not authenticate callers. Keep it on
loopback, a private network, or behind an authenticated reverse proxy.
