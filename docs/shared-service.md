---
description: Run one PageSpeed Insights MCP web service shared by every agent session.
---

# Running One Shared Service

STDIO makes each MCP client launch its own server process. Streamable HTTP lets
every local agent session connect to one long-lived process instead.

The server is stateless at the MCP transport layer. It does not require session
affinity, and all connected clients share one process-wide PageSpeed request
limit.

## Endpoints

| Endpoint | Purpose |
|---|---|
| `http://127.0.0.1:8080/mcp` | Streamable HTTP MCP |
| `http://127.0.0.1:8080/health` | Supervisor health and version metadata |
| `http://127.0.0.1:8080/shutdown` | Manager-authenticated graceful shutdown |

The listener defaults to `127.0.0.1`. Use `--listen-address` only when network
access is intentional. The included service manager is deliberately
loopback-only; network deployments should use a platform service supervisor.

## Prepare credentials

The shared process must receive the Google API key itself. Client-side HTTP MCP
configuration does not launch the server and therefore cannot supply its
environment.

Set `GOOGLE_PSI_API_KEY` in the service environment or place it in a `.env`
file beside the binary. Do not put the key into the HTTP client configuration.

## Manage the process

The repository and GitHub release include a reusable PowerShell service
manager:

```powershell
$manager = ".\scripts\manage-mcp-service.ps1"
$binary = "C:\path\to\psi-mcp-go.exe"

& $manager Start -BinaryPath $binary
& $manager Status -BinaryPath $binary
& $manager Restart -BinaryPath $binary
& $manager Stop -BinaryPath $binary
```

Windows PowerShell is sufficient on Windows. Linux and macOS require
PowerShell 7 (`pwsh`) and `nohup`.

`Start` is idempotent: it checks `/health` and reuses the existing process.
Concurrent start attempts are serialized so multiple agent sessions do not
launch duplicate servers.

Runtime state and logs are stored under the current user's local application
data directory:

```text
shared-mcp-services/google-psi-mcp/
```

The state file records the PID, process creation time, executable path, and a
per-run shutdown credential. `Stop` verifies the process identity, requests a
graceful shutdown, and only falls back to terminating that verified process.
This prevents a stale PID from targeting an unrelated application. On Unix,
the state directory and file are restricted before the shutdown credential is
persisted.

To replace a stopped or running binary:

```powershell
& $manager Update `
  -BinaryPath "C:\path\to\psi-mcp-go.exe" `
  -ReplacementBinaryPath "C:\downloads\psi-mcp-go-windows-amd64.exe"
```

The manager works with either the Go or C# binary because both accept the same
transport, address, and port arguments.

## Configure Copilot CLI

Replace the STDIO entry with an HTTP entry:

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

The `command`, `args`, and `env` properties are removed because Copilot no
longer launches the server. Restart existing Copilot sessions after changing
the configuration.

STDIO remains available for clients that cannot use Streamable HTTP.

## Start automatically

Run the manager's `Start` action from a user-level startup mechanism such as:

- a Copilot `sessionStart` hook
- Windows Task Scheduler
- a systemd user service
- a launchd user agent

Repeated calls are inexpensive health checks and will not create another
server while the existing instance is healthy.

## Concurrency and quotas

The limit of four concurrent PageSpeed requests is process-wide. Calls from
different MCP clients queue behind the same limiter rather than multiplying
the upstream load by the number of agent sessions.

Each `analyze_pages` call still accepts at most ten URLs and preserves input
order in its response.

## Network deployment

Loopback hosting is intended for a single user's trusted local processes. For
any non-loopback listener:

1. Terminate TLS at the service or a trusted reverse proxy.
2. Authenticate and authorize every MCP request.
3. Validate `Origin` at the ingress and preserve the effective scheme and host
   when forwarding accepted requests.
4. Apply request, connection, and rate limits at the ingress.
5. Trust forwarded headers only from a proxy that replaces untrusted values.

The built-in HTTP host does not implement OAuth and should not be exposed
directly to the internet.
