using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using ModelContextProtocol.Server;
using PageSpeedMcp.Config;
using PageSpeedMcp.PageSpeed;
using PageSpeedMcp.Tools;

var apiKey = ApiKeyResolver.Resolve(
    args.SkipWhile(a => a != "--api-key").Skip(1).FirstOrDefault());

if (string.IsNullOrWhiteSpace(apiKey))
{
    await Console.Error.WriteLineAsync(
        "ERROR: No API key provided. Use --api-key <key>, set GOOGLE_PSI_API_KEY env var, or add it to a .env file.")
        .ConfigureAwait(false);
    return 1;
}

var builder = Host.CreateApplicationBuilder(args);

// All logs must go to stderr to avoid corrupting the MCP STDIO stream.
builder.Logging.AddConsole(o => o.LogToStandardErrorThreshold = LogLevel.Trace);
builder.Logging.SetMinimumLevel(LogLevel.Warning);

builder.Services
    .AddHttpClient(nameof(PageSpeedClient), http =>
    {
        http.Timeout = TimeSpan.FromSeconds(60);
    });

builder.Services.AddTransient<PageSpeedClient>(sp =>
{
    var factory = sp.GetRequiredService<IHttpClientFactory>();
    return new PageSpeedClient(factory.CreateClient(nameof(PageSpeedClient)), apiKey!);
});

builder.Services
    .AddMcpServer()
    .WithStdioServerTransport()
    .WithTools<PageSpeedTool>();

var host = builder.Build();
await host.RunAsync().ConfigureAwait(false);
return 0;