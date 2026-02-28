# MCP Tools

The server exposes two tools to your AI assistant.

| Tool | Description |
|------|-------------|
| [`analyze_page`](analyze-page/) | Analyze a single URL with PageSpeed Insights |
| [`analyze_pages`](analyze-pages/) | Analyze multiple URLs in a single call |

---

## Strategy Parameter

Both tools accept an optional `strategy` parameter:

| Value | What it does |
|-------|-------------|
| `"mobile"` | Runs a mobile Lighthouse analysis (default) |
| `"desktop"` | Runs a desktop Lighthouse analysis |
| `"both"` | Runs mobile **then** desktop -- two sequential API calls |

!!! warning "strategy=\"both\" doubles latency"
    The PSI API takes 5-20+ seconds per URL. Using `strategy="both"` makes two back-to-back calls. On a slow or complex page, this can exceed your MCP client's timeout. Prefer `"mobile"` or `"desktop"` for interactive use unless you specifically need both strategies.

---

## Core Web Vitals Thresholds

Google's official thresholds for each metric:

| Metric | Good | Needs Improvement | Poor |
|--------|------|-------------------|------|
| LCP (Largest Contentful Paint) | < 2.5s | 2.5s - 4.0s | > 4.0s |
| CLS (Cumulative Layout Shift) | < 0.1 | 0.1 - 0.25 | > 0.25 |
| FCP (First Contentful Paint) | < 1.8s | 1.8s - 3.0s | > 3.0s |
| TTFB (Time to First Byte) | < 0.8s | 0.8s - 1.8s | > 1.8s |
| TBT (Total Blocking Time) | < 200ms | 200ms - 600ms | > 600ms |
| Speed Index | < 3.4s | 3.4s - 5.8s | > 5.8s |

The `analyze_page` and `analyze_pages` tools return the numeric value **and** the `rating` field (`"good"`, `"needs-improvement"`, or `"poor"`) for each metric.

