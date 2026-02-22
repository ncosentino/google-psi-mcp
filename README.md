# PageSpeed Insights MCP Server -- Google Core Web Vitals for AI Assistants

[![Latest Release](https://img.shields.io/github/v/release/ncosentino/google-psi-mcp?style=flat-square)](https://github.com/ncosentino/google-psi-mcp/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](go/go.mod)
[![.NET Version](https://img.shields.io/badge/.NET-10-512BD4?style=flat-square&logo=dotnet)](csharp/Directory.Build.props)
[![CI](https://img.shields.io/github/actions/workflow/status/ncosentino/google-psi-mcp/ci.yml?label=CI&style=flat-square)](https://github.com/ncosentino/google-psi-mcp/actions/workflows/ci.yml)

> **Zero-dependency MCP server for Google PageSpeed Insights Core Web Vitals.**
> Pre-built native binaries for Linux, macOS, and Windows. No Node.js. No Python. No .NET runtime. No Go toolchain. Download one binary and configure your AI tool.

Expose Google PageSpeed Insights analysis (Largest Contentful Paint, Cumulative Layout Shift, First Contentful Paint, Time to First Byte, Total Blocking Time) directly to AI assistants like Claude, GitHub Copilot, and Cursor via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io). Ask your AI to diagnose Core Web Vitals issues and get actionable, code-level recommendations.

---

## Why This Exists

AI assistants are powerful at diagnosing web performance problems -- but they need real data. This MCP server bridges your AI tool to Google's PageSpeed Insights API v5, giving it:

- **Real Core Web Vitals** (LCP, CLS, FCP, TTFB, TBT, Speed Index) with ratings (good/needs-improvement/poor) per Google's official thresholds
- **Category scores** (performance, SEO, accessibility, best-practices) on a 0-100 scale
- **Prioritized opportunities** with estimated savings
- **Failing audits** with specific descriptions and current values

With this MCP server configured, you can ask your AI: _"Analyze my homepage on mobile and desktop and tell me what's hurting my Core Web Vitals score"_ and get a structured, actionable answer grounded in real lighthouse data.

---

## Quick Start

**Three steps: get an API key, download a binary, add it to your MCP config.**

### Step 1: Get a Google PageSpeed Insights API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/library/pagespeedonline.googleapis.com)
2. Enable the **PageSpeed Insights API**
3. Create an API key (no billing required for PSI)

> The PSI API is free with generous quotas. No billing account is needed.

### Step 2: Download a Binary

Go to the [Releases page](https://github.com/ncosentino/google-psi-mcp/releases/latest) and download the binary for your platform:

| Platform | Go binary | C# binary |
|----------|-----------|-----------|
| Linux x64 | `psi-mcp-go-linux-amd64` | `psi-mcp-csharp-linux-x64` |
| Linux arm64 | `psi-mcp-go-linux-arm64` | `psi-mcp-csharp-linux-arm64` |
| macOS x64 (Intel) | `psi-mcp-go-darwin-amd64` | `psi-mcp-csharp-osx-x64` |
| macOS arm64 (Apple Silicon) | `psi-mcp-go-darwin-arm64` | `psi-mcp-csharp-osx-arm64` |
| Windows x64 | `psi-mcp-go-windows-amd64.exe` | `psi-mcp-csharp-win-x64.exe` |
| Windows arm64 | `psi-mcp-go-windows-arm64.exe` | `psi-mcp-csharp-win-arm64.exe` |

See [Go vs C# -- Which Binary?](#go-vs-c----which-binary) if you're unsure which to pick.

On Linux/macOS, make the binary executable after downloading:

```bash
chmod +x psi-mcp-go-linux-amd64
```

### Step 3: Add to Your AI Tool Config

See the [Setup by Tool](#setup-by-tool) section below for your specific client.

---

## Setup by Tool

Replace `/path/to/binary` with the actual path to your downloaded binary. Replace `your-api-key-here` with your Google PSI API key.

### Claude Code / GitHub Copilot CLI

Edit `~/.claude/claude_desktop_config.json` (or your Copilot CLI MCP config):

```json
{
  "mcpServers": {
    "pagespeed-insights": {
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

Or using CLI argument:

```json
{
  "mcpServers": {
    "pagespeed-insights": {
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "args": ["--api-key", "your-api-key-here"]
    }
  }
}
```

### Claude Desktop

Edit `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "pagespeed-insights": {
      "command": "/path/to/psi-mcp-go-darwin-arm64",
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### Cursor

Open Cursor Settings → MCP → Add MCP Server, or edit `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "pagespeed-insights": {
      "command": "/path/to/psi-mcp-go-linux-amd64",
      "env": {
        "GOOGLE_PSI_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### VS Code with GitHub Copilot

Edit your `.vscode/mcp.json` workspace file or user-level `settings.json`:

```json
{
  "mcp": {
    "servers": {
      "pagespeed-insights": {
        "type": "stdio",
        "command": "/path/to/psi-mcp-go-linux-amd64",
        "env": {
          "GOOGLE_PSI_API_KEY": "your-api-key-here"
        }
      }
    }
  }
}
```

### Visual Studio

In Visual Studio 2022 17.14+, open **Tools → Options → GitHub Copilot → MCP Servers** and add:

```json
{
  "pagespeed-insights": {
    "command": "C:\\path\\to\\psi-mcp-csharp-win-x64.exe",
    "env": {
      "GOOGLE_PSI_API_KEY": "your-api-key-here"
    }
  }
}
```

### Any MCP-Compatible Client

The server communicates over STDIO using standard JSON-RPC 2.0. Any MCP client that supports STDIO transport and the `tools/list` + `tools/call` protocol will work:

```json
{
  "command": "/path/to/binary",
  "args": ["--api-key", "your-key"],
  "transport": "stdio"
}
```

---

## Available Tools

### `analyze_page`

Analyze a single URL with Google PageSpeed Insights.

**Parameters:**

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `url` | string | Yes | -- | The URL to analyze |
| `strategy` | string | No | `"both"` | Analysis strategy: `"mobile"`, `"desktop"`, or `"both"` |

**Example prompt:**

> "Analyze https://www.devleader.ca on both mobile and desktop and tell me what's hurting the Core Web Vitals."

### `analyze_pages`

Analyze multiple URLs in a single call. Returns an array of results.

**Parameters:**

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `urls` | string[] | Yes | -- | Array of URLs to analyze |
| `strategy` | string | No | `"both"` | Analysis strategy: `"mobile"`, `"desktop"`, or `"both"` |

**Example prompt:**

> "Analyze my homepage, about page, and a recent blog post on mobile and compare their Core Web Vitals scores."

### Response Structure

Both tools return a JSON array. Each element represents one URL + strategy combination:

```json
{
  "url": "https://example.com",
  "strategy": "mobile",
  "analyzedAt": "2026-02-21T19:00:00Z",
  "scores": {
    "performance": 72,
    "seo": 92,
    "accessibility": 98,
    "bestPractices": 95
  },
  "coreWebVitals": {
    "fcp":  { "value": 1.8, "unit": "s",  "rating": "needs-improvement" },
    "lcp":  { "value": 3.2, "unit": "s",  "rating": "needs-improvement" },
    "cls":  { "value": 0.08,              "rating": "good" },
    "tbt":  { "value": 85,  "unit": "ms", "rating": "good" },
    "ttfb": { "value": 420, "unit": "ms", "rating": "needs-improvement" },
    "speedIndex": { "value": 2.1, "unit": "s", "rating": "good" }
  },
  "opportunities": [
    {
      "id": "render-blocking-resources",
      "title": "Eliminate render-blocking resources",
      "description": "Resources are blocking the first paint of your page.",
      "savings": "Potential savings of 500 ms",
      "impact": "high"
    }
  ],
  "failingAudits": [
    {
      "id": "uses-text-compression",
      "title": "Enable text compression",
      "description": "Text-based resources should be served with compression.",
      "score": 0,
      "displayValue": "Potential savings of 230 KiB"
    }
  ],
  "passedAuditIds": ["uses-https", "viewport", "document-title"]
}
```

**Core Web Vitals rating thresholds (Google standard):**

| Metric | Good | Needs Improvement | Poor |
|--------|------|-------------------|------|
| LCP | < 2.5s | < 4s | >= 4s |
| CLS | < 0.1 | < 0.25 | >= 0.25 |
| FCP | < 1.8s | < 3s | >= 3s |
| TTFB | < 0.8s | < 1.8s | >= 1.8s |
| TBT | < 200ms | < 600ms | >= 600ms |
| Speed Index | < 3.4s | < 5.8s | >= 5.8s |

---

## Configuration Reference

API key resolution uses this priority order (highest to lowest):

### 1. CLI Argument (Highest Priority)

```bash
/path/to/psi-mcp-go-linux-amd64 --api-key YOUR_KEY
```

In MCP config JSON:

```json
{ "args": ["--api-key", "your-key"] }
```

### 2. Environment Variable

```bash
export GOOGLE_PSI_API_KEY=your-key
```

In MCP config JSON:

```json
{ "env": { "GOOGLE_PSI_API_KEY": "your-key" } }
```

### 3. `.env` File (Lowest Priority -- Dev Convenience)

Create a `.env` file in the working directory:

```
GOOGLE_PSI_API_KEY=your-key
```

This is useful when running the binary directly for development and testing. It is NOT intended for production MCP configurations -- use environment variables or the CLI argument instead.

---

## Go vs C# -- Which Binary?

Both implementations expose identical tools with identical behavior. The only differences are implementation details:

| Aspect | Go | C# Native AOT |
|--------|----|----|
| Binary size | ~8-15 MB | ~20-40 MB |
| Startup time | ~10-50ms | ~50-100ms |
| Runtime dependency | None | None |
| Language | Go 1.26 | C# / .NET 10 |
| MCP SDK | Official `go-sdk` | Official `ModelContextProtocol` |
| Cross-platform | Yes | Yes |

**Recommendation:** Both work great. If you're already working in a .NET ecosystem, the C# binary may feel more natural to contribute to. Otherwise, the Go binary is slightly smaller and faster to start.

---

## Building from Source

### Go

Requires Go 1.26+:

```bash
cd go
go mod tidy
go build -ldflags="-s -w" -trimpath -o psi-mcp-go .

# Cross-compile example (Linux amd64 from any platform):
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o psi-mcp-go-linux-amd64 .
```

Run tests:

```bash
go test ./...
```

### C# (.NET 10 SDK Required)

Requires the [.NET 10 SDK](https://dotnet.microsoft.com/download/dotnet/10.0):

```bash
cd csharp

# Build (non-AOT, for development)
dotnet build PageSpeedMcp.slnx

# Publish Native AOT (no runtime needed, matches release binaries)
dotnet publish src/PageSpeedMcp/PageSpeedMcp.csproj -r linux-x64 -c Release --self-contained true

# Run tests
dotnet test PageSpeedMcp.slnx
```

> **Note:** Native AOT compilation requires platform-specific native tooling (clang on Linux/macOS, MSVC on Windows). The standard `dotnet build` works for development without these tools.

---

## Related Projects

- [google-search-console-mcp](https://github.com/ncosentino/google-search-console-mcp) -- Google Search Console MCP server: query clicks, impressions, CTR, and ranking position from your Search Console properties
- [google-keyword-planner-mcp](https://github.com/ncosentino/google-keyword-planner-mcp) -- Google Ads Keyword Planner MCP server: keyword ideas, search volume, competition, and CPC data

---

## About

### Nick Cosentino -- Dev Leader

This MCP server was built by **[Nick Cosentino](https://www.devleader.ca)**, a software engineer and content creator known as **Dev Leader**. Nick creates practical .NET, C#, ASP.NET Core, Blazor, and software engineering content for intermediate to advanced developers -- covering everything from performance optimization and clean architecture to real-world career advice.

This tool was born out of real work improving the Core Web Vitals of [devleader.ca](https://www.devleader.ca) and the desire to use AI assistants effectively during that process. It serves as a practical example of building Native AOT C# and idiomatic Go MCP servers with zero runtime dependencies.

**Find Nick online:**

- Blog: [https://www.devleader.ca](https://www.devleader.ca)
- YouTube: [https://www.youtube.com/@devleaderca](https://www.youtube.com/@devleaderca)
- Newsletter: [https://weekly.devleader.ca](https://weekly.devleader.ca)
- LinkedIn: [https://linkedin.com/in/nickcosentino](https://linkedin.com/in/nickcosentino)
- Linktree: [https://www.linktr.ee/devleader](https://www.linktr.ee/devleader)

### BrandGhost

[BrandGhost](https://www.brandghost.ai) is a social media automation platform built by Nick that lets content creators cross-post and schedule content across all social platforms in one click. If you create content and want to spend less time on distribution and more time creating, check it out.

---

## Contributing

Contributions are welcome! Please:

1. Open an issue describing the bug or feature request before submitting a PR
2. Run `golangci-lint run` (Go) or `dotnet build` with zero warnings (C#) before submitting
3. Keep both implementations in sync -- a feature added to Go should also be added to C#, and vice versa

---

## License

MIT License -- see [LICENSE](LICENSE) for details.
