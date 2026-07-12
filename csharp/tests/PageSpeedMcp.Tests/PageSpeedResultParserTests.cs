using System.Text.Json;
using PageSpeedMcp.PageSpeed;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests the shared Lighthouse 13.4 response fixture contract.</summary>
public sealed class PageSpeedResultParserTests
{
    /// <summary>Verifies field data and lab data remain separate and correctly normalized.</summary>
    [Fact]
    public void Parse_Lighthouse134Fixture_SeparatesFieldAndLabData()
    {
        var raw = LoadFixture();

        var result = PageSpeedResultParser.Parse("https://example.test/page", "mobile", raw);

        Assert.Equal("13.4.0", result.Metadata.LighthouseVersion);
        Assert.Equal("https://example.test/final", result.Metadata.FinalUrl);
        Assert.Single(result.Metadata.RunWarnings);

        var page = Assert.IsType<FieldExperience>(result.FieldData?.Page);
        var origin = Assert.IsType<FieldExperience>(result.FieldData?.Origin);
        Assert.Equal(3100, page.Metrics["lcp"].Value);
        Assert.Equal(0.08, page.Metrics["cls"].Value);
        Assert.Equal("needs-improvement", page.Metrics["inp"].Rating);
        Assert.Equal("good", origin.Metrics["lcp"].Rating);

        var lab = Assert.IsType<LabData>(result.LabData);
        Assert.Equal("fraction", lab.Categories["agentic-browsing"].ScoreDisplayMode);
        Assert.Equal(420, lab.Metrics["serverResponseTime"].Value);
    }

    /// <summary>Verifies Lighthouse insights retain their structured details and savings.</summary>
    [Fact]
    public void Parse_Lighthouse134Fixture_PreservesInsightsAndAuditDetails()
    {
        var raw = LoadFixture();

        var result = PageSpeedResultParser.Parse("https://example.test/page", "mobile", raw);

        var lab = Assert.IsType<LabData>(result.LabData);
        var insight = Assert.Single(lab.Insights, audit => audit.Id == "render-blocking-insight");
        Assert.Equal(710, insight.MetricSavings?["LCP"]);
        Assert.Equal("table", insight.Details?.GetProperty("type").GetString());

        Assert.Contains(lab.Diagnostics, audit => audit.Id == "uses-text-compression");
        Assert.Contains("llms-txt", lab.PassedAuditIds);
        Assert.Contains("manual-audit", lab.ManualAuditIds);

        var encoded = JsonSerializer.Serialize(result, PsiJsonContext.Default.PageSpeedAnalysis);
        using var document = JsonDocument.Parse(encoded);
        Assert.Equal(
            JsonValueKind.Object,
            document.RootElement
                .GetProperty("labData")
                .GetProperty("insights")[0]
                .GetProperty("details")
                .ValueKind);
    }

    /// <summary>Verifies unavailable response sections remain absent instead of becoming zero values.</summary>
    [Fact]
    public void Parse_WithoutLighthouseOrFieldData_PreservesRequestMetadata()
    {
        var result = PageSpeedResultParser.Parse(
            "https://example.test",
            "desktop",
            new PsiApiResponse());

        Assert.Equal("https://example.test", result.Metadata.InputUrl);
        Assert.Equal("desktop", result.Metadata.Strategy);
        Assert.Null(result.FieldData);
        Assert.Null(result.LabData);
    }

    private static PsiApiResponse LoadFixture()
    {
        var path = Path.Combine(AppContext.BaseDirectory, "testdata", "psi-lighthouse-13.4.json");
        var json = File.ReadAllText(path);
        return JsonSerializer.Deserialize(json, PsiJsonContext.Default.PsiApiResponse)
            ?? throw new InvalidOperationException("PSI fixture deserialized to null.");
    }
}
