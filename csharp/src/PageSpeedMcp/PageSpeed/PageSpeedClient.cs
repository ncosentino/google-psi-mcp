using System.Net.Http.Json;
using System.Text;

namespace PageSpeedMcp.PageSpeed;

/// <summary>Client for the Google PageSpeed Insights API v5.</summary>
internal sealed class PageSpeedClient(HttpClient httpClient, string apiKey)
{
    private const string BaseUrl = "https://www.googleapis.com/pagespeedonline/v5/runPagespeed";

    /// <summary>Analyzes the specified URL using the given strategy.</summary>
    /// <param name="url">The URL to analyze.</param>
    /// <param name="strategy">"mobile" or "desktop".</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    public async Task<PageSpeedResult> AnalyzeAsync(
        string url,
        string strategy,
        CancellationToken cancellationToken = default)
    {
        try
        {
            var requestUrl = BuildRequestUrl(url, strategy);
            var response = await httpClient.GetAsync(requestUrl, cancellationToken).ConfigureAwait(false);

            if (!response.IsSuccessStatusCode)
            {
                var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
                var snippet = body.Length > 300 ? body[..300] + "..." : body;
                return new PageSpeedResult(url, strategy, DateTimeOffset.UtcNow,
                    null, null, null, null, null,
                    Error: $"PSI API returned HTTP {(int)response.StatusCode} {response.StatusCode}: {snippet}");
            }

            var raw = await response.Content
                .ReadFromJsonAsync(PsiJsonContext.Default.PsiApiResponse, cancellationToken)
                .ConfigureAwait(false);

            return PageSpeedResultParser.Parse(url, strategy, raw);
        }
        catch (OperationCanceledException) when (!cancellationToken.IsCancellationRequested)
        {
            return new PageSpeedResult(url, strategy, DateTimeOffset.UtcNow,
                null, null, null, null, null,
                Error: $"Request timed out after {httpClient.Timeout.TotalSeconds:0}s");
        }
        catch (Exception ex)
        {
            return new PageSpeedResult(url, strategy, DateTimeOffset.UtcNow,
                null, null, null, null, null,
                Error: $"{ex.GetType().Name}: {ex.Message}");
        }
    }

    private string BuildRequestUrl(string url, string strategy)
    {
        var sb = new StringBuilder(BaseUrl)
            .Append("?url=").Append(Uri.EscapeDataString(url))
            .Append("&strategy=").Append(strategy)
            .Append("&key=").Append(apiKey)
            .Append("&category=performance")
            .Append("&category=seo")
            .Append("&category=accessibility")
            .Append("&category=best-practices");
        return sb.ToString();
    }
}
