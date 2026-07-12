using System.Net;
using System.Text;
using Microsoft.AspNetCore.Builder;
using ModelContextProtocol.Client;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests the real Streamable HTTP host and MCP request pipeline.</summary>
public sealed class HostingHttpTests
{
    private sealed class ResponseHandler(string responseBody) : HttpMessageHandler
    {
        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken) =>
            Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(responseBody, Encoding.UTF8, "application/json"),
            });
    }

    /// <summary>Verifies all four tools are available over Streamable HTTP.</summary>
    [Fact]
    public async Task BuildHttpHost_ServesAllTools()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            await using var client = await ConnectAsync(app);
            var tools = await client.ListToolsAsync();

            Assert.Equal(4, tools.Count);
            Assert.Contains(tools, tool => tool.Name == "analyze_page");
            Assert.Contains(tools, tool => tool.Name == "analyze_pages");
            Assert.Contains(tools, tool => tool.Name == "get_crux_data");
            Assert.Contains(tools, tool => tool.Name == "get_crux_history");
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies optional PSI inputs work through real MCP binding and dispatch.</summary>
    [Fact]
    public async Task BuildHttpHost_AnalyzePageWithDefaults_ReturnsSuccess()
    {
        var handler = new ResponseHandler(
            """{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}""");
        await using var app = Hosting.BuildHttpHost(
            [],
            "test-key",
            port: 0,
            pageSpeedHandler: handler);
        await app.StartAsync();
        try
        {
            await using var client = await ConnectAsync(app);
            var result = await client.CallToolAsync(
                "analyze_page",
                new Dictionary<string, object?>
                {
                    ["url"] = "https://example.test",
                });

            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies stringified arrays are repaired before C# parameter binding.</summary>
    [Fact]
    public async Task BuildHttpHost_StringifiedCategories_ReturnsSuccess()
    {
        var handler = new ResponseHandler(
            """{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}""");
        await using var app = Hosting.BuildHttpHost(
            [],
            "test-key",
            port: 0,
            pageSpeedHandler: handler);
        await app.StartAsync();
        try
        {
            await using var client = await ConnectAsync(app);
            var result = await client.CallToolAsync(
                "analyze_page",
                new Dictionary<string, object?>
                {
                    ["url"] = "https://example.test",
                    ["strategy"] = "mobile",
                    ["categories"] = """["performance"]""",
                });

            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies CrUX tools resolve their dedicated client through the HTTP host.</summary>
    [Fact]
    public async Task BuildHttpHost_GetCruxData_ReturnsSuccess()
    {
        var handler = new ResponseHandler(ReadFixture("crux-current.json"));
        await using var app = Hosting.BuildHttpHost(
            [],
            "test-key",
            port: 0,
            cruxHandler: handler);
        await app.StartAsync();
        try
        {
            await using var client = await ConnectAsync(app);
            var result = await client.CallToolAsync(
                "get_crux_data",
                new Dictionary<string, object?>
                {
                    ["target"] = "https://example.test/page",
                    ["form_factor"] = "phone",
                });

            Assert.NotEqual(true, result.IsError);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies HTTP host filtering defaults to loopback hosts.</summary>
    [Fact]
    public void BuildHttpHost_DefaultsAllowedHostsToLoopback()
    {
        using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        Assert.Equal(Hosting.DefaultAllowedHosts, app.Configuration["AllowedHosts"]);
    }

    private static async Task<McpClient> ConnectAsync(WebApplication app)
    {
        var bound = new Uri(app.Urls.First());
        var endpoint = new UriBuilder(bound) { Host = "127.0.0.1" }.Uri;
        return await McpClient.CreateAsync(new HttpClientTransport(
            new HttpClientTransportOptions { Endpoint = endpoint }));
    }

    private static string ReadFixture(string name)
    {
        var path = Path.Combine(AppContext.BaseDirectory, "testdata", name);
        return File.ReadAllText(path);
    }
}
