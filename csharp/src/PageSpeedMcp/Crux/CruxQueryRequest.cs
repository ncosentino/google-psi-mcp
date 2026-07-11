namespace PageSpeedMcp.Crux;

/// <summary>A validated current or historical Chrome UX Report request.</summary>
internal sealed record CruxQueryRequest(
    string Target,
    string TargetType,
    string FormFactor,
    IReadOnlyList<string> Metrics,
    int CollectionPeriodCount)
{
    /// <summary>Creates a validated request from MCP tool input.</summary>
    internal static CruxQueryRequest Create(
        string target,
        string? targetType,
        string? formFactor,
        IReadOnlyList<string>? metrics,
        int collectionPeriodCount)
    {
        target = target.Trim();
        if (!Uri.TryCreate(target, UriKind.Absolute, out var parsedTarget) ||
            (parsedTarget.Scheme != Uri.UriSchemeHttp &&
             parsedTarget.Scheme != Uri.UriSchemeHttps))
        {
            throw new ArgumentException(
                "target must be an absolute HTTP or HTTPS URL",
                nameof(target));
        }

        targetType = string.IsNullOrWhiteSpace(targetType)
            ? "url"
            : targetType.Trim().ToLowerInvariant();
        if (targetType is not ("url" or "origin"))
        {
            throw new ArgumentException("target_type must be url or origin", nameof(targetType));
        }
        if (targetType == "origin")
        {
            if (parsedTarget.AbsolutePath is not ("" or "/") ||
                !string.IsNullOrEmpty(parsedTarget.Query) ||
                !string.IsNullOrEmpty(parsedTarget.Fragment))
            {
                throw new ArgumentException(
                    "origin targets must not include a path, query, or fragment",
                    nameof(target));
            }
            target = parsedTarget.GetLeftPart(UriPartial.Authority);
        }
        else
        {
            target = parsedTarget.AbsoluteUri;
        }

        formFactor = string.IsNullOrWhiteSpace(formFactor)
            ? "all"
            : formFactor.Trim().ToLowerInvariant();
        if (formFactor is not ("all" or "phone" or "tablet" or "desktop"))
        {
            throw new ArgumentException(
                "form_factor must be all, phone, tablet, or desktop",
                nameof(formFactor));
        }

        var normalizedMetrics = new List<string>(metrics?.Count ?? 0);
        var seen = new HashSet<string>(StringComparer.Ordinal);
        foreach (var suppliedMetric in metrics ?? [])
        {
            var metric = suppliedMetric.Trim().ToLowerInvariant();
            if (!IsMetricName(metric))
            {
                throw new ArgumentException(
                    $"metric \"{metric}\" is not a valid CrUX metric name",
                    nameof(metrics));
            }
            if (seen.Add(metric))
            {
                normalizedMetrics.Add(metric);
            }
        }

        if (collectionPeriodCount is < 0 or > 40)
        {
            throw new ArgumentOutOfRangeException(
                nameof(collectionPeriodCount),
                "collection_period_count must be between 1 and 40");
        }

        return new CruxQueryRequest(
            target,
            targetType,
            formFactor,
            normalizedMetrics,
            collectionPeriodCount);
    }

    private static bool IsMetricName(string metric)
    {
        if (metric.Length == 0 || metric[0] is < 'a' or > 'z')
        {
            return false;
        }
        return metric.All(character =>
            character is >= 'a' and <= 'z' ||
            character is >= '0' and <= '9' ||
            character == '_');
    }
}
