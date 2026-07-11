using System.Text.Json.Serialization;

namespace PageSpeedMcp.Crux;

/// <summary>Current Chrome UX Report data for one URL or origin.</summary>
internal sealed record CruxResult(
    string Target,
    string TargetType,
    string? FormFactor,
    CruxCollectionPeriod CollectionPeriod,
    IReadOnlyDictionary<string, CruxMetric> Metrics,
    CruxUrlNormalization? UrlNormalization);

/// <summary>Historical Chrome UX Report timeseries data.</summary>
internal sealed record CruxHistoryResult(
    string Target,
    string TargetType,
    string? FormFactor,
    IReadOnlyList<CruxCollectionPeriod> CollectionPeriods,
    IReadOnlyDictionary<string, CruxHistoryMetric> Metrics);

/// <summary>One CrUX aggregation window.</summary>
internal sealed record CruxCollectionPeriod(CruxDate FirstDate, CruxDate LastDate);

/// <summary>A calendar date returned by CrUX.</summary>
internal sealed record CruxDate(int Year, int Month, int Day);

/// <summary>URL normalization performed by CrUX.</summary>
internal sealed record CruxUrlNormalization(string OriginalUrl, string NormalizedUrl);

/// <summary>Current statistical aggregations for one CrUX metric.</summary>
internal sealed record CruxMetric(
    IReadOnlyList<CruxHistogramBin> Histogram,
    double? P75,
    IReadOnlyDictionary<string, double>? Fractions);

/// <summary>One current CrUX histogram bucket.</summary>
internal sealed record CruxHistogramBin(double? Start, double? End, double Density);

/// <summary>Historical statistical aggregations for one CrUX metric.</summary>
internal sealed record CruxHistoryMetric(
    IReadOnlyList<CruxHistoryHistogramBin> Histogram,
    IReadOnlyList<double?>? P75,
    IReadOnlyDictionary<string, IReadOnlyList<double?>>? Fractions);

/// <summary>One CrUX histogram bucket across collection periods.</summary>
internal sealed record CruxHistoryHistogramBin(
    double? Start,
    double? End,
    IReadOnlyList<double?> Densities);

internal sealed record CruxQueryBody(
    [property: JsonPropertyName("url")] string? Url,
    [property: JsonPropertyName("origin")] string? Origin,
    [property: JsonPropertyName("formFactor")] string? FormFactor,
    [property: JsonPropertyName("metrics")] IReadOnlyList<string>? Metrics,
    [property: JsonPropertyName("collectionPeriodCount")] int? CollectionPeriodCount);

/// <summary>System.Text.Json source generation context for CrUX requests and responses.</summary>
[JsonSerializable(typeof(CruxQueryBody))]
[JsonSerializable(typeof(CruxResult))]
[JsonSerializable(typeof(CruxHistoryResult))]
[JsonSourceGenerationOptions(
    PropertyNamingPolicy = JsonKnownNamingPolicy.CamelCase,
    WriteIndented = false,
    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull)]
internal partial class CruxJsonContext : JsonSerializerContext;
