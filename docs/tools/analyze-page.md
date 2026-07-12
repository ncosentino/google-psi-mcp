---
description: Analyze one URL with PSI field data, Lighthouse lab data, and Lighthouse 13 insights.
---

# analyze_page

Analyze one absolute HTTP or HTTPS URL.

## Parameters

| Parameter | Type | Required | Default |
|---|---|---|---|
| `url` | string | Yes | - |
| `strategy` | string | No | `both` |
| `categories` | string[] | No | performance, SEO, accessibility, best practices |
| `locale` | string | No | PSI default |

Valid strategies are `mobile`, `desktop`, and `both`.

Valid categories are:

- `performance`
- `seo`
- `accessibility`
- `best-practices`
- `agentic-browsing` (experimental)

## Response

The response contains `results` and `errors`. Every successful result has:

- `metadata`: input strategy, PSI timestamp, Lighthouse version, redirects,
  warnings, and runtime errors.
- `fieldData`: page and origin CrUX measurements. Missing data remains absent;
  page-to-origin fallback is explicit.
- `labData`: open category map, lab metrics, Lighthouse 13 insights,
  diagnostics, audit details, metric savings, and entity classifications.

Field metrics use the upstream p75 rating and preserve histogram distributions.
Lab metrics retain their Lighthouse score and unit instead of receiving
field-data ratings.

## Examples

```text
Analyze https://www.devleader.ca on mobile. Compare real-user LCP, CLS, and INP
with the Lighthouse lab metrics and prioritize the current insights.
```

```text
Run the experimental agentic-browsing category for https://www.devleader.ca
and summarize its WebMCP and llms.txt findings.
```
