using System.Text.Json;
using System.Text.Json.Serialization;

namespace PageSpeedMcp.PageSpeed;

/// <summary>One PageSpeed Insights analysis for a URL and strategy.</summary>
internal sealed record PageSpeedAnalysis(
    AnalysisMetadata Metadata,
    FieldData? FieldData,
    LabData? LabData);

/// <summary>Results and per-request failures from an MCP analysis call.</summary>
internal sealed record AnalysisResponse(
    IReadOnlyList<PageSpeedAnalysis> Results,
    IReadOnlyList<AnalysisFailure> Errors);

/// <summary>One failed URL and strategy analysis.</summary>
internal sealed record AnalysisFailure(
    string InputUrl,
    string Strategy,
    string Code,
    string Message,
    bool Retryable);

/// <summary>Source and timing metadata for a PageSpeed Insights result.</summary>
internal sealed record AnalysisMetadata(
    string InputUrl,
    string Strategy,
    DateTimeOffset? AnalysisTimestamp,
    DateTimeOffset? FetchTime,
    string? LighthouseVersion,
    string? RequestedUrl,
    string? FinalUrl,
    string? FinalDisplayedUrl,
    string? MainDocumentUrl,
    IReadOnlyList<string> RunWarnings,
    LighthouseRuntimeError? RuntimeError);

/// <summary>A fatal Lighthouse runtime failure.</summary>
internal sealed record LighthouseRuntimeError(string Code, string Message);

/// <summary>Page-level and origin-level real-user measurements.</summary>
internal sealed record FieldData(FieldExperience? Page, FieldExperience? Origin);

/// <summary>One PSI Chrome UX Report loading experience.</summary>
internal sealed record FieldExperience(
    string Id,
    string? InitialUrl,
    string? OverallRating,
    bool OriginFallback,
    IReadOnlyDictionary<string, FieldMetric> Metrics);

/// <summary>A p75 real-user metric and its distribution.</summary>
internal sealed record FieldMetric(
    string Id,
    int Percentile,
    double Value,
    string? Unit,
    string Rating,
    IReadOnlyList<FieldDistribution> Distributions);

/// <summary>One real-user metric histogram bucket.</summary>
internal sealed record FieldDistribution(double? Min, double? Max, double Proportion);

/// <summary>Normalized Lighthouse category, metric, and audit results.</summary>
internal sealed record LabData(
    IReadOnlyDictionary<string, CategoryResult> Categories,
    IReadOnlyDictionary<string, LabMetric> Metrics,
    IReadOnlyList<LighthouseAudit> Insights,
    IReadOnlyList<LighthouseAudit> Diagnostics,
    IReadOnlyList<LighthouseAudit> UnscoredAudits,
    IReadOnlyList<string> PassedAuditIds,
    IReadOnlyList<string> NotApplicableAuditIds,
    IReadOnlyList<string> ManualAuditIds,
    IReadOnlyList<LighthouseEntity> Entities);

/// <summary>One Lighthouse category result.</summary>
internal sealed record CategoryResult(
    string Id,
    string Title,
    string? Description,
    double? Score,
    string? ScoreDisplayMode);

/// <summary>One synthetic Lighthouse metric result.</summary>
internal sealed record LabMetric(
    string Id,
    string Title,
    string? Description,
    double? Score,
    double? Value,
    string? Unit,
    string? DisplayValue);

/// <summary>An actionable or informative Lighthouse audit.</summary>
internal sealed record LighthouseAudit(
    string Id,
    string Title,
    string? Description,
    double? Score,
    string? ScoreDisplayMode,
    string? DisplayValue,
    string? Explanation,
    string? ErrorMessage,
    IReadOnlyList<string> Warnings,
    double? NumericValue,
    string? NumericUnit,
    IReadOnlyDictionary<string, double>? MetricSavings,
    JsonElement? Details);

/// <summary>A first-party or third-party entity identified by Lighthouse.</summary>
internal sealed record LighthouseEntity(
    string Name,
    string? Category,
    string? Homepage,
    IReadOnlyList<string> Origins,
    bool IsFirstParty,
    bool IsUnrecognized);

internal sealed class PsiApiResponse
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("analysisUTCTimestamp")]
    public string? AnalysisUtcTimestamp { get; set; }

    [JsonPropertyName("loadingExperience")]
    public LoadingExperienceRaw? LoadingExperience { get; set; }

    [JsonPropertyName("originLoadingExperience")]
    public LoadingExperienceRaw? OriginLoadingExperience { get; set; }

    [JsonPropertyName("lighthouseResult")]
    public LighthouseResultRaw? LighthouseResult { get; set; }
}

internal sealed class LoadingExperienceRaw
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("initial_url")]
    public string? InitialUrl { get; set; }

    [JsonPropertyName("overall_category")]
    public string? OverallCategory { get; set; }

    [JsonPropertyName("origin_fallback")]
    public bool OriginFallback { get; set; }

    [JsonPropertyName("metrics")]
    public Dictionary<string, FieldMetricRaw>? Metrics { get; set; }
}

internal sealed class FieldMetricRaw
{
    [JsonPropertyName("percentile")]
    public double Percentile { get; set; }

    [JsonPropertyName("distributions")]
    public List<FieldDistributionRaw>? Distributions { get; set; }

    [JsonPropertyName("category")]
    public string? Category { get; set; }
}

internal sealed class FieldDistributionRaw
{
    [JsonPropertyName("min")]
    public double? Min { get; set; }

    [JsonPropertyName("max")]
    public double? Max { get; set; }

    [JsonPropertyName("proportion")]
    public double Proportion { get; set; }
}

internal sealed class LighthouseResultRaw
{
    [JsonPropertyName("requestedUrl")]
    public string? RequestedUrl { get; set; }

    [JsonPropertyName("finalUrl")]
    public string? FinalUrl { get; set; }

    [JsonPropertyName("finalDisplayedUrl")]
    public string? FinalDisplayedUrl { get; set; }

    [JsonPropertyName("mainDocumentUrl")]
    public string? MainDocumentUrl { get; set; }

    [JsonPropertyName("lighthouseVersion")]
    public string? LighthouseVersion { get; set; }

    [JsonPropertyName("fetchTime")]
    public string? FetchTime { get; set; }

    [JsonPropertyName("runWarnings")]
    public JsonElement[]? RunWarnings { get; set; }

    [JsonPropertyName("runtimeError")]
    public LighthouseRuntimeError? RuntimeError { get; set; }

    [JsonPropertyName("categories")]
    public Dictionary<string, CategoryRaw>? Categories { get; set; }

    [JsonPropertyName("audits")]
    public Dictionary<string, AuditRaw>? Audits { get; set; }

    [JsonPropertyName("entities")]
    public List<LighthouseEntity>? Entities { get; set; }
}

internal sealed class CategoryRaw
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("title")]
    public string? Title { get; set; }

    [JsonPropertyName("description")]
    public string? Description { get; set; }

    [JsonPropertyName("score")]
    public double? Score { get; set; }

    [JsonPropertyName("categoryScoreDisplayMode")]
    public string? CategoryScoreDisplayMode { get; set; }

    [JsonPropertyName("auditRefs")]
    public List<AuditRefRaw>? AuditRefs { get; set; }
}

internal sealed class AuditRefRaw
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("group")]
    public string? Group { get; set; }
}

internal sealed class AuditRaw
{
    [JsonPropertyName("id")]
    public string? Id { get; set; }

    [JsonPropertyName("title")]
    public string? Title { get; set; }

    [JsonPropertyName("description")]
    public string? Description { get; set; }

    [JsonPropertyName("score")]
    public double? Score { get; set; }

    [JsonPropertyName("scoreDisplayMode")]
    public string? ScoreDisplayMode { get; set; }

    [JsonPropertyName("displayValue")]
    public string? DisplayValue { get; set; }

    [JsonPropertyName("explanation")]
    public string? Explanation { get; set; }

    [JsonPropertyName("errorMessage")]
    public string? ErrorMessage { get; set; }

    [JsonPropertyName("warnings")]
    public JsonElement[]? Warnings { get; set; }

    [JsonPropertyName("numericValue")]
    public double? NumericValue { get; set; }

    [JsonPropertyName("numericUnit")]
    public string? NumericUnit { get; set; }

    [JsonPropertyName("metricSavings")]
    public Dictionary<string, double>? MetricSavings { get; set; }

    [JsonPropertyName("details")]
    public JsonElement? Details { get; set; }
}

/// <summary>System.Text.Json source generation context for AOT-safe serialization.</summary>
[JsonSerializable(typeof(PsiApiResponse))]
[JsonSerializable(typeof(AnalysisResponse))]
[JsonSerializable(typeof(PageSpeedAnalysis))]
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    WriteIndented = false,
    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull)]
internal partial class PsiJsonContext : JsonSerializerContext;
