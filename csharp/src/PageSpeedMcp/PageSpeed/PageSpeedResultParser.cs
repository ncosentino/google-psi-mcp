namespace PageSpeedMcp.PageSpeed;

/// <summary>Parses a raw PSI API response into a strongly-typed <see cref="PageSpeedResult"/>.</summary>
internal static class PageSpeedResultParser
{
    internal static PageSpeedResult Parse(string url, string strategy, PsiApiResponse? raw)
    {
        var lhr = raw?.LighthouseResult;

        var scores = ParseScores(lhr?.Categories);
        var cwv = ParseCoreWebVitals(lhr?.Audits);
        var (opportunities, failing, passed) = ParseAudits(lhr?.Audits);

        return new PageSpeedResult(
            Url: url,
            Strategy: strategy,
            AnalyzedAt: DateTimeOffset.UtcNow,
            Scores: scores,
            CoreWebVitals: cwv,
            Opportunities: opportunities,
            FailingAudits: failing,
            PassedAuditIds: passed);
    }

    private static CategoryScores ParseScores(Dictionary<string, CategoryRaw>? cats)
    {
        static int ToScore(Dictionary<string, CategoryRaw>? d, string key) =>
            d is not null && d.TryGetValue(key, out var c) ? (int)Math.Round(c.Score * 100) : 0;

        return new CategoryScores(
            Performance: ToScore(cats, "performance"),
            Seo: ToScore(cats, "seo"),
            Accessibility: ToScore(cats, "accessibility"),
            BestPractices: ToScore(cats, "best-practices"));
    }

    private static CoreWebVitals ParseCoreWebVitals(Dictionary<string, AuditRaw>? audits)
    {
        return new CoreWebVitals(
            Fcp: ParseTimeMetric(audits, "first-contentful-paint", FcpRating),
            Lcp: ParseTimeMetric(audits, "largest-contentful-paint", LcpRating),
            Cls: ParseClsMetric(audits),
            Tbt: ParseTimeMetric(audits, "total-blocking-time", TbtRating),
            Ttfb: ParseTimeMetric(audits, "server-response-time", TtfbRating),
            SpeedIndex: ParseTimeMetric(audits, "speed-index", SiRating));
    }

    private static MetricValue ParseTimeMetric(
        Dictionary<string, AuditRaw>? audits,
        string id,
        Func<double, string> ratingFn)
    {
        if (audits is null || !audits.TryGetValue(id, out var a) || a.NumericValue is null)
            return new MetricValue(0, null, string.Empty);

        var secs = Math.Round(a.NumericValue.Value / 1000, 2);
        return new MetricValue(secs, "s", ratingFn(secs));
    }

    private static MetricValue ParseClsMetric(Dictionary<string, AuditRaw>? audits)
    {
        if (audits is null || !audits.TryGetValue("cumulative-layout-shift", out var a) || a.NumericValue is null)
            return new MetricValue(0, null, string.Empty);

        var v = Math.Round(a.NumericValue.Value, 2);
        return new MetricValue(v, null, ClsRating(v));
    }

    private static (
        IReadOnlyList<Opportunity> Opportunities,
        IReadOnlyList<AuditResult> Failing,
        IReadOnlyList<string> Passed) ParseAudits(Dictionary<string, AuditRaw>? audits)
    {
        if (audits is null)
            return ([], [], []);

        var opportunities = new List<Opportunity>();
        var failing = new List<AuditResult>();
        var passed = new List<string>();

        foreach (var (id, a) in audits)
        {
            if (a.Score is null)
                continue;

            if (a.Score >= 0.9)
            {
                passed.Add(id);
                continue;
            }

            if (a.Details?.Type == "opportunity")
            {
                opportunities.Add(new Opportunity(id, a.Title, a.Description, a.DisplayValue, ImpactFromScore(a.Score.Value)));
            }
            else
            {
                failing.Add(new AuditResult(id, a.Title, a.Description, a.Score.Value, a.DisplayValue));
            }
        }

        return (opportunities, failing, passed);
    }

    private static string ImpactFromScore(double score) => score switch
    {
        < 0.5 => "high",
        < 0.75 => "medium",
        _ => "low",
    };

    // CWV rating functions follow Google's standard thresholds.
    private static string FcpRating(double s) => s switch { < 1.8 => "good", < 3.0 => "needs-improvement", _ => "poor" };
    private static string LcpRating(double s) => s switch { < 2.5 => "good", < 4.0 => "needs-improvement", _ => "poor" };
    private static string ClsRating(double v) => v switch { < 0.1 => "good", < 0.25 => "needs-improvement", _ => "poor" };
    private static string TbtRating(double s) => s switch { < 0.2 => "good", < 0.6 => "needs-improvement", _ => "poor" };
    private static string TtfbRating(double s) => s switch { < 0.8 => "good", < 1.8 => "needs-improvement", _ => "poor" };
    private static string SiRating(double s) => s switch { < 3.4 => "good", < 5.8 => "needs-improvement", _ => "poor" };
}
