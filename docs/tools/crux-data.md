---
description: Query current Chrome UX Report data for a URL or origin.
---

# get_crux_data

Query the current 28-day Chrome UX Report aggregation.

## Parameters

| Parameter | Type | Required | Default |
|---|---|---|---|
| `target` | string | Yes | - |
| `target_type` | string | No | `url` |
| `form_factor` | string | No | `all` |
| `metrics` | string[] | No | all available |

`target_type` is `url` or `origin`. Origin targets cannot contain a path, query,
or fragment. `form_factor` is `all`, `phone`, `tablet`, or `desktop`.

Omitting `metrics` requests every metric available for that target. Current
CrUX metrics include Core Web Vitals, FCP, TTFB, RTT, navigation types, form
factors, LCP resource type, and LCP image subparts.

The response preserves histograms, p75 values, fractions, collection dates, and
URL normalization.
It also includes `rawResponse`, the complete successful Google API JSON response
without filtering or truncation. Unknown fields, numeric strings, and nulls are
preserved alongside the existing normalized metrics, without another API call.

!!! warning "API enablement"
    Enable the Chrome UX Report API and allow it in the API key restrictions.
    Enabling only PageSpeed Insights is insufficient.
