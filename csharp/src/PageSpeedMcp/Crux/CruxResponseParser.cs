using System.Globalization;
using System.Text.Json;

namespace PageSpeedMcp.Crux;

/// <summary>Parses current and historical Chrome UX Report responses.</summary>
internal static class CruxResponseParser
{
    /// <summary>Parses a current CrUX response.</summary>
    internal static CruxResult ParseCurrent(JsonElement root)
    {
        var record = root.GetProperty("record");
        var (target, targetType, formFactor) = ParseKey(record.GetProperty("key"));
        var metrics = new Dictionary<string, CruxMetric>(StringComparer.Ordinal);

        foreach (var metricProperty in record.GetProperty("metrics").EnumerateObject())
        {
            var metric = metricProperty.Value;
            var histogram = metric.TryGetProperty("histogram", out var histogramElement)
                ? histogramElement.EnumerateArray()
                    .Select(bin => new CruxHistogramBin(
                        ReadOptionalNumber(bin, "start"),
                        ReadOptionalNumber(bin, "end"),
                        bin.GetProperty("density").GetDouble()))
                    .ToArray()
                : [];

            double? p75 = null;
            if (metric.TryGetProperty("percentiles", out var percentiles) &&
                percentiles.TryGetProperty("p75", out var p75Element))
            {
                p75 = ReadFlexibleNumber(p75Element);
            }

            IReadOnlyDictionary<string, double>? fractions = null;
            if (metric.TryGetProperty("fractions", out var fractionsElement))
            {
                fractions = fractionsElement.EnumerateObject()
                    .ToDictionary(
                        property => property.Name,
                        property => property.Value.GetDouble(),
                        StringComparer.Ordinal);
            }

            metrics[metricProperty.Name] = new CruxMetric(histogram, p75, fractions);
        }

        CruxUrlNormalization? normalization = null;
        if (root.TryGetProperty("urlNormalizationDetails", out var normalizationElement))
        {
            normalization = new CruxUrlNormalization(
                normalizationElement.GetProperty("originalUrl").GetString() ?? string.Empty,
                normalizationElement.GetProperty("normalizedUrl").GetString() ?? string.Empty);
        }

        return new CruxResult(
            target,
            targetType,
            formFactor,
            ParseCollectionPeriod(record.GetProperty("collectionPeriod")),
            metrics,
            normalization);
    }

    /// <summary>Parses a historical CrUX response.</summary>
    internal static CruxHistoryResult ParseHistory(JsonElement root)
    {
        var record = root.GetProperty("record");
        var (target, targetType, formFactor) = ParseKey(record.GetProperty("key"));
        var periods = record.GetProperty("collectionPeriods")
            .EnumerateArray()
            .Select(ParseCollectionPeriod)
            .ToArray();
        var metrics = new Dictionary<string, CruxHistoryMetric>(StringComparer.Ordinal);

        foreach (var metricProperty in record.GetProperty("metrics").EnumerateObject())
        {
            var metric = metricProperty.Value;
            var histogram = metric.TryGetProperty("histogramTimeseries", out var histogramElement)
                ? histogramElement.EnumerateArray()
                    .Select(bin => new CruxHistoryHistogramBin(
                        ReadOptionalNumber(bin, "start"),
                        ReadOptionalNumber(bin, "end"),
                        bin.GetProperty("densities")
                            .EnumerateArray()
                            .Select(ReadFlexibleNumber)
                            .ToArray()))
                    .ToArray()
                : [];

            IReadOnlyList<double?>? p75 = null;
            if (metric.TryGetProperty("percentilesTimeseries", out var percentiles) &&
                percentiles.TryGetProperty("p75s", out var p75Element))
            {
                p75 = p75Element.EnumerateArray().Select(ReadFlexibleNumber).ToArray();
            }

            IReadOnlyDictionary<string, IReadOnlyList<double?>>? fractions = null;
            if (metric.TryGetProperty("fractionTimeseries", out var fractionsElement))
            {
                fractions = fractionsElement.EnumerateObject()
                    .ToDictionary(
                        property => property.Name,
                        property => (IReadOnlyList<double?>)property.Value
                            .GetProperty("fractions")
                            .EnumerateArray()
                            .Select(ReadFlexibleNumber)
                            .ToArray(),
                        StringComparer.Ordinal);
            }

            metrics[metricProperty.Name] = new CruxHistoryMetric(histogram, p75, fractions);
        }

        return new CruxHistoryResult(target, targetType, formFactor, periods, metrics);
    }

    private static (string Target, string TargetType, string? FormFactor) ParseKey(
        JsonElement key)
    {
        var targetType = key.TryGetProperty("url", out var urlElement) ? "url" : "origin";
        var target = targetType == "url"
            ? urlElement.GetString() ?? string.Empty
            : key.GetProperty("origin").GetString() ?? string.Empty;
        var formFactor = key.TryGetProperty("formFactor", out var formFactorElement)
            ? formFactorElement.GetString()?.ToLowerInvariant()
            : null;
        return (target, targetType, formFactor);
    }

    private static CruxCollectionPeriod ParseCollectionPeriod(JsonElement element) =>
        new(ParseDate(element.GetProperty("firstDate")), ParseDate(element.GetProperty("lastDate")));

    private static CruxDate ParseDate(JsonElement element) =>
        new(
            element.GetProperty("year").GetInt32(),
            element.GetProperty("month").GetInt32(),
            element.GetProperty("day").GetInt32());

    private static double? ReadOptionalNumber(JsonElement element, string propertyName) =>
        element.TryGetProperty(propertyName, out var value) ? ReadFlexibleNumber(value) : null;

    private static double? ReadFlexibleNumber(JsonElement element)
    {
        if (element.ValueKind == JsonValueKind.Null)
        {
            return null;
        }
        if (element.ValueKind == JsonValueKind.Number)
        {
            return element.GetDouble();
        }
        if (element.ValueKind == JsonValueKind.String)
        {
            var value = element.GetString();
            if (string.Equals(value, "NaN", StringComparison.Ordinal))
            {
                return null;
            }
            if (double.TryParse(
                value,
                NumberStyles.Float,
                CultureInfo.InvariantCulture,
                out var parsed))
            {
                return parsed;
            }
        }
        throw new JsonException($"Expected a number, numeric string, null, or NaN but found {element}.");
    }
}
