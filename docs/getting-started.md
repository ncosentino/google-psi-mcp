# Getting Started

Get up and running in three steps: get an API key, download a binary, and configure your AI tool.

---

## Step 1 -- Get a PageSpeed Insights API Key

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create or select a project
3. Navigate to **APIs & Services → Library**
4. Search for **PageSpeed Insights API** and enable it
5. Navigate to **APIs & Services → Credentials**
6. Click **Create Credentials → API Key**
7. Copy the key

!!! tip "Restrict your API key"
    In Credentials, click on the key you just created and add an **API restriction** to the PageSpeed Insights API only. This limits blast radius if the key is ever exposed.

!!! note "The PSI API is free"
    There is no billing required and no quota you're likely to hit during normal interactive use. Google's standard PageSpeed Insights quota is 25,000 requests per day per project.

---

## Step 2 -- Download a Binary

Pre-built binaries are available for Linux, macOS, and Windows in both Go and C# implementations.

| Platform | Go binary | C# binary |
|----------|-----------|-----------|
| Linux x64 | `psi-mcp-go-linux-amd64` | `psi-mcp-csharp-linux-x64` |
| Linux ARM64 | `psi-mcp-go-linux-arm64` | `psi-mcp-csharp-linux-arm64` |
| macOS x64 | `psi-mcp-go-darwin-amd64` | `psi-mcp-csharp-osx-x64` |
| macOS ARM64 | `psi-mcp-go-darwin-arm64` | `psi-mcp-csharp-osx-arm64` |
| Windows x64 | `psi-mcp-go-windows-amd64.exe` | `psi-mcp-csharp-win-x64.exe` |

Download from [GitHub Releases](https://github.com/ncosentino/google-psi-mcp/releases). For most users, the Go binary is the better choice -- it's smaller (8-15 MB vs 20-40 MB) and starts faster.

After downloading, make the binary executable on Linux/macOS:

```bash
chmod +x psi-mcp-go-linux-amd64
```

---

## Step 3 -- Configure Your AI Tool

Set the `GOOGLE_PSI_API_KEY` environment variable and point your AI tool at the binary.

See [Setup by Tool](setup-by-tool/) for full configuration snippets for Claude Desktop, GitHub Copilot CLI, Cursor, Zed, and others.

!!! tip "Timeout settings"
    The PSI API can take 10-20+ seconds per URL. If your MCP client supports a custom timeout, set it to at least **60 seconds**. Using `strategy="both"` on slow pages can approach that limit.

