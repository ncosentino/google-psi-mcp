using System.Net;
using System.Text;
using System.Text.Json;
using PageSpeedMcp.Infrastructure;
using PageSpeedMcp.PageSpeed;
using PageSpeedMcp.Tools;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests quota-safe batching and structured partial failures.</summary>
public sealed class PageSpeedToolBatchTests
{
    private sealed class TrackingHandler(
        HttpStatusCode statusCode,
        string responseBody,
        TimeSpan delay) : HttpMessageHandler
    {
        private int _active;
        private int _maxActive;
        private int _requests;

        internal int MaxActive => _maxActive;
        internal int Requests => _requests;

        protected override async Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            Interlocked.Increment(ref _requests);
            var active = Interlocked.Increment(ref _active);
            while (true)
            {
                var currentMax = Volatile.Read(ref _maxActive);
                if (active <= currentMax ||
                    Interlocked.CompareExchange(ref _maxActive, active, currentMax) == currentMax)
                {
                    break;
                }
            }

            try
            {
                await Task.Delay(delay, cancellationToken);
                return new HttpResponseMessage(statusCode)
                {
                    Content = new StringContent(responseBody, Encoding.UTF8, "application/json"),
                };
            }
            finally
            {
                Interlocked.Decrement(ref _active);
            }
        }
    }

    /// <summary>Verifies both-strategy batches run concurrently without exceeding four requests.</summary>
    [Fact]
    public async Task AnalyzePages_BoundsConcurrencyAndReturnsAllResults()
    {
        var handler = new TrackingHandler(
            HttpStatusCode.OK,
            """{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}""",
            TimeSpan.FromMilliseconds(10));
        using var httpClient = new HttpClient(handler);
        using var limiter = new PageSpeedRequestLimiter();
        var tool = new PageSpeedTool(
            new PageSpeedClient(httpClient, "test-key"),
            limiter);

        var json = await tool.AnalyzePages(
            [
                "https://example.test/one",
                "https://example.test/two",
                "https://example.test/three",
            ],
            strategy: "both",
            categories: ["performance"],
            locale: null);

        using var document = JsonDocument.Parse(json);
        Assert.Equal(6, document.RootElement.GetProperty("results").GetArrayLength());
        Assert.Empty(document.RootElement.GetProperty("errors").EnumerateArray());
        Assert.InRange(handler.MaxActive, 2, 4);
    }

    /// <summary>Verifies oversized batches are rejected before making API requests.</summary>
    [Fact]
    public async Task AnalyzePages_OversizedBatch_IsRejectedBeforeHttp()
    {
        var handler = new TrackingHandler(
            HttpStatusCode.OK,
            """{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}""",
            TimeSpan.Zero);
        using var httpClient = new HttpClient(handler);
        using var limiter = new PageSpeedRequestLimiter();
        var tool = new PageSpeedTool(
            new PageSpeedClient(httpClient, "test-key"),
            limiter);
        var urls = Enumerable
            .Range(0, 11)
            .Select(index => $"https://example.test/{index}")
            .ToArray();

        await Assert.ThrowsAsync<ArgumentException>(
            () => tool.AnalyzePages(urls, "mobile", null, null));
        Assert.Equal(0, handler.Requests);
    }

    /// <summary>Verifies non-transient API failures are returned with structured metadata.</summary>
    [Fact]
    public async Task AnalyzePage_UpstreamBadRequest_ReturnsStructuredFailure()
    {
        var handler = new TrackingHandler(
            HttpStatusCode.BadRequest,
            """{"error":{"message":"invalid request"}}""",
            TimeSpan.Zero);
        using var httpClient = new HttpClient(handler);
        using var limiter = new PageSpeedRequestLimiter();
        var tool = new PageSpeedTool(
            new PageSpeedClient(httpClient, "test-key"),
            limiter);

        var json = await tool.AnalyzePage(
            "https://example.test",
            "mobile",
            ["performance"],
            null);

        using var document = JsonDocument.Parse(json);
        var failure = document.RootElement.GetProperty("errors")[0];
        Assert.Equal("upstream_rejected", failure.GetProperty("code").GetString());
        Assert.False(failure.GetProperty("retryable").GetBoolean());
        Assert.Equal(1, handler.Requests);
    }
}
