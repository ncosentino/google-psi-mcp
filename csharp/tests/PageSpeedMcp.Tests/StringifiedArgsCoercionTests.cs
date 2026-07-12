using System.Text.Json;
using ModelContextProtocol.Protocol;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests defensive coercion of stringified array arguments.</summary>
public sealed class StringifiedArgsCoercionTests
{
    /// <summary>Verifies a JSON-encoded array string becomes a genuine array.</summary>
    [Fact]
    public void CoerceStringifiedArrayArgs_StringifiedArray_IsReplaced()
    {
        var request = RequestFor(
            "analyze_page",
            "categories",
            JsonOf("\"[\\\"performance\\\"]\""));

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(
            request,
            StringifiedArgsCoercion.ToolArrayFields);

        Assert.Equal(JsonValueKind.Array, request.Arguments!["categories"].ValueKind);
        Assert.Equal("performance", request.Arguments["categories"][0].GetString());
    }

    /// <summary>Verifies genuine arrays remain unchanged.</summary>
    [Fact]
    public void CoerceStringifiedArrayArgs_GenuineArray_IsUnchanged()
    {
        var request = RequestFor(
            "get_crux_data",
            "metrics",
            JsonOf("[\"largest_contentful_paint\"]"));

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(
            request,
            StringifiedArgsCoercion.ToolArrayFields);

        Assert.Equal(JsonValueKind.Array, request.Arguments!["metrics"].ValueKind);
    }

    /// <summary>Verifies invalid JSON strings remain available for normal validation.</summary>
    [Fact]
    public void CoerceStringifiedArrayArgs_InvalidString_IsUnchanged()
    {
        var request = RequestFor(
            "analyze_pages",
            "urls",
            JsonOf("\"not-json\""));

        StringifiedArgsCoercion.CoerceStringifiedArrayArgs(
            request,
            StringifiedArgsCoercion.ToolArrayFields);

        Assert.Equal(JsonValueKind.String, request.Arguments!["urls"].ValueKind);
    }

    private static CallToolRequestParams RequestFor(
        string tool,
        string field,
        JsonElement value) =>
        new()
        {
            Name = tool,
            Arguments = new Dictionary<string, JsonElement> { [field] = value },
        };

    private static JsonElement JsonOf(string json)
    {
        using var document = JsonDocument.Parse(json);
        return document.RootElement.Clone();
    }
}
