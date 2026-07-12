using System.Globalization;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using ModelContextProtocol.Server;
using PageSpeedMcp;
using PageSpeedMcp.Config;
using PageSpeedMcp.Tools;

var apiKey = ApiKeyResolver.Resolve(
    args.SkipWhile(argument => argument != "--api-key").Skip(1).FirstOrDefault());

if (string.IsNullOrWhiteSpace(apiKey))
{
    await Console.Error.WriteLineAsync(
        "ERROR: No API key provided. Use --api-key <key>, set GOOGLE_PSI_API_KEY, or add it to a .env file.")
        .ConfigureAwait(false);
    return 1;
}

var transport = args
    .SkipWhile(argument => argument != "--transport")
    .Skip(1)
    .FirstOrDefault() ?? "stdio";

if (transport == "http")
{
    var portValue = Environment.GetEnvironmentVariable("PORT") is { Length: > 0 } configuredPort
        ? configuredPort
        : "8080";
    var app = Hosting.BuildHttpHost(
        args,
        apiKey,
        int.Parse(portValue, CultureInfo.InvariantCulture));
    await app.RunAsync().ConfigureAwait(false);
    return 0;
}
if (transport != "stdio")
{
    await Console.Error.WriteLineAsync(
        $"ERROR: Invalid transport \"{transport}\". Expected stdio or http.")
        .ConfigureAwait(false);
    return 1;
}

var builder = Host.CreateApplicationBuilder(args);
Hosting.ConfigureCommonServices(builder, apiKey);
builder.Services
    .AddMcpServer()
    .WithStringifiedArgsCoercion()
    .WithStdioServerTransport()
    .WithTools<PageSpeedTool>()
    .WithTools<CruxTool>();

var host = builder.Build();
await host.RunAsync().ConfigureAwait(false);
return 0;
