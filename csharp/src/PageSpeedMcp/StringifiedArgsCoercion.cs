using System.Text.Json;
using ModelContextProtocol.Protocol;

namespace PageSpeedMcp;

/// <summary>Repairs array arguments that an MCP client encoded as JSON strings.</summary>
internal static class StringifiedArgsCoercion
{
    internal static readonly IReadOnlyDictionary<string, string[]> ToolArrayFields =
        new Dictionary<string, string[]>
        {
            ["analyze_page"] = ["categories"],
            ["analyze_pages"] = ["urls", "categories"],
            ["get_crux_data"] = ["metrics"],
            ["get_crux_history"] = ["metrics"],
        };

    /// <summary>Coerces declared stringified arrays before tool parameter binding.</summary>
    internal static void CoerceStringifiedArrayArgs(
        CallToolRequestParams requestParams,
        IReadOnlyDictionary<string, string[]> arrayFieldsByTool)
    {
        if (requestParams.Arguments is null ||
            !arrayFieldsByTool.TryGetValue(requestParams.Name, out var fields))
        {
            return;
        }

        foreach (var field in fields)
        {
            if (requestParams.Arguments.TryGetValue(field, out var value) &&
                TryCoerceStringifiedArray(value, out var coerced))
            {
                requestParams.Arguments[field] = coerced;
            }
        }
    }

    private static bool TryCoerceStringifiedArray(
        JsonElement value,
        out JsonElement coerced)
    {
        coerced = default;
        if (value.ValueKind != JsonValueKind.String ||
            string.IsNullOrEmpty(value.GetString()))
        {
            return false;
        }

        try
        {
            using var document = JsonDocument.Parse(value.GetString()!);
            if (document.RootElement.ValueKind != JsonValueKind.Array)
            {
                return false;
            }
            coerced = document.RootElement.Clone();
            return true;
        }
        catch (JsonException)
        {
            return false;
        }
    }
}
