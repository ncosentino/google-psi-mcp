---
description: MCP tool to analyze a single URL with Google PageSpeed Insights. Parameters, full response schema (Core Web Vitals, category scores, opportunities, failing audits), and example prompts.
---

# analyze_page

Analyze a single URL with the Google PageSpeed Insights API.

---

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | Yes | The full URL to analyze (include `https://`) |
| `strategy` | string | No | `"mobile"` (default), `"desktop"`, or `"both"` |

!!! warning "strategy=\"both\" -- latency warning"
    Using `strategy="both"` makes **two sequential API calls** -- one mobile, one desktop. The PSI API can take 5-20+ seconds per call. On a slow or complex page, `strategy="both"` can easily take 30+ seconds, approaching the timeout limit of most MCP clients. Use `"mobile"` or `"desktop"` for interactive use. Reserve `"both"` for pages you know respond quickly.

---

## Response

The tool returns a structured result with the following sections for each strategy analyzed:

### Category Scores

Overall Lighthouse category scores (0-100):

- `performance` -- overall performance score
- `seo` -- SEO audit score
- `accessibility` -- accessibility score
- `bestPractices` -- best practices score

### Core Web Vitals

Each metric includes a numeric value and a `rating`:

| Field | Description |
|-------|-------------|
| `lcp` | Largest Contentful Paint (seconds) |
| `cls` | Cumulative Layout Shift (unitless score) |
| `fcp` | First Contentful Paint (seconds) |
| `ttfb` | Time to First Byte (seconds) |
| `tbt` | Total Blocking Time (milliseconds) |
| `speedIndex` | Speed Index (seconds) |

Each has a `rating` field: `"good"`, `"needs-improvement"`, or `"poor"`.

### Opportunities

A list of improvements the page could make, each with:

- `id` -- the Lighthouse audit ID
- `title` -- human-readable description
- `description` -- detailed explanation
- `estimatedSavings` -- estimated time savings if addressed (seconds)

### Failing Audits

Audits the page did not pass:

- `id`, `title`, `description` -- same structure as opportunities
- `displayValue` -- the measured value (e.g., "3 elements")

### Passed Audits

List of audit IDs that passed, for reference.

---

## Example Prompts

```
Analyze https://www.devleader.ca on mobile and tell me what's hurting my LCP score.
```

```
Check https://www.example.com on desktop. What failing audits should I fix first?
```

```
Analyze https://www.mysite.com on mobile. Summarize the opportunities with the highest estimated savings.
```

