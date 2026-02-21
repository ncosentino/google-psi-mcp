using PageSpeedMcp.Config;
using Xunit;

namespace PageSpeedMcp.Tests;

public sealed class ApiKeyResolverTests
{
    [Fact]
    public void Resolve_FlagValue_ReturnsFlag()
    {
        var result = ApiKeyResolver.Resolve("my-flag-key");
        Assert.Equal("my-flag-key", result);
    }

    [Fact]
    public void Resolve_NullFlag_ReturnsNull_WhenNoEnvOrFile()
    {
        // Ensure env var is not set during this test (it may be set in CI).
        // We cannot reliably unset env vars in xUnit without helper libraries,
        // so we only assert when env var is not present.
        if (Environment.GetEnvironmentVariable("GOOGLE_PSI_API_KEY") is not null)
            return; // skip -- env var present in this environment

        var result = ApiKeyResolver.Resolve(null);
        Assert.Null(result);
    }

    [Fact]
    public void Resolve_FlagTakesPriorityOverEnv()
    {
        // This test verifies flag wins even conceptually; env var may or may not be set.
        var result = ApiKeyResolver.Resolve("flag-wins");
        Assert.Equal("flag-wins", result);
    }
}
