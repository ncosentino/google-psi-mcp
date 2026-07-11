using Microsoft.AspNetCore.Builder;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using ModelContextProtocol.Server;
using PageSpeedMcp.Crux;
using PageSpeedMcp.PageSpeed;
using PageSpeedMcp.Tools;

namespace PageSpeedMcp;

/// <summary>Builds the MCP server hosts for STDIO and Streamable HTTP.</summary>
internal static class Hosting
{
    /// <summary>Default Host header allow-list for HTTP transport.</summary>
    internal const string DefaultAllowedHosts = "localhost;127.0.0.1;[::1]";

    /// <summary>Builds an HTTP host without starting it.</summary>
    internal static WebApplication BuildHttpHost(
        string[] args,
        string apiKey,
        int port,
        HttpMessageHandler? pageSpeedHandler = null,
        HttpMessageHandler? cruxHandler = null)
    {
        var builder = WebApplication.CreateBuilder(args);
        if (string.IsNullOrWhiteSpace(builder.Configuration["AllowedHosts"]))
        {
            builder.Configuration["AllowedHosts"] = DefaultAllowedHosts;
        }
        builder.WebHost.UseUrls($"http://0.0.0.0:{port}");

        ConfigureCommonServices(builder, apiKey, pageSpeedHandler, cruxHandler);
        builder.Services
            .AddMcpServer()
            .WithStringifiedArgsCoercion()
            .WithHttpTransport(options => options.Stateless = true)
            .WithTools<PageSpeedTool>()
            .WithTools<CruxTool>();

        var app = builder.Build();
        app.MapMcp();
        return app;
    }

    /// <summary>Registers services shared by both transports.</summary>
    internal static void ConfigureCommonServices(
        IHostApplicationBuilder builder,
        string apiKey,
        HttpMessageHandler? pageSpeedHandler = null,
        HttpMessageHandler? cruxHandler = null)
    {
        builder.Logging.AddConsole(options =>
            options.LogToStandardErrorThreshold = LogLevel.Trace);
        builder.Logging.SetMinimumLevel(LogLevel.Warning);

        var pageSpeedClient = builder.Services.AddHttpClient(
            nameof(PageSpeedClient),
            http => http.Timeout = TimeSpan.FromSeconds(120));
        if (pageSpeedHandler is not null)
        {
            pageSpeedClient.ConfigurePrimaryHttpMessageHandler(() => pageSpeedHandler);
        }

        var cruxClient = builder.Services.AddHttpClient(
            nameof(CruxClient),
            http => http.Timeout = TimeSpan.FromSeconds(30));
        if (cruxHandler is not null)
        {
            cruxClient.ConfigurePrimaryHttpMessageHandler(() => cruxHandler);
        }

        builder.Services.AddTransient<PageSpeedClient>(services =>
        {
            var factory = services.GetRequiredService<IHttpClientFactory>();
            return new PageSpeedClient(factory.CreateClient(nameof(PageSpeedClient)), apiKey);
        });
        builder.Services.AddTransient<CruxClient>(services =>
        {
            var factory = services.GetRequiredService<IHttpClientFactory>();
            return new CruxClient(factory.CreateClient(nameof(CruxClient)), apiKey);
        });
    }
}
