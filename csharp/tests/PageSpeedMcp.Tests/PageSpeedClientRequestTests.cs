using System.Net;
using System.Text;
using System.Web;
using PageSpeedMcp.PageSpeed;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests PSI request validation and query construction.</summary>
public sealed class PageSpeedClientRequestTests
{
    private sealed class CapturingHandler : HttpMessageHandler
    {
        internal Uri? RequestUri { get; private set; }

        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            RequestUri = request.RequestUri;
            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(
                    """{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}""",
                    Encoding.UTF8,
                    "application/json"),
            });
        }
    }

    /// <summary>Verifies the four stable Lighthouse categories are requested by default.</summary>
    [Fact]
    public async Task AnalyzeAsync_DefaultCategories_AreSentToPsi()
    {
        var handler = new CapturingHandler();
        using var httpClient = new HttpClient(handler);
        var client = new PageSpeedClient(httpClient, "test-key");
        var request = PageSpeedAnalysisRequest.Create(
            "https://example.test",
            "mobile",
            categories: null,
            locale: null);

        await client.AnalyzeAsync(request);

        var query = HttpUtility.ParseQueryString(
            Assert.IsType<Uri>(handler.RequestUri).Query);
        var categories = query.GetValues("category");
        Assert.NotNull(categories);
        Assert.Equal(
            ["performance", "seo", "accessibility", "best-practices"],
            categories);
        Assert.Null(query["locale"]);
    }

    /// <summary>Verifies custom categories and locale are forwarded without duplicates.</summary>
    [Fact]
    public async Task AnalyzeAsync_CustomCategoriesAndLocale_AreSentToPsi()
    {
        var handler = new CapturingHandler();
        using var httpClient = new HttpClient(handler);
        var client = new PageSpeedClient(httpClient, "test-key");
        var request = PageSpeedAnalysisRequest.Create(
            "https://example.test",
            "desktop",
            ["agentic-browsing", "performance", "agentic-browsing"],
            "en-CA");

        await client.AnalyzeAsync(request);

        var query = HttpUtility.ParseQueryString(
            Assert.IsType<Uri>(handler.RequestUri).Query);
        var categories = query.GetValues("category");
        Assert.NotNull(categories);
        Assert.Equal(["agentic-browsing", "performance"], categories);
        Assert.Equal("en-CA", query["locale"]);
    }

    /// <summary>Verifies invalid URL, strategy, and category values fail before HTTP.</summary>
    [Theory]
    [InlineData("/relative", "mobile", null)]
    [InlineData("ftp://example.test", "mobile", null)]
    [InlineData("https://example.test", "tablet", null)]
    [InlineData("https://example.test", "mobile", "pwa")]
    public void Create_InvalidInput_IsRejected(string url, string strategy, string? category)
    {
        var categories = category is null ? null : new[] { category };

        Assert.Throws<ArgumentException>(
            () => PageSpeedAnalysisRequest.Create(url, strategy, categories, null));
    }

    /// <summary>Verifies empty and explicit strategy selections resolve predictably.</summary>
    [Fact]
    public void ResolveStrategies_ValidSelections_AreExpanded()
    {
        Assert.Equal(
            ["mobile", "desktop"],
            PageSpeedAnalysisRequest.ResolveStrategies(null));
        Assert.Equal(
            ["mobile", "desktop"],
            PageSpeedAnalysisRequest.ResolveStrategies("both"));
        Assert.Equal(["mobile"], PageSpeedAnalysisRequest.ResolveStrategies("MOBILE"));
        Assert.Equal(["desktop"], PageSpeedAnalysisRequest.ResolveStrategies("desktop"));
        Assert.Throws<ArgumentException>(
            () => PageSpeedAnalysisRequest.ResolveStrategies("tablet"));
    }
}
