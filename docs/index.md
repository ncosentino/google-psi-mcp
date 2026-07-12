---
description: Native MCP server for PageSpeed Insights, Lighthouse 13 insights, current CrUX data, and CrUX history.
---

# Google PageSpeed Insights and CrUX MCP

Use Google PageSpeed Insights and Chrome UX Report data directly from an MCP
client without installing Node.js, Python, Go, or the .NET runtime.

The server distinguishes:

- **Field data**: aggregated real-user CrUX measurements, including LCP, CLS,
  and INP.
- **Lab data**: one synthetic Lighthouse run, including scores, metrics,
  insights, diagnostics, redirects, warnings, and Lighthouse version.

## Tools

| Tool | Description |
|---|---|
| [`analyze_page`](tools/analyze-page.md) | Analyze one URL with PSI |
| [`analyze_pages`](tools/analyze-pages.md) | Analyze up to 10 URLs |
| [`get_crux_data`](tools/crux-data.md) | Query current CrUX data |
| [`get_crux_history`](tools/crux-history.md) | Query weekly CrUX history |

Both Go and C# implementations support STDIO and stateless Streamable HTTP.

[Get Started :material-arrow-right:](getting-started.md){ .md-button .md-button--primary }

!!! note "Agentic Browsing"
    Lighthouse 13's `agentic-browsing` category is available as an explicit,
    experimental category. It is not part of the default analysis.

!!! warning "Direct CrUX API access"
    The Chrome UX Report API must be enabled and allowed by the configured API
    key before the two direct CrUX tools can be used.
