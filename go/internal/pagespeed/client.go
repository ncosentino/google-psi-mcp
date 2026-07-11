// Package pagespeed provides a client for the Google PageSpeed Insights API v5.
package pagespeed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ncosentino/google-psi-mcp/go/internal/apihttp"
)

const (
	defaultAPIBaseURL = "https://www.googleapis.com/pagespeedonline/v5/runPagespeed"
	httpTimeout       = 120 * time.Second
)

var defaultCategories = []string{"performance", "seo", "accessibility", "best-practices"}

var validCategories = map[string]struct{}{
	"performance":      {},
	"seo":              {},
	"accessibility":    {},
	"best-practices":   {},
	"agentic-browsing": {},
}

// Client calls the Google PageSpeed Insights API.
type Client struct {
	apiKey     string
	httpClient *http.Client
	apiBaseURL string
}

// NewClient returns a Client configured with the provided API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: httpTimeout},
		apiBaseURL: defaultAPIBaseURL,
	}
}

// AnalysisRequest contains one validated PageSpeed Insights API request.
type AnalysisRequest struct {
	// URL is the absolute HTTP or HTTPS URL to analyze.
	URL string
	// Strategy is mobile or desktop.
	Strategy string
	// Categories contains the Lighthouse categories to run.
	Categories []string
	// Locale localizes Lighthouse display strings when supplied.
	Locale string
}

// NewAnalysisRequest validates and normalizes a PageSpeed Insights request.
func NewAnalysisRequest(
	targetURL,
	strategy string,
	categories []string,
	locale string,
) (AnalysisRequest, error) {
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(targetURL))
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return AnalysisRequest{}, fmt.Errorf("url must be an absolute HTTP or HTTPS URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return AnalysisRequest{}, fmt.Errorf("url scheme must be http or https")
	}

	strategy = strings.ToLower(strings.TrimSpace(strategy))
	if strategy != "mobile" && strategy != "desktop" {
		return AnalysisRequest{}, fmt.Errorf("strategy must be mobile or desktop")
	}

	normalizedCategories, err := normalizeCategories(categories)
	if err != nil {
		return AnalysisRequest{}, err
	}

	return AnalysisRequest{
		URL:        parsedURL.String(),
		Strategy:   strategy,
		Categories: normalizedCategories,
		Locale:     strings.TrimSpace(locale),
	}, nil
}

// ResolveStrategies validates a tool strategy selection and expands both.
func ResolveStrategies(strategy string) ([]string, error) {
	switch strings.ToLower(strings.TrimSpace(strategy)) {
	case "", "both":
		return []string{"mobile", "desktop"}, nil
	case "mobile":
		return []string{"mobile"}, nil
	case "desktop":
		return []string{"desktop"}, nil
	default:
		return nil, fmt.Errorf("strategy must be mobile, desktop, or both")
	}
}

// Analyze runs one validated PageSpeed Insights request.
func (c *Client) Analyze(ctx context.Context, analysisRequest AnalysisRequest) (*AnalysisResult, error) {
	response, err := apihttp.Do(ctx, c.httpClient, func() (*http.Request, error) {
		return c.buildRequest(ctx, analysisRequest)
	})
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, &apihttp.StatusError{
			Service:     "PSI API",
			StatusCode:  response.StatusCode,
			BodySnippet: truncate(string(response.Body), 300),
		}
	}

	var raw apiResponse
	if err := json.Unmarshal(response.Body, &raw); err != nil {
		return nil, fmt.Errorf("parsing PSI response: %w", err)
	}

	return parseResult(analysisRequest.URL, analysisRequest.Strategy, &raw), nil
}

func (c *Client) buildRequest(
	ctx context.Context,
	analysisRequest AnalysisRequest,
) (*http.Request, error) {
	params := url.Values{}
	params.Set("url", analysisRequest.URL)
	params.Set("strategy", analysisRequest.Strategy)
	params.Set("key", c.apiKey)
	for _, category := range analysisRequest.Categories {
		params.Add("category", category)
	}
	if analysisRequest.Locale != "" {
		params.Set("locale", analysisRequest.Locale)
	}

	fullURL := c.apiBaseURL + "?" + params.Encode()
	return http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
}

func normalizeCategories(categories []string) ([]string, error) {
	if len(categories) == 0 {
		return append([]string(nil), defaultCategories...), nil
	}

	normalized := make([]string, 0, len(categories))
	seen := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		category = strings.ToLower(strings.TrimSpace(category))
		if _, ok := validCategories[category]; !ok {
			return nil, fmt.Errorf(
				"category %q is invalid: must be performance, seo, accessibility, best-practices, or agentic-browsing",
				category,
			)
		}
		if _, duplicate := seen[category]; duplicate {
			continue
		}
		seen[category] = struct{}{}
		normalized = append(normalized, category)
	}
	return normalized, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
