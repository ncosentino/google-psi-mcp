---
description: Diagnose PSI timeouts, CrUX permissions, HTTP host filtering, and MCP argument errors.
---

# Troubleshooting

## PSI analysis is slow

PSI loads the page in Chrome and runs Lighthouse. A request can take tens of
seconds. The server allows 120 seconds per PSI request, retries transient
failures, and runs at most four batch requests concurrently.

For interactive use, prefer a single strategy when both mobile and desktop are
not required. Successful batch results are preserved even when another URL
fails.

## CrUX returns PERMISSION_DENIED

The direct CrUX tools use `chromeuxreport.googleapis.com`, not the PageSpeed
Insights API.

Confirm:

1. The Chrome UX Report API is enabled in the key's project.
2. The key's API restrictions allow the Chrome UX Report API.
3. The target has enough eligible Chrome traffic for CrUX data.

## A metric is absent

CrUX only returns metrics that meet its privacy and eligibility thresholds.
Absent data is not converted to zero. Historical `"NaN"` and `null` values are
returned as JSON `null`.

## HTTP requests return 403

For Go, add the request host to `--allowed-hosts`. For C#, configure ASP.NET
Core `AllowedHosts`.

Cross-site browser requests are intentionally rejected. HTTP transport should
normally be accessed by an MCP client or reverse proxy, not arbitrary browser
JavaScript.

## Array parameter validation fails

The server repairs known MCP clients that encode array parameters as JSON
strings. If validation still fails, confirm that the decoded value is actually
an array and that values such as category, strategy, form factor, and metric
names are valid.

## API key errors

The key is resolved from `--api-key`, `GOOGLE_PSI_API_KEY`, or `.env`, in that
order. The server exits at startup when no key is available.
