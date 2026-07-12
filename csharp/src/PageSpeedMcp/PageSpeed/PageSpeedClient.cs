using System.Net.Http.Json;
using System.Text;
using PageSpeedMcp.Infrastructure;

namespace PageSpeedMcp.PageSpeed;

/// <summary>Client for the Google PageSpeed Insights API v5.</summary>
internal sealed class PageSpeedClient(HttpClient httpClient, string apiKey)
{
    private const string BaseUrl = "https://www.googleapis.com/pagespeedonline/v5/runPagespeed";

    /// <summary>Analyzes the specified URL using the given strategy.</summary>
    /// <param name="url">The URL to analyze.</param>
    /// <param name="strategy">"mobile" or "desktop".</param>
    /// <param name="cancellationToken">Cancellation token.</param>
    public async Task<PageSpeedAnalysis> AnalyzeAsync(
        PageSpeedAnalysisRequest request,
        CancellationToken cancellationToken = default)
    {
        var requestUrl = BuildRequestUrl(request);
        using var response = await httpClient
            .SendWithRetryAsync(
                () => new HttpRequestMessage(HttpMethod.Get, requestUrl),
                cancellationToken)
            .ConfigureAwait(false);

        if (!response.IsSuccessStatusCode)
        {
            var body = await response.Content.ReadAsStringAsync(cancellationToken).ConfigureAwait(false);
            var snippet = body.Length > 300 ? body[..300] + "..." : body;
            throw new HttpRequestException(
                $"PSI API returned HTTP {(int)response.StatusCode} {response.StatusCode}: {snippet}",
                inner: null,
                response.StatusCode);
        }

        var raw = await response.Content
            .ReadFromJsonAsync(PsiJsonContext.Default.PsiApiResponse, cancellationToken)
            .ConfigureAwait(false);

        return PageSpeedResultParser.Parse(request.Url, request.Strategy, raw);
    }

    private string BuildRequestUrl(PageSpeedAnalysisRequest request)
    {
        var sb = new StringBuilder(BaseUrl)
            .Append("?url=").Append(Uri.EscapeDataString(request.Url))
            .Append("&strategy=").Append(request.Strategy)
            .Append("&key=").Append(apiKey);

        foreach (var category in request.Categories)
        {
            sb.Append("&category=").Append(Uri.EscapeDataString(category));
        }
        if (request.Locale is not null)
        {
            sb.Append("&locale=").Append(Uri.EscapeDataString(request.Locale));
        }
        return sb.ToString();
    }
}
