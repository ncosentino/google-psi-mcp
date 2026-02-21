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
        var requestUrl = BuildRequestUrl(url, strategy);
        var response = await httpClient.GetAsync(requestUrl, cancellationToken).ConfigureAwait(false);
        response.EnsureSuccessStatusCode();

        var raw = await response.Content
            .ReadFromJsonAsync(PsiJsonContext.Default.PsiApiResponse, cancellationToken)
            .ConfigureAwait(false);

        return PageSpeedResultParser.Parse(url, strategy, raw);
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
