namespace PageSpeedMcp.Infrastructure;

/// <summary>Bounds PageSpeed API concurrency across every MCP client connected to the process.</summary>
internal sealed class PageSpeedRequestLimiter : IDisposable
{
    internal const int DefaultMaxConcurrency = 4;

    private readonly SemaphoreSlim _semaphore;

    internal PageSpeedRequestLimiter(int maxConcurrency = DefaultMaxConcurrency)
    {
        ArgumentOutOfRangeException.ThrowIfLessThan(maxConcurrency, 1);
        _semaphore = new SemaphoreSlim(maxConcurrency, maxConcurrency);
    }

    internal async Task<T> RunAsync<T>(
        Func<CancellationToken, Task<T>> operation,
        CancellationToken cancellationToken)
    {
        await _semaphore.WaitAsync(cancellationToken).ConfigureAwait(false);
        try
        {
            return await operation(cancellationToken).ConfigureAwait(false);
        }
        finally
        {
            _semaphore.Release();
        }
    }

    public void Dispose() => _semaphore.Dispose();
}
