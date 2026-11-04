---
description: Query up to 40 weekly Chrome UX Report collection periods.
---

# get_crux_history

Query weekly CrUX timeseries for a URL or origin.

The tool accepts the same target, form-factor, and metric controls as
[`get_crux_data`](crux-data.md), plus:

| Parameter | Type | Required | Default |
|---|---|---|---|
| `collection_period_count` | integer | No | `25` |

The allowed range is 1 through 40. Each data point is a 28-day rolling
aggregation, and adjacent weekly periods overlap.

CrUX may represent an unavailable historical value as `"NaN"` or `null`.
The server normalizes both to JSON `null` in its metric summaries.
The included `rawResponse` preserves the complete successful Google API JSON
response, including the original `"NaN"` strings, nulls, numeric strings, and
unknown fields. It is not filtered or truncated and requires no additional call.

```text
Show the last 40 phone collection periods for the devleader.ca origin. Explain
whether LCP, CLS, and INP are improving.
```
