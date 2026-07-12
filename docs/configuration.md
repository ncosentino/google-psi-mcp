---
description: Configure API access, STDIO or HTTP transport, ports, and allowed hosts.
---

# Configuration

## API key

The server resolves one Google API key in this order:

1. `--api-key`
2. `GOOGLE_PSI_API_KEY`
3. `GOOGLE_PSI_API_KEY` in a `.env` file in the working directory

The PageSpeed Insights API must be enabled. The direct CrUX tools additionally
require the Chrome UX Report API.

When API restrictions are enabled on the key, allow both APIs if both tool
families will be used. A PageSpeed-only restriction causes `get_crux_data` and
`get_crux_history` to return `PERMISSION_DENIED`.

## STDIO transport

STDIO is the default:

```bash
./psi-mcp-go-linux-amd64 --api-key YOUR_KEY
```

Explicit form:

```bash
./psi-mcp-go-linux-amd64 --transport stdio --api-key YOUR_KEY
```

## HTTP transport

```bash
./psi-mcp-go-linux-amd64 \
  --transport http \
  --listen-address 127.0.0.1 \
  --port 8080 \
  --allowed-hosts localhost,127.0.0.1
```

Both implementations accept:

- `--listen-address`, falling back to `MCP_LISTEN_ADDRESS` and then `127.0.0.1`
- `--port`, falling back to `PORT` and then `8080`

The MCP endpoint is `/mcp`. The health endpoint is `/health`.

Go uses the comma-separated `--allowed-hosts` option. C# uses standard ASP.NET
Core `AllowedHosts` configuration, whose default is:

```text
localhost;127.0.0.1;[::1]
```

The HTTP transport is stateless and includes request-size, host, and
cross-origin protections. It does not authenticate callers. Keep the default
loopback listener unless an authenticated reverse proxy or private network
protects the service.

## Analysis limits

- Maximum URLs per `analyze_pages` call: 10
- Maximum concurrent PSI requests per process: 4
- Transient attempts: 3
- PSI HTTP timeout: 120 seconds
- CrUX HTTP timeout: 30 seconds

The process-wide limit is shared by every connected HTTP client and every STDIO
request handled by that process.
