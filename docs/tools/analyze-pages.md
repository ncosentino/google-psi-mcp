---
description: MCP tool to analyze multiple URLs with Google PageSpeed Insights in one call. Includes timeout risk table, batch size recommendations, and example prompts.
---

# analyze_pages

Analyze multiple URLs in a single tool call.

---

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `urls` | string[] | Yes | Array of full URLs to analyze |
| `strategy` | string | No | `"mobile"` (default), `"desktop"`, or `"both"` |

---

## Timeout Warning

!!! danger "Batch calls can be slow -- read before using"
    The PSI API takes **5-20+ seconds per URL per strategy**. `analyze_pages` makes these calls sequentially:

    - 3 URLs × `strategy="mobile"` = 3 calls = **15-60+ seconds**
    - 3 URLs × `strategy="both"` = 6 calls = **30-120+ seconds**

    Most MCP clients (Claude Desktop, GitHub Copilot CLI) have a **30-60 second timeout**. Batches that exceed this will fail mid-way and return no results.

    **Recommendations for large batches:**

    - Use `strategy="mobile"` or `strategy="desktop"` -- never `"both"` for multi-URL calls
    - Keep batches to 2-3 URLs at a time for interactive use
    - For larger audits, analyze URLs individually with `analyze_page`

---

## Response

Returns an array, one entry per URL. Each entry has the same structure as `analyze_page`:

- Category scores (performance, SEO, accessibility, best-practices)
- Core Web Vitals with ratings (LCP, CLS, FCP, TTFB, TBT, Speed Index)
- Opportunities with estimated savings
- Failing audits
- Passed audit IDs

See [analyze_page](analyze-page/) for the full response schema.

---

## Example Prompts

```
Analyze these three pages on mobile: https://www.example.com, https://www.example.com/about, https://www.example.com/blog
Summarize which page has the worst Core Web Vitals.
```

```
Check https://www.mysite.com and https://www.mysite.com/pricing on desktop.
Which has a better performance score?
```

