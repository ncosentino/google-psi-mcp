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
- `rawResponse`: the complete successful Google API JSON response, including
  unknown fields and details of audits omitted from the normalized summary.

Field metrics use the upstream p75 rating and preserve histogram distributions.
Lab metrics retain their Lighthouse score and unit instead of receiving
field-data ratings.

Use `rawResponse.lighthouseResult.audits["lcp-breakdown-insight"].details`
for Google's LCP phase breakdown when supplied, even if Lighthouse marked the
audit as passed. Other returned audits, configuration, environment, network
details, timings, and screenshot data remain accessible in the same raw report.
The server does not remove fields based on scores or a fixed allow-list.

`rawResponse` is included by default and requires no additional API request.
It preserves JSON values and types, not byte-for-byte formatting. Reports may
be large or contain sensitive page data; the MCP client may also have its own
output limits. Fields or trace artifacts Google does not return are not created.

## Examples

```text
Analyze https://www.devleader.ca on mobile. Compare real-user LCP, CLS, and INP
with the Lighthouse lab metrics and prioritize the current insights.
```

```text
Run the experimental agentic-browsing category for https://www.devleader.ca
and summarize its WebMCP and llms.txt findings.
```
