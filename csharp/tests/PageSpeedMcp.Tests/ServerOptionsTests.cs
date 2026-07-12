using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests shared-service command-line and environment configuration.</summary>
public sealed class ServerOptionsTests
{
    /// <summary>Verifies the HTTP service defaults to a loopback listener.</summary>
    [Fact]
    public void Parse_DefaultsToLoopback()
    {
        var options = ServerOptions.Parse([], _ => null);

        Assert.Equal("stdio", options.Transport);
        Assert.Equal(ServerOptions.DefaultListenAddress, options.ListenAddress);
        Assert.Equal(ServerOptions.DefaultPort, options.Port);
        Assert.Null(options.ShutdownToken);
    }

    /// <summary>Verifies explicit arguments override environment values.</summary>
    [Fact]
    public void Parse_ArgumentsOverrideEnvironment()
    {
        var environment = new Dictionary<string, string?>
        {
            ["MCP_LISTEN_ADDRESS"] = "192.0.2.1",
            ["PORT"] = "9000",
            ["MCP_SHUTDOWN_TOKEN"] = "test-token",
        };

        var options = ServerOptions.Parse(
            [
                "--transport=http",
                "--listen-address",
                "127.0.0.2",
                "--port",
                "8081",
            ],
            name => environment.GetValueOrDefault(name));

        Assert.Equal("http", options.Transport);
        Assert.Equal("127.0.0.2", options.ListenAddress);
        Assert.Equal(8081, options.Port);
        Assert.Equal("test-token", options.ShutdownToken);
    }

    /// <summary>Verifies invalid transports, missing values, and invalid ports are rejected.</summary>
    [Theory]
    [InlineData("--transport", "sse")]
    [InlineData("--listen-address", "")]
    [InlineData("--port", "0")]
    [InlineData("--port", "65536")]
    [InlineData("--port", "invalid")]
    public void Parse_InvalidConfiguration_IsRejected(string option, string value)
    {
        var args = option == "--transport"
            ? new[] { option, value }
            : new[] { "--transport", "http", option, value };
        Assert.Throws<ArgumentException>(() =>
            ServerOptions.Parse(args, _ => null));
    }
}
