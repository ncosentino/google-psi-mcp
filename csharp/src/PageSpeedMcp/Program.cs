using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using ModelContextProtocol.Server;
using PageSpeedMcp;
using PageSpeedMcp.Config;
using PageSpeedMcp.Tools;

if (ServerOptions.IsHelpRequested(args))
{
    await Console.Out.WriteLineAsync(ServerOptions.Usage).ConfigureAwait(false);
    return 0;
}

ServerOptions options;
try
{
    options = ServerOptions.Parse(args);
}
catch (ArgumentException exception)
{
    await Console.Error.WriteLineAsync($"ERROR: {exception.Message}").ConfigureAwait(false);
    return 1;
}

var apiKey = ApiKeyResolver.Resolve(
    args.SkipWhile(argument => argument != "--api-key").Skip(1).FirstOrDefault());

if (string.IsNullOrWhiteSpace(apiKey))
{
    await Console.Error.WriteLineAsync(
        "ERROR: No API key provided. Use --api-key <key>, set GOOGLE_PSI_API_KEY, or add it to a .env file.")
        .ConfigureAwait(false);
    return 1;
}

if (options.Transport == "http")
{
    var app = Hosting.BuildHttpHost(
        args,
        apiKey,
        options.Port,
        options.ListenAddress,
        options.ShutdownToken);
    await app.RunAsync().ConfigureAwait(false);
    return 0;
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
