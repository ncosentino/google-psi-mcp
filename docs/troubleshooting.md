# Troubleshooting

## Timeouts

The most common problem with this MCP server is timeouts. The PageSpeed Insights API is a real Lighthouse analysis -- it actually loads your page in a headless browser and runs performance audits. That takes time.

### Why timeouts happen

- The PSI API takes **5-20+ seconds per URL per strategy**, depending on the page's complexity and Google's infrastructure load
- `strategy="both"` makes **two sequential API calls** (mobile then desktop) -- that's 10-40+ seconds for a single URL
- `analyze_pages` multiplies this: N URLs × S strategies = N×S sequential calls
- Most MCP clients (Claude Desktop, GitHub Copilot CLI) have a **30-60 second timeout** for tool calls

### How to avoid timeouts

**Use a single strategy for interactive analysis:**

Instead of `strategy="both"`, use `strategy="mobile"` (the default) or `strategy="desktop"`. You get results faster, and you can always run the other strategy separately if needed.

**Reduce batch size:**

With `analyze_pages`, stick to 2-3 URLs per call. For larger audits, call `analyze_page` on each URL individually so a single timeout doesn't lose all results.

**If your client supports timeout configuration:**

Set the MCP tool timeout to at least **90 seconds** to give PSI room to respond even on slow pages.

### Which calls are most at risk

| Scenario | Estimated call time | Risk level |
|----------|-------------------|------------|
| `analyze_page`, `strategy="mobile"` | 5-20s | Low |
| `analyze_page`, `strategy="desktop"` | 5-20s | Low |
| `analyze_page`, `strategy="both"` | 10-40s | Medium-High |
| `analyze_pages` (2 URLs, mobile) | 10-40s | Medium |
| `analyze_pages` (3+ URLs, `both`) | 30-120s | **Very High** |

### The request didn't time out but returned no data

If the API key is invalid or not authorized for the PageSpeed Insights API, the server returns an error immediately. Confirm:

1. The `GOOGLE_PSI_API_KEY` environment variable is set and non-empty
2. The PageSpeed Insights API is enabled in your Google Cloud project (not just created)
3. The key is not restricted to a different API

---

## API Key Issues

**Error at startup: "API key is required"**

The server couldn't find a key in any of the three sources (CLI argument, environment variable, `.env` file). Verify the variable name is exactly `GOOGLE_PSI_API_KEY` -- it's case-sensitive.

**"API key not valid" error from Google**

The key exists but the PageSpeed Insights API isn't enabled for the project that owns it. Go to Google Cloud Console → APIs & Services → Library and enable the PageSpeed Insights API.

