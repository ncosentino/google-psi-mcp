// Package pagespeed provides a client for the Google PageSpeed Insights API v5.
package pagespeed

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	apiBaseURL = "https://www.googleapis.com/pagespeedonline/v5/runPagespeed"
	httpTimeout = 60 * time.Second
)

// Client calls the Google PageSpeed Insights API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a Client configured with the provided API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: httpTimeout},
	}
}

// Analyze runs a PageSpeed Insights analysis for the given URL and strategy.
// Strategy must be "mobile" or "desktop".
func (c *Client) Analyze(ctx context.Context, targetURL, strategy string) (*Result, error) {
	req, err := c.buildRequest(ctx, targetURL, strategy)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PSI API returned HTTP %d: %s", resp.StatusCode, truncate(string(body), 300))
	}

	var raw apiResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing PSI response: %w", err)
	}

	return parseResult(targetURL, strategy, &raw), nil
}

func (c *Client) buildRequest(ctx context.Context, targetURL, strategy string) (*http.Request, error) {
	params := url.Values{}
	params.Set("url", targetURL)
	params.Set("strategy", strategy)
	params.Set("key", c.apiKey)
	for _, cat := range []string{"performance", "seo", "accessibility", "best-practices"} {
		params.Add("category", cat)
	}

	fullURL := apiBaseURL + "?" + params.Encode()
	return http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
