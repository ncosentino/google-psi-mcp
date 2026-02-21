using System.Text.Json.Serialization;

namespace PageSpeedMcp.PageSpeed;

/// <summary>A parsed PageSpeed Insights result for a single URL and strategy.</summary>
internal sealed record PageSpeedResult(
    string Url,
    string Strategy,
    DateTimeOffset AnalyzedAt,
    CategoryScores? Scores,
    CoreWebVitals? CoreWebVitals,
    IReadOnlyList<Opportunity>? Opportunities,
    IReadOnlyList<AuditResult>? FailingAudits,
    IReadOnlyList<string>? PassedAuditIds,
    string? Error = null);

/// <summary>Category scores in the range 0-100.</summary>
internal sealed record CategoryScores(int Performance, int Seo, int Accessibility, int BestPractices);

/// <summary>Core Web Vitals metrics.</summary>
internal sealed record CoreWebVitals(
    MetricValue Fcp,
    MetricValue Lcp,
    MetricValue Cls,
    MetricValue Tbt,
    MetricValue Ttfb,
    MetricValue SpeedIndex);

/// <summary>A single metric reading.</summary>
internal sealed record MetricValue(double Value, string? Unit, string Rating);

/// <summary>A performance improvement opportunity.</summary>
internal sealed record Opportunity(string Id, string Title, string Description, string? Savings, string Impact);

/// <summary>A failing Lighthouse audit.</summary>
internal sealed record AuditResult(string Id, string Title, string Description, double Score, string? DisplayValue);

// --- PSI API raw response types ---

internal sealed class PsiApiResponse
{
    [JsonPropertyName("lighthouseResult")]
    public LighthouseResult? LighthouseResult { get; set; }
}

internal sealed class LighthouseResult
{
    [JsonPropertyName("categories")]
    public Dictionary<string, CategoryRaw>? Categories { get; set; }

    [JsonPropertyName("audits")]
    public Dictionary<string, AuditRaw>? Audits { get; set; }
}

internal sealed class CategoryRaw
{
    [JsonPropertyName("score")]
    public double Score { get; set; }
}

internal sealed class AuditRaw
{
    [JsonPropertyName("id")]
    public string Id { get; set; } = string.Empty;

    [JsonPropertyName("title")]
    public string Title { get; set; } = string.Empty;

    [JsonPropertyName("description")]
    public string Description { get; set; } = string.Empty;

    [JsonPropertyName("score")]
    public double? Score { get; set; }

    [JsonPropertyName("numericValue")]
    public double? NumericValue { get; set; }

    [JsonPropertyName("displayValue")]
    public string? DisplayValue { get; set; }

    [JsonPropertyName("details")]
    public AuditDetailsRaw? Details { get; set; }
}

internal sealed class AuditDetailsRaw
{
    [JsonPropertyName("type")]
    public string? Type { get; set; }
}

/// <summary>System.Text.Json source generation context for AOT-safe serialization.</summary>
[JsonSerializable(typeof(PsiApiResponse))]
[JsonSerializable(typeof(PageSpeedResult[]))]
[JsonSerializable(typeof(PageSpeedResult))]
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    WriteIndented = false,
    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull)]
internal partial class PsiJsonContext : JsonSerializerContext;
