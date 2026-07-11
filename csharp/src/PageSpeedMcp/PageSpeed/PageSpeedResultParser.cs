using System.Globalization;
using System.Text.Json;

namespace PageSpeedMcp.PageSpeed;

/// <summary>Parses PSI responses into the stable MCP response contract.</summary>
internal static class PageSpeedResultParser
{
    private sealed record FieldMetricDefinition(string Name, string? Unit, double Scale);

    private static readonly IReadOnlyDictionary<string, FieldMetricDefinition> FieldMetricDefinitions =
        new Dictionary<string, FieldMetricDefinition>(StringComparer.Ordinal)
        {
            ["CUMULATIVE_LAYOUT_SHIFT_SCORE"] = new("cls", null, 0.01),
            ["EXPERIMENTAL_TIME_TO_FIRST_BYTE"] = new("ttfb", "ms", 1),
            ["FIRST_CONTENTFUL_PAINT_MS"] = new("fcp", "ms", 1),
            ["INTERACTION_TO_NEXT_PAINT"] = new("inp", "ms", 1),
            ["LARGEST_CONTENTFUL_PAINT_MS"] = new("lcp", "ms", 1),
        };

    private static readonly IReadOnlyDictionary<string, string> LabMetricNames =
        new Dictionary<string, string>(StringComparer.Ordinal)
        {
            ["first-contentful-paint"] = "fcp",
            ["largest-contentful-paint"] = "lcp",
            ["cumulative-layout-shift"] = "cls",
            ["total-blocking-time"] = "tbt",
            ["server-response-time"] = "serverResponseTime",
            ["speed-index"] = "speedIndex",
        };

    /// <summary>Parses one PSI response for the supplied request URL and strategy.</summary>
    internal static PageSpeedAnalysis Parse(string inputUrl, string strategy, PsiApiResponse? raw)
    {
        var lighthouse = raw?.LighthouseResult;
        return new PageSpeedAnalysis(
            new AnalysisMetadata(
                InputUrl: inputUrl,
                Strategy: strategy,
                AnalysisTimestamp: ParseTimestamp(raw?.AnalysisUtcTimestamp),
                FetchTime: ParseTimestamp(lighthouse?.FetchTime),
                LighthouseVersion: lighthouse?.LighthouseVersion,
                RequestedUrl: lighthouse?.RequestedUrl,
                FinalUrl: lighthouse?.FinalUrl,
                FinalDisplayedUrl: lighthouse?.FinalDisplayedUrl,
                MainDocumentUrl: lighthouse?.MainDocumentUrl,
                RunWarnings: NormalizeMessages(lighthouse?.RunWarnings),
                RuntimeError: lighthouse?.RuntimeError),
            ParseFieldData(raw?.LoadingExperience, raw?.OriginLoadingExperience),
            lighthouse is null ? null : ParseLabData(lighthouse));
    }

    private static DateTimeOffset? ParseTimestamp(string? value) =>
        DateTimeOffset.TryParse(
            value,
            CultureInfo.InvariantCulture,
            DateTimeStyles.AssumeUniversal | DateTimeStyles.AdjustToUniversal,
            out var parsed)
            ? parsed
            : null;

    private static FieldData? ParseFieldData(LoadingExperienceRaw? page, LoadingExperienceRaw? origin)
    {
        if (page is null && origin is null)
        {
            return null;
        }

        return new FieldData(ParseFieldExperience(page), ParseFieldExperience(origin));
    }

    private static FieldExperience? ParseFieldExperience(LoadingExperienceRaw? raw)
    {
        if (raw is null)
        {
            return null;
        }

        var metrics = new Dictionary<string, FieldMetric>(StringComparer.Ordinal);
        foreach (var (id, metric) in raw.Metrics ?? [])
        {
            var definition = FieldMetricDefinitions.TryGetValue(id, out var known)
                ? known
                : new FieldMetricDefinition(id.ToLowerInvariant(), null, 1);

            var distributions = (metric.Distributions ?? [])
                .Select(d => new FieldDistribution(
                    Scale(d.Min, definition.Scale),
                    Scale(d.Max, definition.Scale),
                    d.Proportion))
                .ToArray();

            metrics[definition.Name] = new FieldMetric(
                Id: id,
                Percentile: 75,
                Value: metric.Percentile * definition.Scale,
                Unit: definition.Unit,
                Rating: NormalizeRating(metric.Category),
                Distributions: distributions);
        }

        return new FieldExperience(
            Id: raw.Id ?? string.Empty,
            InitialUrl: raw.InitialUrl,
            OverallRating: NormalizeRating(raw.OverallCategory),
            OriginFallback: raw.OriginFallback,
            Metrics: metrics);
    }

    private static double? Scale(double? value, double scale) => value * scale;

    private static string NormalizeRating(string? category) => category?.ToUpperInvariant() switch
    {
        "FAST" => "good",
        "AVERAGE" => "needs-improvement",
        "SLOW" => "poor",
        _ => "unavailable",
    };

    private static LabData ParseLabData(LighthouseResultRaw raw)
    {
        var categories = new Dictionary<string, CategoryResult>(StringComparer.Ordinal);
        var insightIds = new HashSet<string>(StringComparer.Ordinal);

        foreach (var (key, category) in raw.Categories ?? [])
        {
            categories[key] = new CategoryResult(
                Id: category.Id ?? key,
                Title: category.Title ?? string.Empty,
                Description: category.Description,
                Score: category.Score,
                ScoreDisplayMode: NormalizeCategoryScoreDisplayMode(category.CategoryScoreDisplayMode));

            foreach (var auditRef in category.AuditRefs ?? [])
            {
                if (string.Equals(auditRef.Group, "insights", StringComparison.OrdinalIgnoreCase) &&
                    !string.IsNullOrEmpty(auditRef.Id))
                {
                    insightIds.Add(auditRef.Id);
                }
            }
        }

        var audits = raw.Audits ?? [];
        var metrics = new Dictionary<string, LabMetric>(StringComparer.Ordinal);
        foreach (var (id, friendlyName) in LabMetricNames)
        {
            if (!audits.TryGetValue(id, out var audit))
            {
                continue;
            }

            metrics[friendlyName] = new LabMetric(
                Id: id,
                Title: audit.Title ?? string.Empty,
                Description: audit.Description,
                Score: audit.Score,
                Value: audit.NumericValue,
                Unit: audit.NumericUnit,
                DisplayValue: audit.DisplayValue);
        }

        var insights = new List<LighthouseAudit>();
        var diagnostics = new List<LighthouseAudit>();
        var unscored = new List<LighthouseAudit>();
        var passed = new List<string>();
        var notApplicable = new List<string>();
        var manual = new List<string>();

        foreach (var (id, audit) in audits)
        {
            if (LabMetricNames.ContainsKey(id))
            {
                continue;
            }

            switch (audit.ScoreDisplayMode?.ToLowerInvariant())
            {
                case "notapplicable":
                    notApplicable.Add(id);
                    continue;
                case "manual":
                    manual.Add(id);
                    continue;
            }

            if (audit.Score >= 0.9)
            {
                passed.Add(id);
                continue;
            }

            var normalized = NormalizeAudit(id, audit);
            if (insightIds.Contains(id) || id.EndsWith("-insight", StringComparison.Ordinal))
            {
                insights.Add(normalized);
            }
            else if (audit.Score is null)
            {
                unscored.Add(normalized);
            }
            else
            {
                diagnostics.Add(normalized);
            }
        }

        insights.Sort((left, right) => string.CompareOrdinal(left.Id, right.Id));
        diagnostics.Sort((left, right) => string.CompareOrdinal(left.Id, right.Id));
        unscored.Sort((left, right) => string.CompareOrdinal(left.Id, right.Id));
        passed.Sort(StringComparer.Ordinal);
        notApplicable.Sort(StringComparer.Ordinal);
        manual.Sort(StringComparer.Ordinal);

        return new LabData(
            Categories: categories,
            Metrics: metrics,
            Insights: insights,
            Diagnostics: diagnostics,
            UnscoredAudits: unscored,
            PassedAuditIds: passed,
            NotApplicableAuditIds: notApplicable,
            ManualAuditIds: manual,
            Entities: raw.Entities ?? []);
    }

    private static string? NormalizeCategoryScoreDisplayMode(string? value)
    {
        const string prefix = "CATEGORY_SCORE_DISPLAY_MODE_";
        if (string.IsNullOrWhiteSpace(value))
        {
            return null;
        }

        return value.StartsWith(prefix, StringComparison.Ordinal)
            ? value[prefix.Length..].ToLowerInvariant()
            : value.ToLowerInvariant();
    }

    private static LighthouseAudit NormalizeAudit(string id, AuditRaw raw) =>
        new(
            Id: raw.Id ?? id,
            Title: raw.Title ?? string.Empty,
            Description: raw.Description,
            Score: raw.Score,
            ScoreDisplayMode: raw.ScoreDisplayMode,
            DisplayValue: raw.DisplayValue,
            Explanation: raw.Explanation,
            ErrorMessage: raw.ErrorMessage,
            Warnings: NormalizeMessages(raw.Warnings),
            NumericValue: raw.NumericValue,
            NumericUnit: raw.NumericUnit,
            MetricSavings: raw.MetricSavings,
            Details: raw.Details);

    private static IReadOnlyList<string> NormalizeMessages(JsonElement[]? values)
    {
        if (values is null || values.Length == 0)
        {
            return [];
        }

        return values
            .Select(value => value.ValueKind == JsonValueKind.String
                ? value.GetString() ?? string.Empty
                : value.GetRawText())
            .ToArray();
    }
}
