---
description: Reference for PageSpeed Insights analysis, current CrUX data, and CrUX history MCP tools.
---

# MCP Tools

| Tool | Upstream API | Purpose |
|---|---|---|
| [`analyze_page`](analyze-page.md) | PageSpeed Insights v5 | Analyze one URL |
| [`analyze_pages`](analyze-pages.md) | PageSpeed Insights v5 | Analyze up to 10 URLs |
| [`get_crux_data`](crux-data.md) | Chrome UX Report API | Current real-user measurements |
| [`get_crux_history`](crux-history.md) | CrUX History API | Weekly real-user timeseries |

## PSI versus CrUX

PSI combines a Lighthouse lab run with a reduced CrUX field-data section. The
direct CrUX tools provide the broader metric catalog, form-factor controls,
collection periods, and historical trends.

## PSI strategies

| Value | Result |
|---|---|
| `mobile` | One mobile Lighthouse run |
| `desktop` | One desktop Lighthouse run |
| `both` | Mobile and desktop runs |

`both` is the default. Batch analysis uses at most four concurrent requests and
returns successful results alongside structured per-request errors.

## PSI categories

The default categories are performance, SEO, accessibility, and best practices.
Use `categories` to request a subset or to opt into experimental
`agentic-browsing`.
