using System.ComponentModel;
using System.Text.Json;
using ModelContextProtocol.Server;
using PageSpeedMcp.Crux;

namespace PageSpeedMcp.Tools;

/// <summary>MCP tools for current and historical Chrome UX Report data.</summary>
[McpServerToolType]
internal sealed class CruxTool(CruxClient client)
{
    [McpServerTool(Name = "get_crux_data")]
    [Description("Get current Chrome UX Report real-user data for a URL or origin. Supports all current CrUX metrics, including Core Web Vitals, LCP subparts, navigation types, RTT, resource types, and form-factor fractions. Requires the Chrome UX Report API to be enabled and allowed for the configured API key.")]
    internal async Task<string> GetCruxData(
        [Description("Absolute URL or origin to query.")] string target,
        [Description("Whether target is a url or origin. Defaults to url.")] string target_type = "url",
        [Description("Form factor: all, phone, tablet, or desktop. Defaults to all.")] string form_factor = "all",
        [Description("Optional CrUX metric names. Omit to request every available metric.")] string[]? metrics = null,
        CancellationToken cancellationToken = default)
    {
        var request = CruxQueryRequest.Create(
            target,
            target_type,
            form_factor,
            metrics,
            collectionPeriodCount: 0);
        var result = await client.QueryCurrentAsync(request, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(result, CruxJsonContext.Default.CruxResult);
    }

    [McpServerTool(Name = "get_crux_history")]
    [Description("Get up to 40 weekly Chrome UX Report collection periods for a URL or origin. Returns real-user metric timeseries with null values for unavailable periods. Requires the Chrome UX Report API to be enabled and allowed for the configured API key.")]
    internal async Task<string> GetCruxHistory(
        [Description("Absolute URL or origin to query.")] string target,
        [Description("Whether target is a url or origin. Defaults to url.")] string target_type = "url",
        [Description("Form factor: all, phone, tablet, or desktop. Defaults to all.")] string form_factor = "all",
        [Description("Optional CrUX metric names. Omit to request every available metric.")] string[]? metrics = null,
        [Description("Number of weekly collection periods to return, from 1 to 40. Defaults to 25.")] int collection_period_count = 25,
        CancellationToken cancellationToken = default)
    {
        var request = CruxQueryRequest.Create(
            target,
            target_type,
            form_factor,
            metrics,
            collection_period_count);
        var result = await client.QueryHistoryAsync(request, cancellationToken).ConfigureAwait(false);
        return JsonSerializer.Serialize(result, CruxJsonContext.Default.CruxHistoryResult);
    }
}
