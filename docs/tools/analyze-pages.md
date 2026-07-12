---
description: Analyze up to 10 URLs with bounded PageSpeed Insights concurrency and partial results.
---

# analyze_pages

Analyze multiple absolute HTTP or HTTPS URLs.

## Parameters

| Parameter | Type | Required | Default |
|---|---|---|---|
| `urls` | string[] | Yes | - |
| `strategy` | string | No | `both` |
| `categories` | string[] | No | performance, SEO, accessibility, best practices |
| `locale` | string | No | PSI default |

The tool accepts between 1 and 10 URLs. It runs at most four PSI requests at
once and retries transient network, HTTP 429, and HTTP 5xx failures up to three
attempts.

Results remain in request order. One failed URL or strategy does not discard
successful analyses:

```json
{
  "results": [],
  "errors": [
    {
      "inputUrl": "https://example.com/",
      "strategy": "mobile",
      "code": "rate_limited",
      "message": "PSI API returned HTTP 429: ...",
      "retryable": true
    }
  ]
}
```

See [`analyze_page`](analyze-page.md) for the successful result structure.

## Example

```text
Analyze my homepage, about page, and latest article on mobile. Compare their
real-user Core Web Vitals and Lighthouse render-blocking insights.
```
