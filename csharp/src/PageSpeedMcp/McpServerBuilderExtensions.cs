using ModelContextProtocol.Server;

namespace PageSpeedMcp;

/// <summary>Shared MCP server builder configuration.</summary>
internal static class McpServerBuilderExtensions
{
    /// <summary>Registers array-argument coercion for every transport.</summary>
    internal static IMcpServerBuilder WithStringifiedArgsCoercion(
        this IMcpServerBuilder builder) =>
        builder.WithRequestFilters(filters =>
        {
            filters.AddCallToolFilter(next => async (context, cancellationToken) =>
            {
                if (context.Params is not null)
                {
                    StringifiedArgsCoercion.CoerceStringifiedArrayArgs(
                        context.Params,
                        StringifiedArgsCoercion.ToolArrayFields);
                }
                return await next(context, cancellationToken).ConfigureAwait(false);
            });
        });
}
