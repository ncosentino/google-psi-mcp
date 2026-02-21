using System.Runtime.CompilerServices;

namespace PageSpeedMcp.Config;

/// <summary>Resolves the Google PageSpeed Insights API key from multiple sources.</summary>
/// <remarks>Priority: CLI argument > GOOGLE_PSI_API_KEY env var > .env file.</remarks>
internal static class ApiKeyResolver
{
    private const string EnvVarName = "GOOGLE_PSI_API_KEY";
    private const string DotEnvFile = ".env";

    /// <summary>Returns the API key from the highest-priority available source.</summary>
    internal static string? Resolve(string? flagValue)
    {
        if (!string.IsNullOrWhiteSpace(flagValue))
            return flagValue;

        var envValue = Environment.GetEnvironmentVariable(EnvVarName);
        if (!string.IsNullOrWhiteSpace(envValue))
            return envValue;

        return ReadFromDotEnv();
    }

    [MethodImpl(MethodImplOptions.NoInlining)]
    private static string? ReadFromDotEnv()
    {
        if (!File.Exists(DotEnvFile))
            return null;

        foreach (var line in File.ReadLines(DotEnvFile))
        {
            var trimmed = line.Trim();
            if (trimmed.StartsWith('#') || trimmed.Length == 0)
                continue;

            var prefix = EnvVarName + "=";
            if (trimmed.StartsWith(prefix, StringComparison.Ordinal))
            {
                var value = trimmed[prefix.Length..].Trim('"', '\'');
                return string.IsNullOrWhiteSpace(value) ? null : value;
            }
        }

        return null;
    }
}
