namespace PageSpeedMcp.PageSpeed;

/// <summary>A validated PageSpeed Insights API request.</summary>
internal sealed record PageSpeedAnalysisRequest(
    string Url,
    string Strategy,
    IReadOnlyList<string> Categories,
    string? Locale)
{
    private static readonly HashSet<string> ValidCategories =
    [
        "performance",
        "seo",
        "accessibility",
        "best-practices",
        "agentic-browsing",
    ];

    /// <summary>Creates a validated request from MCP tool input.</summary>
    internal static PageSpeedAnalysisRequest Create(
        string url,
        string strategy,
        IReadOnlyList<string>? categories,
        string? locale)
    {
        url = url.Trim();
        if (!Uri.TryCreate(url, UriKind.Absolute, out var parsedUrl) ||
            (parsedUrl.Scheme != Uri.UriSchemeHttp && parsedUrl.Scheme != Uri.UriSchemeHttps))
        {
            throw new ArgumentException("url must be an absolute HTTP or HTTPS URL", nameof(url));
        }

        strategy = strategy.Trim().ToLowerInvariant();
        if (strategy is not ("mobile" or "desktop"))
        {
            throw new ArgumentException("strategy must be mobile or desktop", nameof(strategy));
        }

        var normalizedCategories = NormalizeCategories(categories);
        return new PageSpeedAnalysisRequest(
            parsedUrl.AbsoluteUri,
            strategy,
            normalizedCategories,
            string.IsNullOrWhiteSpace(locale) ? null : locale.Trim());
    }

    /// <summary>Validates a tool strategy and expands the both selection.</summary>
    internal static IReadOnlyList<string> ResolveStrategies(string? strategy) =>
        strategy?.Trim().ToLowerInvariant() switch
        {
            null or "" or "both" => ["mobile", "desktop"],
            "mobile" => ["mobile"],
            "desktop" => ["desktop"],
            _ => throw new ArgumentException(
                "strategy must be mobile, desktop, or both",
                nameof(strategy)),
        };

    private static IReadOnlyList<string> NormalizeCategories(IReadOnlyList<string>? categories)
    {
        if (categories is null || categories.Count == 0)
        {
            return ["performance", "seo", "accessibility", "best-practices"];
        }

        var normalized = new List<string>(categories.Count);
        var seen = new HashSet<string>(StringComparer.Ordinal);
        foreach (var suppliedCategory in categories)
        {
            var category = suppliedCategory.Trim().ToLowerInvariant();
            if (!ValidCategories.Contains(category))
            {
                throw new ArgumentException(
                    $"category \"{category}\" is invalid: must be performance, seo, accessibility, best-practices, or agentic-browsing",
                    nameof(categories));
            }
            if (seen.Add(category))
            {
                normalized.Add(category);
            }
        }
        return normalized;
    }
}
