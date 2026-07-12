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
PORT=8080 ./psi-mcp-go-linux-amd64 \
  --transport http \
  --allowed-hosts localhost,127.0.0.1
```

Both implementations read the listen port from `PORT`, defaulting to `8080`.

Go uses the comma-separated `--allowed-hosts` option. C# uses standard ASP.NET
Core `AllowedHosts` configuration, whose default is:

```text
localhost;127.0.0.1;[::1]
```

The HTTP transport is stateless and includes host and cross-origin protections.
It does not authenticate callers. Use an authenticated reverse proxy or private
network for any non-local deployment.

## Analysis limits

- Maximum URLs per `analyze_pages` call: 10
- Maximum concurrent PSI requests: 4
- Transient attempts: 3
- PSI HTTP timeout: 120 seconds
- CrUX HTTP timeout: 30 seconds
