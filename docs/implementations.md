---
description: Compare the parity-tested Go and C# Native AOT implementations.
---

# Go vs C#

Both implementations expose the same four tools, validation rules, output
contract, retry policy, batch limits, and transports.

| Aspect | Go | C# |
|---|---|---|
| Runtime | Native Go binary | .NET Native AOT |
| MCP SDK | `go-sdk` 1.6.1 | `ModelContextProtocol` 1.4.1 |
| Transport | STDIO and Streamable HTTP | STDIO and Streamable HTTP |
| HTTP mode | Stateless | Stateless |
| Default listener | `127.0.0.1:8080` | `127.0.0.1:8080` |
| HTTP endpoints | `/mcp` and `/health` | `/mcp` and `/health` |
| PSI concurrency | Four per process | Four per process |
| Runtime dependency | None | None |

Choose based on deployment and contribution preferences. The test suites consume
the same sanitized PSI 13.4 and CrUX fixtures to prevent response-contract drift.
The shared PowerShell service manager can operate either binary.
