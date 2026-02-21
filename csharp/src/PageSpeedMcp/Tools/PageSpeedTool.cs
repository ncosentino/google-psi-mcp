using System.ComponentModel;
using System.Text.Json;
using ModelContextProtocol.Server;
using PageSpeedMcp.PageSpeed;

namespace PageSpeedMcp.Tools;

/// <summary>MCP tools for Google PageSpeed Insights analysis.</summary>
[McpServerToolType]
internal sealed class PageSpeedTool(PageSpeedClient client)
{
    [McpServerTool(Name = "analyze_page")]
    [Description("Analyze a single URL using Google PageSpeed Insights. Returns Core Web Vitals (FCP, LCP, CLS, TBT, TTFB), category scores (performance, SEO, accessibility, best-practices), and actionable audit findings.")]
    internal async Task<string> AnalyzePage(
        [Description("The URL to analyze.")] string url,
        [Description("Analysis strategy: mobile, desktop, or both. Defaults to both.")] string strategy = "both",
        CancellationToken cancellationToken = default)
    {
        var results = await RunAnalysisAsync([url], strategy, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(results, PsiJsonContext.Default.PageSpeedResultArray);
    }

    [McpServerTool(Name = "analyze_pages")]
    [Description("Analyze multiple URLs using Google PageSpeed Insights in a single call. Returns an array of results, one per URL.")]
    internal async Task<string> AnalyzePages(
        [Description("Array of URLs to analyze.")] string[] urls,
        [Description("Analysis strategy: mobile, desktop, or both. Defaults to both.")] string strategy = "both",
        CancellationToken cancellationToken = default)
    {
        var results = await RunAnalysisAsync(urls, strategy, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(results, PsiJsonContext.Default.PageSpeedResultArray);
    }

    private async Task<PageSpeedResult[]> RunAnalysisAsync(
        string[] urls,
        string strategy,
        CancellationToken cancellationToken)
    {
        string[] strategies = strategy.Equals("both", StringComparison.OrdinalIgnoreCase)
            ? ["mobile", "desktop"]
            : [strategy.ToLowerInvariant()];

        var tasks = urls
            .SelectMany(u => strategies.Select(s => client.AnalyzeAsync(u, s, cancellationToken)))
            .ToArray();

        return await Task.WhenAll(tasks).ConfigureAwait(false);
    }
}