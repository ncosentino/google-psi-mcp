using System.Net;

namespace PageSpeedMcp.Infrastructure;

/// <summary>Shared transient retry behavior for Google API HTTP clients.</summary>
internal static class HttpClientRetryExtensions
{
    private const int MaxAttempts = 3;
    private static readonly TimeSpan BaseDelay = TimeSpan.FromMilliseconds(250);

    /// <summary>Sends a fresh request per attempt and retries network, 429, and 5xx failures.</summary>
    internal static async Task<HttpResponseMessage> SendWithRetryAsync(
        this HttpClient httpClient,
        Func<HttpRequestMessage> requestFactory,
        CancellationToken cancellationToken)
    {
        Exception? lastException = null;
        for (var attempt = 1; attempt <= MaxAttempts; attempt++)
        {
            using var request = requestFactory();
            HttpResponseMessage response;
            try
            {
                response = await httpClient
                    .SendAsync(request, HttpCompletionOption.ResponseHeadersRead, cancellationToken)
                    .ConfigureAwait(false);
            }
            catch (HttpRequestException ex) when (attempt < MaxAttempts)
            {
                lastException = ex;
                await Task.Delay(BaseDelay * attempt, cancellationToken).ConfigureAwait(false);
                continue;
            }
            catch (OperationCanceledException ex)
                when (!cancellationToken.IsCancellationRequested && attempt < MaxAttempts)
            {
                lastException = ex;
                await Task.Delay(BaseDelay * attempt, cancellationToken).ConfigureAwait(false);
                continue;
            }

            if (!IsRetryable(response.StatusCode) || attempt == MaxAttempts)
            {
                return response;
            }

            var delay = RetryDelay(response, attempt);
            response.Dispose();
            if (delay > TimeSpan.Zero)
            {
                await Task.Delay(delay, cancellationToken).ConfigureAwait(false);
            }
        }

        throw new HttpRequestException(
            $"HTTP request failed after {MaxAttempts} attempts.",
            lastException);
    }

    private static bool IsRetryable(HttpStatusCode statusCode) =>
        statusCode == HttpStatusCode.TooManyRequests || (int)statusCode >= 500;

    private static TimeSpan RetryDelay(HttpResponseMessage response, int attempt)
    {
        if (response.Headers.RetryAfter?.Delta is { } delta)
        {
            return delta;
        }
        if (response.Headers.RetryAfter?.Date is { } date)
        {
            var delay = date - DateTimeOffset.UtcNow;
            return delay > TimeSpan.Zero ? delay : TimeSpan.Zero;
        }
        return BaseDelay * attempt;
    }
}
