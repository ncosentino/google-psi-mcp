// Package apihttp provides shared HTTP behavior for Google API clients.
package apihttp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	maxAttempts = 3
	baseDelay   = 250 * time.Millisecond
)

// Response contains the buffered result of one HTTP request.
type Response struct {
	// StatusCode is the upstream HTTP status code.
	StatusCode int
	// Header contains the upstream response headers.
	Header http.Header
	// Body contains the complete upstream response body.
	Body []byte
}

// StatusError reports a non-success Google API response.
type StatusError struct {
	// Service identifies the upstream Google API.
	Service string
	// StatusCode is the upstream HTTP status code.
	StatusCode int
	// BodySnippet contains a bounded response excerpt.
	BodySnippet string
}

// Error returns the formatted upstream status failure.
func (e *StatusError) Error() string {
	return fmt.Sprintf(
		"%s returned HTTP %d: %s",
		e.Service,
		e.StatusCode,
		e.BodySnippet,
	)
}

// Retryable reports whether the status is transient.
func (e *StatusError) Retryable() bool {
	return IsRetryableStatus(e.StatusCode)
}

// Do sends a request and retries transient transport and HTTP failures.
func Do(
	ctx context.Context,
	httpClient *http.Client,
	buildRequest func() (*http.Request, error),
) (*Response, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		request, err := buildRequest()
		if err != nil {
			return nil, err
		}

		httpResponse, err := httpClient.Do(request)
		if err != nil {
			lastErr = err
			if attempt == maxAttempts || ctx.Err() != nil {
				break
			}
			if err := wait(ctx, baseDelay*time.Duration(attempt)); err != nil {
				return nil, err
			}
			continue
		}

		body, readErr := io.ReadAll(httpResponse.Body)
		closeErr := httpResponse.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("reading response body: %w", readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("closing response body: %w", closeErr)
		}

		response := &Response{
			StatusCode: httpResponse.StatusCode,
			Header:     httpResponse.Header.Clone(),
			Body:       body,
		}
		if !IsRetryableStatus(response.StatusCode) || attempt == maxAttempts {
			return response, nil
		}

		delay := retryDelay(response.Header.Get("Retry-After"), attempt)
		if err := wait(ctx, delay); err != nil {
			return nil, err
		}
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, fmt.Errorf("executing HTTP request after %d attempts: %w", maxAttempts, lastErr)
}

// IsRetryableStatus reports whether a Google API status should be retried.
func IsRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func retryDelay(retryAfter string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if timestamp, err := http.ParseTime(retryAfter); err == nil {
		if delay := time.Until(timestamp); delay > 0 {
			return delay
		}
		return 0
	}
	return baseDelay * time.Duration(attempt)
}

func wait(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
