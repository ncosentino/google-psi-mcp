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
}
