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
| Runtime dependency | None | None |

Choose based on deployment and contribution preferences. The test suites consume
the same sanitized PSI 13.4 and CrUX fixtures to prevent response-contract drift.
