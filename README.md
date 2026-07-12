# PageSpeed Insights and CrUX MCP Server

[![Latest Release](https://img.shields.io/github/v/release/ncosentino/google-psi-mcp?style=flat-square)](https://github.com/ncosentino/google-psi-mcp/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg?style=flat-square)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](go/go.mod)
[![.NET Version](https://img.shields.io/badge/.NET-10-512BD4?style=flat-square&logo=dotnet)](csharp/Directory.Build.props)
[![CI](https://img.shields.io/github/actions/workflow/status/ncosentino/google-psi-mcp/ci.yml?label=CI&style=flat-square)](https://github.com/ncosentino/google-psi-mcp/actions/workflows/ci.yml)

Native Go and C# MCP servers for Google PageSpeed Insights and the Chrome UX
Report. The server keeps real-user field data separate from synthetic Lighthouse
lab data and exposes current Lighthouse 13 insights without flattening their
structured details.

Pre-built binaries require no Go, .NET, Node.js, or Python runtime.

## Capabilities

- Real-user page and origin data from PSI's embedded CrUX response
- True Core Web Vitals: LCP, CLS, and INP
- Lighthouse lab metrics, category scores, execution metadata, and redirects
- Lighthouse 13 insight details and metric savings
- Optional experimental Agentic Browsing, WebMCP, and `llms.txt` audits
- Direct current CrUX data, including LCP subparts, RTT, and navigation types
- Up to 40 weekly CrUX History API collection periods
- STDIO and stateless Streamable HTTP MCP transports
- Matching Go and C# Native AOT implementations

## Google API Setup

1. Create or select a project in [Google Cloud Console](https://console.cloud.google.com/).
2. Enable the [PageSpeed Insights API](https://console.cloud.google.com/apis/library/pagespeedonline.googleapis.com).
3. Create an API key.
4. To use `get_crux_data` or `get_crux_history`, also enable the
   [Chrome UX Report API](https://console.cloud.google.com/apis/library/chromeuxreport.googleapis.com).
5. If the key has API restrictions, allow every API you intend to use.

The direct CrUX tools return Google's permission error when the Chrome UX Report
API is disabled or excluded by the key's restrictions. PSI analysis continues to
work independently.

## Installation

Download a binary from [GitHub Releases](https://github.com/ncosentino/google-psi-mcp/releases/latest):

| Platform | Go | C# Native AOT |
|---|---|---|
| Linux x64 | `psi-mcp-go-linux-amd64` | `psi-mcp-csharp-linux-x64` |
| Linux ARM64 | `psi-mcp-go-linux-arm64` | `psi-mcp-csharp-linux-arm64` |
| macOS Intel | `psi-mcp-go-darwin-amd64` | `psi-mcp-csharp-osx-x64` |
| macOS Apple Silicon | `psi-mcp-go-darwin-arm64` | `psi-mcp-csharp-osx-arm64` |
| Windows x64 | `psi-mcp-go-windows-amd64.exe` | `psi-mcp-csharp-win-x64.exe` |
| Windows ARM64 | `psi-mcp-go-windows-arm64.exe` | `psi-mcp-csharp-win-arm64.exe` |

On Linux or macOS:

```bash
chmod +x psi-mcp-go-linux-amd64
```

## STDIO Configuration

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

The key can instead be passed with `--api-key` or loaded from a
`GOOGLE_PSI_API_KEY` entry in a `.env` file.

## HTTP Transport

```bash
PORT=8080 ./psi-mcp-go-linux-amd64 \
  --transport http \
  --allowed-hosts localhost,127.0.0.1
```

The C# binary accepts the same `--transport http` flag and `PORT` environment
variable. C# uses ASP.NET Core's `AllowedHosts` configuration; Go uses
`--allowed-hosts`.

HTTP transport is stateless. It includes host-header and cross-origin defenses,
but it does **not** add application authentication. Do not expose it publicly
without an authenticated reverse proxy or private network.

## MCP Tools

| Tool | Purpose |
|---|---|
| `analyze_page` | Analyze one URL with PSI |
| `analyze_pages` | Analyze up to 10 URLs with bounded concurrency |
| `get_crux_data` | Query current CrUX data for a URL or origin |
| `get_crux_history` | Query up to 40 CrUX history periods |

### `analyze_page`

| Parameter | Type | Required | Default |
|---|---|---|---|
| `url` | string | Yes | - |
| `strategy` | string | No | `both` |
| `categories` | string[] | No | performance, SEO, accessibility, best practices |
| `locale` | string | No | PSI default |

Valid strategies are `mobile`, `desktop`, and `both`. Valid categories are
`performance`, `seo`, `accessibility`, `best-practices`, and
`agentic-browsing`.

Agentic Browsing is experimental and is never included implicitly.

### `analyze_pages`

Accepts the same controls as `analyze_page`, with `urls` replacing `url`.
Batches are limited to 10 URLs and four concurrent PSI requests. Transient
network, HTTP 429, and HTTP 5xx failures are retried up to three attempts.

The result contains deterministic `results` and `errors` arrays so one failed
URL does not discard successful analyses.

### PSI response model

```json
{
  "results": [
    {
      "metadata": {
        "inputUrl": "https://example.com/",
        "strategy": "mobile",
        "lighthouseVersion": "13.4.0",
        "requestedUrl": "https://example.com/",
        "finalUrl": "https://www.example.com/"
      },
      "fieldData": {
        "page": {
          "overallRating": "needs-improvement",
          "originFallback": false,
          "metrics": {
            "lcp": {
              "id": "LARGEST_CONTENTFUL_PAINT_MS",
              "percentile": 75,
              "value": 3100,
              "unit": "ms",
              "rating": "needs-improvement",
              "distributions": []
            },
            "inp": {
              "id": "INTERACTION_TO_NEXT_PAINT",
              "percentile": 75,
              "value": 240,
              "unit": "ms",
              "rating": "needs-improvement",
              "distributions": []
            }
          }
        }
      },
      "labData": {
        "categories": {
          "performance": {
            "id": "performance",
            "title": "Performance",
            "score": 0.82,
            "scoreDisplayMode": "gauge"
          }
        },
        "metrics": {},
        "insights": [
          {
            "id": "render-blocking-insight",
            "title": "Render-blocking requests",
            "metricSavings": {
              "LCP": 710
            },
            "details": {
              "type": "table",
              "items": []
            }
          }
        ],
        "diagnostics": [],
        "unscoredAudits": [],
        "passedAuditIds": [],
        "notApplicableAuditIds": [],
        "manualAuditIds": [],
        "entities": []
      }
    }
  ],
  "errors": []
}
```

Field ratings come from Google's CrUX categories. Lab metrics retain Lighthouse
scores and units; the server does not mislabel synthetic metrics as real-user
Core Web Vitals.

### `get_crux_data`

| Parameter | Type | Required | Default |
|---|---|---|---|
| `target` | string | Yes | - |
| `target_type` | string | No | `url` |
| `form_factor` | string | No | `all` |
| `metrics` | string[] | No | all available |

`target_type` is `url` or `origin`. `form_factor` is `all`, `phone`, `tablet`,
or `desktop`.

### `get_crux_history`

Accepts the same inputs plus `collection_period_count`, from 1 to 40 and
defaulting to 25. Unavailable periods are returned as JSON `null`, never `NaN`.

## Building

```bash
cd go
go test ./...
golangci-lint run
go build -ldflags="-s -w -X main.version=dev" -trimpath -o psi-mcp-go .
```

```bash
cd csharp
dotnet test PageSpeedMcp.slnx -c Release
dotnet publish src/PageSpeedMcp/PageSpeedMcp.csproj \
  -r linux-x64 -c Release --self-contained true
```

## Related Projects

- [google-search-console-mcp](https://github.com/ncosentino/google-search-console-mcp)
- [google-keyword-planner-mcp](https://github.com/ncosentino/google-keyword-planner-mcp)

## About

Built by **[Nick Cosentino](https://www.devleader.ca)** (Dev Leader).

- [Blog](https://www.devleader.ca)
- [YouTube](https://www.youtube.com/@devleaderca)
- [Newsletter](https://weekly.devleader.ca)
- [LinkedIn](https://linkedin.com/in/nickcosentino)

## Contributing

1. Open an issue before submitting a feature.
2. Keep Go and C# behavior in parity.
3. Run Go lint/tests, C# build/tests, and the strict documentation build.

## License

MIT License - see [LICENSE](LICENSE).
