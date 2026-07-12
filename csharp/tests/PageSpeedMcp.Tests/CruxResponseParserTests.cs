using System.Text.Json;
using PageSpeedMcp.Crux;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests normalized current and historical CrUX response parsing.</summary>
public sealed class CruxResponseParserTests
{
    /// <summary>Verifies current histograms, percentiles, fractions, and normalization.</summary>
    [Fact]
    public void ParseCurrent_PreservesStatisticalAggregations()
    {
        using var document = LoadFixture("crux-current.json");

        var result = CruxResponseParser.ParseCurrent(document.RootElement);

        Assert.Equal("https://example.test/page", result.Target);
        Assert.Equal("url", result.TargetType);
        Assert.Equal("phone", result.FormFactor);
        Assert.Equal(3100, result.Metrics["largest_contentful_paint"].P75);
        Assert.Equal(0.08, result.Metrics["cumulative_layout_shift"].P75);
        Assert.Equal(
            0.15,
            result.Metrics["navigation_types"].Fractions?["back_forward_cache"]);
        Assert.Equal(
            "https://example.test/page",
            result.UrlNormalization?.NormalizedUrl);
    }

    /// <summary>Verifies unavailable history values become JSON-compatible nulls.</summary>
    [Fact]
    public void ParseHistory_ConvertsUnavailableValuesToNull()
    {
        using var document = LoadFixture("crux-history.json");

        var result = CruxResponseParser.ParseHistory(document.RootElement);

        Assert.Equal(3, result.CollectionPeriods.Count);
        var lcp = result.Metrics["largest_contentful_paint"];
        Assert.Null(lcp.P75?[1]);
        Assert.Null(lcp.Histogram[0].Densities[1]);
        Assert.Equal(0.09, result.Metrics["cumulative_layout_shift"].P75?[0]);
        Assert.Equal(
            0.66,
            result.Metrics["navigation_types"].Fractions?["navigate"][2]);

        var encoded = JsonSerializer.Serialize(
            result,
            CruxJsonContext.Default.CruxHistoryResult);
        Assert.Contains("\"p75\":[3200,null,2900]", encoded, StringComparison.Ordinal);
    }

    private static JsonDocument LoadFixture(string name)
    {
        var path = Path.Combine(AppContext.BaseDirectory, "testdata", name);
        return JsonDocument.Parse(File.ReadAllText(path));
    }
}
