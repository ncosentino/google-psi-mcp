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
    [Description("Analyze a single URL using Google PageSpeed Insights. Separates real-user CrUX field data from synthetic Lighthouse lab data and returns Lighthouse 13 insights with structured details. Agentic browsing is experimental and must be requested explicitly.")]
    internal async Task<string> AnalyzePage(
        [Description("The URL to analyze.")] string url,
        [Description("Analysis strategy: mobile, desktop, or both. Defaults to both.")] string strategy = "both",
        [Description("Lighthouse categories to run. Defaults to performance, seo, accessibility, and best-practices. agentic-browsing is experimental and opt-in.")] string[]? categories = null,
        [Description("Optional locale for Lighthouse display strings, such as en-US or fr.")] string? locale = null,
        CancellationToken cancellationToken = default)
    {
        var response = await RunAnalysisAsync(
            [url],
            strategy,
            categories,
            locale,
            cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(response, PsiJsonContext.Default.AnalysisResponse);
    }

    [McpServerTool(Name = "analyze_pages")]
    [Description("Analyze multiple URLs using Google PageSpeed Insights. Returns separate real-user field data and Lighthouse lab data for every URL and strategy. Agentic browsing is experimental and must be requested explicitly.")]
    internal async Task<string> AnalyzePages(
        [Description("Array of URLs to analyze.")] string[] urls,
        [Description("Analysis strategy: mobile, desktop, or both. Defaults to both.")] string strategy = "both",
        [Description("Lighthouse categories to run. Defaults to performance, seo, accessibility, and best-practices. agentic-browsing is experimental and opt-in.")] string[]? categories = null,
        [Description("Optional locale for Lighthouse display strings, such as en-US or fr.")] string? locale = null,
        CancellationToken cancellationToken = default)
    {
        var response = await RunAnalysisAsync(
            urls,
            strategy,
            categories,
            locale,
            cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(response, PsiJsonContext.Default.AnalysisResponse);
    }

    private async Task<AnalysisResponse> RunAnalysisAsync(
        string[] urls,
        string strategy,
        string[]? categories,
        string? locale,
        CancellationToken cancellationToken)
    {
        var strategies = PageSpeedAnalysisRequest.ResolveStrategies(strategy);
        var requests = urls
            .SelectMany(url => strategies.Select(
                selectedStrategy => PageSpeedAnalysisRequest.Create(
                    url,
                    selectedStrategy,
                    categories,
                    locale)))
            .ToArray();

        var tasks = requests
            .Select(request => RunSingleAnalysisAsync(request, cancellationToken))
            .ToArray();

        var entries = await Task.WhenAll(tasks).ConfigureAwait(false);
        return new AnalysisResponse(
            entries.Where(entry => entry.Result is not null).Select(entry => entry.Result!).ToArray(),
            entries.Where(entry => entry.Error is not null).Select(entry => entry.Error!).ToArray());
    }

    private async Task<AnalysisEntry> RunSingleAnalysisAsync(
        PageSpeedAnalysisRequest request,
        CancellationToken cancellationToken)
    {
        try
        {
            return new AnalysisEntry(
                await client.AnalyzeAsync(request, cancellationToken).ConfigureAwait(false),
                null);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return new AnalysisEntry(
                null,
                new AnalysisFailure(
                    request.Url,
                    request.Strategy,
                    "The PageSpeed Insights request timed out."));
        }
        catch (HttpRequestException ex)
        {
            return new AnalysisEntry(
                null,
                new AnalysisFailure(request.Url, request.Strategy, ex.Message));
        }
        catch (JsonException ex)
        {
            return new AnalysisEntry(
                null,
                new AnalysisFailure(
                    request.Url,
                    request.Strategy,
                    $"The PageSpeed Insights response was invalid: {ex.Message}"));
        }
    }

    private sealed record AnalysisEntry(PageSpeedAnalysis? Result, AnalysisFailure? Error);
}