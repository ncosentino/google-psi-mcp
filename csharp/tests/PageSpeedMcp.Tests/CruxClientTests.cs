using System.Net;
using System.Text;
using System.Text.Json;
using PageSpeedMcp.Crux;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests CrUX request validation and HTTP request construction.</summary>
public sealed class CruxClientTests
{
    private sealed class CapturingHandler(string responseBody) : HttpMessageHandler
    {
        internal JsonDocument? RequestBody { get; private set; }

        protected override async Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            var json = await request.Content!
                .ReadAsStringAsync(cancellationToken)
                .ConfigureAwait(false);
            RequestBody = JsonDocument.Parse(json);
            return new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(responseBody, Encoding.UTF8, "application/json"),
            };
        }
    }

    /// <summary>Verifies current URL, form-factor, and metric inputs are forwarded.</summary>
    [Fact]
    public async Task QueryCurrentAsync_SendsValidatedRequest()
    {
        var handler = new CapturingHandler(ReadFixture("crux-current.json"));
        using var httpClient = new HttpClient(handler);
        var client = new CruxClient(
            httpClient,
            "test-key",
            "https://example.test/current",
            "https://example.test/history");
        var request = CruxQueryRequest.Create(
            "https://example.test/page",
            "url",
            "phone",
            ["largest_contentful_paint", "largest_contentful_paint"],
            0);

        var result = await client.QueryCurrentAsync(request);

        Assert.Equal("https://example.test/page", result.Target);
        var body = Assert.IsType<JsonDocument>(handler.RequestBody).RootElement;
        Assert.Equal("PHONE", body.GetProperty("formFactor").GetString());
        Assert.Single(body.GetProperty("metrics").EnumerateArray());
        Assert.False(body.TryGetProperty("collectionPeriodCount", out _));
    }

    /// <summary>Verifies history requests include their requested collection-period count.</summary>
    [Fact]
    public async Task QueryHistoryAsync_SendsCollectionPeriodCount()
    {
        var handler = new CapturingHandler(ReadFixture("crux-history.json"));
        using var httpClient = new HttpClient(handler);
        var client = new CruxClient(
            httpClient,
            "test-key",
            "https://example.test/current",
            "https://example.test/history");
        var request = CruxQueryRequest.Create(
            "https://example.test",
            "origin",
            "desktop",
            metrics: null,
            collectionPeriodCount: 3);

        await client.QueryHistoryAsync(request);

        var body = Assert.IsType<JsonDocument>(handler.RequestBody).RootElement;
        Assert.Equal(3, body.GetProperty("collectionPeriodCount").GetInt32());
        Assert.False(body.TryGetProperty("metrics", out _));
    }

    /// <summary>Verifies invalid target, dimension, metric, and count values are rejected.</summary>
    [Theory]
    [InlineData("/relative", "url", "all", null, 0)]
    [InlineData("https://example.test/path", "origin", "all", null, 0)]
    [InlineData("https://example.test", "site", "all", null, 0)]
    [InlineData("https://example.test", "origin", "watch", null, 0)]
    [InlineData("https://example.test", "origin", "all", "LCP!", 0)]
    [InlineData("https://example.test", "origin", "all", null, 41)]
    public void Create_InvalidInput_IsRejected(
        string target,
        string targetType,
        string formFactor,
        string? metric,
        int collectionPeriodCount)
    {
        var metrics = metric is null ? null : new[] { metric };

        Assert.ThrowsAny<ArgumentException>(() => CruxQueryRequest.Create(
            target,
            targetType,
            formFactor,
            metrics,
            collectionPeriodCount));
    }

    private static string ReadFixture(string name)
    {
        var path = Path.Combine(AppContext.BaseDirectory, "testdata", name);
        return File.ReadAllText(path);
    }
}
