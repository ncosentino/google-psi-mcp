using System.Net;
using System.Net.Http.Headers;
using PageSpeedMcp.Infrastructure;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests shared Google API retry behavior.</summary>
public sealed class HttpClientRetryExtensionsTests
{
    private sealed class TransientHandler : HttpMessageHandler
    {
        private int _requests;

        internal int Requests => _requests;

        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            var requestNumber = Interlocked.Increment(ref _requests);
            var response = new HttpResponseMessage(
                requestNumber == 1
                    ? HttpStatusCode.ServiceUnavailable
                    : HttpStatusCode.OK);
            if (requestNumber == 1)
            {
                response.Headers.RetryAfter = new RetryConditionHeaderValue(TimeSpan.Zero);
            }
            return Task.FromResult(response);
        }
    }

    private sealed class TimeoutHandler : HttpMessageHandler
    {
        private int _requests;

        internal int Requests => _requests;

        protected override Task<HttpResponseMessage> SendAsync(
            HttpRequestMessage request,
            CancellationToken cancellationToken)
        {
            if (Interlocked.Increment(ref _requests) == 1)
            {
                throw new TaskCanceledException(
                    "The request timed out.",
                    new TimeoutException());
            }
            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK));
        }
    }

    /// <summary>Verifies transient HTTP statuses are retried with fresh requests.</summary>
    [Fact]
    public async Task SendWithRetryAsync_TransientStatus_IsRetried()
    {
        var handler = new TransientHandler();
        using var httpClient = new HttpClient(handler);

        using var response = await httpClient.SendWithRetryAsync(
            () => new HttpRequestMessage(HttpMethod.Get, "https://example.test"),
            CancellationToken.None);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(2, handler.Requests);
    }

    /// <summary>Verifies client-side timeouts are retried without swallowing caller cancellation.</summary>
    [Fact]
    public async Task SendWithRetryAsync_ClientTimeout_IsRetried()
    {
        var handler = new TimeoutHandler();
        using var httpClient = new HttpClient(handler);

        using var response = await httpClient.SendWithRetryAsync(
            () => new HttpRequestMessage(HttpMethod.Get, "https://example.test"),
            CancellationToken.None);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.Equal(2, handler.Requests);
    }
}
