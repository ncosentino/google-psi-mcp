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
        var response = await RunAnalysisAsync([url], strategy, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(response, PsiJsonContext.Default.AnalysisResponse);
    }

    [McpServerTool(Name = "analyze_pages")]
    [Description("Analyze multiple URLs using Google PageSpeed Insights in a single call. Returns an array of results, one per URL.")]
    internal async Task<string> AnalyzePages(
        [Description("Array of URLs to analyze.")] string[] urls,
        [Description("Analysis strategy: mobile, desktop, or both. Defaults to both.")] string strategy = "both",
        CancellationToken cancellationToken = default)
    {
        var response = await RunAnalysisAsync(urls, strategy, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(response, PsiJsonContext.Default.AnalysisResponse);
    }

    private async Task<AnalysisResponse> RunAnalysisAsync(
        string[] urls,
        string strategy,
        CancellationToken cancellationToken)
    {
        string[] strategies = strategy.Equals("both", StringComparison.OrdinalIgnoreCase)
            ? ["mobile", "desktop"]
            : [strategy.ToLowerInvariant()];

        var tasks = urls
            .SelectMany(url => strategies.Select(
                selectedStrategy => RunSingleAnalysisAsync(url, selectedStrategy, cancellationToken)))
            .ToArray();

        var entries = await Task.WhenAll(tasks).ConfigureAwait(false);
        return new AnalysisResponse(
            entries.Where(entry => entry.Result is not null).Select(entry => entry.Result!).ToArray(),
            entries.Where(entry => entry.Error is not null).Select(entry => entry.Error!).ToArray());
    }

    private async Task<AnalysisEntry> RunSingleAnalysisAsync(
        string url,
        string strategy,
        CancellationToken cancellationToken)
    {
        try
        {
            return new AnalysisEntry(
                await client.AnalyzeAsync(url, strategy, cancellationToken).ConfigureAwait(false),
                null);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return new AnalysisEntry(
                null,
                new AnalysisFailure(url, strategy, "The PageSpeed Insights request timed out."));
        }
        catch (HttpRequestException ex)
        {
            return new AnalysisEntry(null, new AnalysisFailure(url, strategy, ex.Message));
        }
        catch (JsonException ex)
        {
            return new AnalysisEntry(
                null,
                new AnalysisFailure(url, strategy, $"The PageSpeed Insights response was invalid: {ex.Message}"));
        }
    }

    private sealed record AnalysisEntry(PageSpeedAnalysis? Result, AnalysisFailure? Error);
}