using System.Net.Http.Headers;
using System.Text.Json;
using PageSpeedMcp.Infrastructure;

namespace PageSpeedMcp.Crux;

/// <summary>Client for current and historical Chrome UX Report data.</summary>
internal sealed class CruxClient
{
    private const string DefaultCurrentApiUrl =
        "https://chromeuxreport.googleapis.com/v1/records:queryRecord";
    private const string DefaultHistoryApiUrl =
        "https://chromeuxreport.googleapis.com/v1/records:queryHistoryRecord";

    private readonly HttpClient _httpClient;
    private readonly string _apiKey;
    private readonly string _currentApiUrl;
    private readonly string _historyApiUrl;

    /// <summary>Creates a client using the production CrUX endpoints.</summary>
    internal CruxClient(HttpClient httpClient, string apiKey)
        : this(httpClient, apiKey, DefaultCurrentApiUrl, DefaultHistoryApiUrl)
    {
    }

    /// <summary>Creates a client using explicit endpoints for testing.</summary>
    internal CruxClient(
        HttpClient httpClient,
        string apiKey,
        string currentApiUrl,
        string historyApiUrl)
    {
        _httpClient = httpClient;
        _apiKey = apiKey;
        _currentApiUrl = currentApiUrl;
        _historyApiUrl = historyApiUrl;
    }

    /// <summary>Queries current 28-day Chrome UX Report data.</summary>
    internal async Task<CruxResult> QueryCurrentAsync(
        CruxQueryRequest request,
        CancellationToken cancellationToken = default)
    {
        using var document = await PostAsync(
            _currentApiUrl,
            CreateBody(request, includeHistoryCount: false),
            cancellationToken).ConfigureAwait(false);
        return CruxResponseParser.ParseCurrent(document.RootElement);
    }

    /// <summary>Queries historical Chrome UX Report timeseries data.</summary>
    internal async Task<CruxHistoryResult> QueryHistoryAsync(
        CruxQueryRequest request,
        CancellationToken cancellationToken = default)
    {
        using var document = await PostAsync(
            _historyApiUrl,
            CreateBody(request, includeHistoryCount: true),
            cancellationToken).ConfigureAwait(false);
        return CruxResponseParser.ParseHistory(document.RootElement);
    }

    private static CruxQueryBody CreateBody(
        CruxQueryRequest request,
        bool includeHistoryCount) =>
        new(
            Url: request.TargetType == "url" ? request.Target : null,
            Origin: request.TargetType == "origin" ? request.Target : null,
            FormFactor: request.FormFactor == "all"
                ? null
                : request.FormFactor.ToUpperInvariant(),
            Metrics: request.Metrics.Count == 0 ? null : request.Metrics,
            CollectionPeriodCount: includeHistoryCount
                ? request.CollectionPeriodCount == 0 ? 25 : request.CollectionPeriodCount
                : null);

    private async Task<JsonDocument> PostAsync(
        string endpoint,
        CruxQueryBody body,
        CancellationToken cancellationToken)
    {
        var requestUri = $"{endpoint}?key={Uri.EscapeDataString(_apiKey)}";
        var encodedBody = JsonSerializer.SerializeToUtf8Bytes(
            body,
            CruxJsonContext.Default.CruxQueryBody);
        using var response = await _httpClient
            .SendWithRetryAsync(
                () =>
                {
                    var request = new HttpRequestMessage(HttpMethod.Post, requestUri);
                    request.Content = new ByteArrayContent(encodedBody);
                    request.Content.Headers.ContentType =
                        new MediaTypeHeaderValue("application/json");
                    return request;
                },
                cancellationToken)
            .ConfigureAwait(false);

        if (!response.IsSuccessStatusCode)
        {
            var responseBody = await response.Content
                .ReadAsStringAsync(cancellationToken)
                .ConfigureAwait(false);
            var snippet = responseBody.Length > 500
                ? responseBody[..500] + "..."
                : responseBody;
            throw new HttpRequestException(
                $"CrUX API returned HTTP {(int)response.StatusCode} {response.StatusCode}: {snippet}",
                inner: null,
                response.StatusCode);
        }

        await using var stream = await response.Content
            .ReadAsStreamAsync(cancellationToken)
            .ConfigureAwait(false);
        return await JsonDocument
            .ParseAsync(stream, cancellationToken: cancellationToken)
            .ConfigureAwait(false);
    }
}
