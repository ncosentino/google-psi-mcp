using System.Net;
using System.Text;
using System.Text.Json;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using ModelContextProtocol.Client;
using PageSpeedMcp.Infrastructure;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests the real Streamable HTTP host and MCP request pipeline.</summary>
public sealed class HostingHttpTests
{
    private sealed class ResponseHandler(string responseBody) : HttpMessageHandler
    {
        private int _active;
        private int _maxActive;

        internal int MaxActive => _maxActive;

        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken) =>
            SendTrackedAsync(cancellationToken);

        private async Task<HttpResponseMessage> SendTrackedAsync(
            CancellationToken cancellationToken)
        {
            var active = Interlocked.Increment(ref _active);
            while (true)
            {
                var currentMax = Volatile.Read(ref _maxActive);
                if (active <= currentMax ||
                    Interlocked.CompareExchange(ref _maxActive, active, currentMax) == currentMax)
                {
                    break;
                }
            }
            try
            {
                await Task.Delay(TimeSpan.FromMilliseconds(20), cancellationToken);
                return new HttpResponseMessage(HttpStatusCode.OK)
                {
                    Content = new StringContent(responseBody, Encoding.UTF8, "application/json"),
                };
            }
            finally
            {
                Interlocked.Decrement(ref _active);
            }
        }
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

    /// <summary>Verifies the shared service exposes machine-readable health metadata.</summary>
    [Fact]
    public async Task BuildHttpHost_ServesHealth()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            using var httpClient = new HttpClient();
            using var response = await httpClient.GetAsync(GetEndpoint(app, Hosting.HealthPath));
            response.EnsureSuccessStatusCode();

            using var document = JsonDocument.Parse(await response.Content.ReadAsStringAsync());
            Assert.Equal("ok", document.RootElement.GetProperty("status").GetString());
            Assert.Equal(
                "google-psi-mcp",
                document.RootElement.GetProperty("service").GetString());
            Assert.Equal("http", document.RootElement.GetProperty("transport").GetString());
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies cross-site browser requests cannot reach the MCP endpoint.</summary>
    [Fact]
    public async Task BuildHttpHost_RejectsCrossSiteOrigin()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            using var httpClient = new HttpClient();
            using var request = new HttpRequestMessage(
                HttpMethod.Post,
                GetEndpoint(app, Hosting.McpPath));
            request.Headers.Add("Origin", "https://evil.example");
            request.Content = new StringContent("{}", Encoding.UTF8, "application/json");

            using var response = await httpClient.SendAsync(request);

            Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies the trailing-slash MCP route has the same Origin protection.</summary>
    [Fact]
    public async Task BuildHttpHost_RejectsCrossSiteOriginWithTrailingSlash()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            using var httpClient = new HttpClient();
            using var request = new HttpRequestMessage(
                HttpMethod.Post,
                GetEndpoint(app, Hosting.McpPath + "/"));
            request.Headers.Add("Origin", "https://evil.example");
            request.Content = new StringContent("{}", Encoding.UTF8, "application/json");

            using var response = await httpClient.SendAsync(request);

            Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies a browser request from the MCP endpoint's own origin is accepted.</summary>
    [Fact]
    public void IsAllowedOrigin_AcceptsSameOrigin()
    {
        var context = new DefaultHttpContext();
        context.Request.Scheme = "http";
        context.Request.Host = new HostString("127.0.0.1", 8080);

        Assert.True(Hosting.IsAllowedOrigin(
            context.Request,
            "http://127.0.0.1:8080"));
    }

    /// <summary>Verifies Host filtering rejects a forged hostname before MCP dispatch.</summary>
    [Fact]
    public async Task BuildHttpHost_RejectsDisallowedHost()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            using var httpClient = new HttpClient();
            using var request = new HttpRequestMessage(
                HttpMethod.Get,
                GetEndpoint(app, Hosting.HealthPath));
            request.Headers.Host = "evil.example";

            using var response = await httpClient.SendAsync(request);

            Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    /// <summary>Verifies local shutdown requires the per-run bearer token.</summary>
    [Fact]
    public async Task BuildHttpHost_ShutdownRequiresBearerToken()
    {
        await using var app = Hosting.BuildHttpHost(
            [],
            "test-key",
            port: 0,
            shutdownToken: "secret-token");
        await app.StartAsync();

        using var httpClient = new HttpClient();
        using var rejectedRequest = new HttpRequestMessage(
            HttpMethod.Post,
            GetEndpoint(app, Hosting.ShutdownPath));
        rejectedRequest.Headers.Authorization =
            new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", "wrong-token");
        using var rejectedResponse = await httpClient.SendAsync(rejectedRequest);
        Assert.Equal(HttpStatusCode.Unauthorized, rejectedResponse.StatusCode);

        var stopped = new TaskCompletionSource<bool>(
            TaskCreationOptions.RunContinuationsAsynchronously);
        app.Lifetime.ApplicationStopping.Register(() => stopped.SetResult(true));
        using var acceptedRequest = new HttpRequestMessage(
            HttpMethod.Post,
            GetEndpoint(app, Hosting.ShutdownPath));
        acceptedRequest.Headers.Authorization =
            new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", "secret-token");
        using var acceptedResponse = await httpClient.SendAsync(acceptedRequest);

        Assert.Equal(HttpStatusCode.Accepted, acceptedResponse.StatusCode);
        await stopped.Task.WaitAsync(TimeSpan.FromSeconds(5));
    }

    /// <summary>Verifies all clients share one process-wide PSI concurrency limit.</summary>
    [Fact]
    public async Task BuildHttpHost_SharesConcurrencyLimitAcrossClients()
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
            var clients = await Task.WhenAll(
                Enumerable.Range(0, 8).Select(_ => ConnectAsync(app)));
            try
            {
                var calls = clients.Select(async (client, index) =>
                    await client.CallToolAsync(
                        "analyze_page",
                        new Dictionary<string, object?>
                        {
                            ["url"] = $"https://example.test/{index}",
                            ["strategy"] = "mobile",
                        }));
                var results = await Task.WhenAll(calls);

                Assert.All(results, result => Assert.NotEqual(true, result.IsError));
                Assert.InRange(
                    handler.MaxActive,
                    2,
                    PageSpeedRequestLimiter.DefaultMaxConcurrency);
            }
            finally
            {
                foreach (var client in clients)
                {
                    await client.DisposeAsync();
                }
            }
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

    /// <summary>Verifies the HTTP host binds loopback unless explicitly configured otherwise.</summary>
    [Fact]
    public async Task BuildHttpHost_DefaultsToLoopbackBinding()
    {
        await using var app = Hosting.BuildHttpHost([], "test-key", port: 0);
        await app.StartAsync();
        try
        {
            var bound = new Uri(app.Urls.First());
            Assert.Equal(ServerOptions.DefaultListenAddress, bound.Host);
        }
        finally
        {
            await app.StopAsync();
        }
    }

    private static async Task<McpClient> ConnectAsync(WebApplication app)
    {
        return await McpClient.CreateAsync(new HttpClientTransport(
            new HttpClientTransportOptions
            {
                Endpoint = GetEndpoint(app, Hosting.McpPath),
            }));
    }

    private static Uri GetEndpoint(WebApplication app, string path)
    {
        var bound = new Uri(app.Urls.First());
        return new UriBuilder(bound)
        {
            Host = ServerOptions.DefaultListenAddress,
            Path = path,
        }.Uri;
    }

    private static string ReadFixture(string name)
    {
        var path = Path.Combine(AppContext.BaseDirectory, "testdata", name);
        return File.ReadAllText(path);
    }
}
