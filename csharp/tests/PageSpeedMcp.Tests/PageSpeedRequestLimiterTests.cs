using PageSpeedMcp.Infrastructure;
using Xunit;

namespace PageSpeedMcp.Tests;

/// <summary>Tests process-wide PageSpeed API concurrency control.</summary>
public sealed class PageSpeedRequestLimiterTests
{
    /// <summary>Verifies concurrent callers cannot exceed the configured limit.</summary>
    [Fact]
    public async Task RunAsync_BoundsConcurrentOperations()
    {
        using var limiter = new PageSpeedRequestLimiter(2);
        var active = 0;
        var maxActive = 0;

        var tasks = Enumerable.Range(0, 6).Select(_ =>
            limiter.RunAsync(
                async cancellationToken =>
                {
                    var current = Interlocked.Increment(ref active);
                    while (true)
                    {
                        var currentMax = Volatile.Read(ref maxActive);
                        if (current <= currentMax ||
                            Interlocked.CompareExchange(
                                ref maxActive,
                                current,
                                currentMax) == currentMax)
                        {
                            break;
                        }
                    }
                    try
                    {
                        await Task.Delay(10, cancellationToken);
                        return current;
                    }
                    finally
                    {
                        Interlocked.Decrement(ref active);
                    }
                },
                CancellationToken.None));

        await Task.WhenAll(tasks);

        Assert.Equal(2, maxActive);
    }

    /// <summary>Verifies cancellation interrupts a caller waiting for capacity.</summary>
    [Fact]
    public async Task RunAsync_CancellationStopsWaitingCaller()
    {
        using var limiter = new PageSpeedRequestLimiter(1);
        var release = new TaskCompletionSource<bool>(
            TaskCreationOptions.RunContinuationsAsynchronously);
        var occupied = limiter.RunAsync(
            async _ =>
            {
                await release.Task;
                return true;
            },
            CancellationToken.None);

        using var cancellation = new CancellationTokenSource();
        cancellation.Cancel();

        await Assert.ThrowsAnyAsync<OperationCanceledException>(() =>
            limiter.RunAsync(_ => Task.FromResult(true), cancellation.Token));

        release.SetResult(true);
        await occupied;
    }
}
