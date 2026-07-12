---
description: Enable Google APIs, download a native binary, and configure the MCP server.
---

# Getting Started

## 1. Configure Google APIs

In Google Cloud Console:

1. Create or select a project.
2. Enable the PageSpeed Insights API.
3. Create an API key.
4. Enable the Chrome UX Report API if you need direct current or historical
   CrUX tools.
5. Restrict the key to the APIs it must call.

No OAuth flow or billing account is required.

## 2. Download a binary

Download the appropriate Go or C# Native AOT binary from
[GitHub Releases](https://github.com/ncosentino/google-psi-mcp/releases).

On Linux or macOS:

```bash
chmod +x psi-mcp-go-linux-amd64
```

## 3. Configure your MCP client

```json
{
  "mcpServers": {
    "pagespeed-insights": {
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

Restart the MCP client after changing its configuration.

## 4. Test the installation

```text
Analyze https://www.devleader.ca on mobile and separate the real-user field
data from the Lighthouse lab findings.
```

For CrUX:

```text
Get current phone CrUX data for the origin https://www.devleader.ca.
```

If the PSI call works but CrUX returns `PERMISSION_DENIED`, update the API
enablement or key restrictions for the Chrome UX Report API.
