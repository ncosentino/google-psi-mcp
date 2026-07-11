package crux

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	defaultCurrentAPIURL = "https://chromeuxreport.googleapis.com/v1/records:queryRecord"
	defaultHistoryAPIURL = "https://chromeuxreport.googleapis.com/v1/records:queryHistoryRecord"
	httpTimeout          = 30 * time.Second
)

var metricNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Client calls the current and historical Chrome UX Report APIs.
type Client struct {
	apiKey        string
	httpClient    *http.Client
	currentAPIURL string
	historyAPIURL string
}

// NewClient returns a Chrome UX Report client configured with the provided API key.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:        apiKey,
		httpClient:    &http.Client{Timeout: httpTimeout},
		currentAPIURL: defaultCurrentAPIURL,
		historyAPIURL: defaultHistoryAPIURL,
	}
}

// QueryRequest contains one validated current or historical CrUX request.
type QueryRequest struct {
	// Target is the absolute URL or origin to query.
	Target string
	// TargetType is url or origin.
	TargetType string
	// FormFactor is all, phone, tablet, or desktop.
	FormFactor string
	// Metrics optionally limits the returned CrUX metrics.
	Metrics []string
	// CollectionPeriodCount limits history results to between 1 and 40 periods.
	CollectionPeriodCount int
}

// NewQueryRequest validates and normalizes Chrome UX Report tool input.
func NewQueryRequest(
	target,
	targetType,
	formFactor string,
	metrics []string,
	collectionPeriodCount int,
) (QueryRequest, error) {
	parsedTarget, err := url.ParseRequestURI(strings.TrimSpace(target))
	if err != nil || parsedTarget.Scheme == "" || parsedTarget.Host == "" {
		return QueryRequest{}, fmt.Errorf("target must be an absolute HTTP or HTTPS URL")
	}
	if parsedTarget.Scheme != "http" && parsedTarget.Scheme != "https" {
		return QueryRequest{}, fmt.Errorf("target scheme must be http or https")
	}

	targetType = strings.ToLower(strings.TrimSpace(targetType))
	if targetType == "" {
		targetType = "url"
	}
	if targetType != "url" && targetType != "origin" {
		return QueryRequest{}, fmt.Errorf("target_type must be url or origin")
	}
	if targetType == "origin" {
		if parsedTarget.RawQuery != "" || parsedTarget.Fragment != "" ||
			(parsedTarget.Path != "" && parsedTarget.Path != "/") {
			return QueryRequest{}, fmt.Errorf("origin targets must not include a path, query, or fragment")
		}
		parsedTarget.Path = ""
	}

	formFactor = strings.ToLower(strings.TrimSpace(formFactor))
	if formFactor == "" {
		formFactor = "all"
	}
	switch formFactor {
	case "all", "phone", "tablet", "desktop":
	default:
		return QueryRequest{}, fmt.Errorf("form_factor must be all, phone, tablet, or desktop")
	}

	normalizedMetrics := make([]string, 0, len(metrics))
	seen := make(map[string]struct{}, len(metrics))
	for _, metric := range metrics {
		metric = strings.ToLower(strings.TrimSpace(metric))
		if !metricNamePattern.MatchString(metric) {
			return QueryRequest{}, fmt.Errorf("metric %q is not a valid CrUX metric name", metric)
		}
		if _, duplicate := seen[metric]; duplicate {
			continue
		}
		seen[metric] = struct{}{}
		normalizedMetrics = append(normalizedMetrics, metric)
	}

	if collectionPeriodCount < 0 || collectionPeriodCount > 40 {
		return QueryRequest{}, fmt.Errorf("collection_period_count must be between 1 and 40")
	}

	return QueryRequest{
		Target:                parsedTarget.String(),
		TargetType:            targetType,
		FormFactor:            formFactor,
		Metrics:               normalizedMetrics,
		CollectionPeriodCount: collectionPeriodCount,
	}, nil
}

// QueryCurrent returns current 28-day Chrome UX Report data.
func (c *Client) QueryCurrent(ctx context.Context, request QueryRequest) (*Result, error) {
	var raw rawResponse
	if err := c.post(ctx, c.currentAPIURL, request, false, &raw); err != nil {
		return nil, err
	}
	result := parseCurrent(&raw)
	if result == nil {
		return nil, fmt.Errorf("CrUX API response did not contain a record")
	}
	return result, nil
}

// QueryHistory returns Chrome UX Report timeseries data.
func (c *Client) QueryHistory(ctx context.Context, request QueryRequest) (*HistoryResult, error) {
	var raw rawHistoryResponse
	if err := c.post(ctx, c.historyAPIURL, request, true, &raw); err != nil {
		return nil, err
	}
	result := parseHistory(&raw)
	if result == nil {
		return nil, fmt.Errorf("CrUX History API response did not contain a record")
	}
	return result, nil
}

func (c *Client) post(
	ctx context.Context,
	endpoint string,
	request QueryRequest,
	includeHistoryCount bool,
	output any,
) error {
	body := make(map[string]any, 4)
	body[request.TargetType] = request.Target
	if request.FormFactor != "all" {
		body["formFactor"] = strings.ToUpper(request.FormFactor)
	}
	if len(request.Metrics) > 0 {
		body["metrics"] = request.Metrics
	}
	if includeHistoryCount {
		count := request.CollectionPeriodCount
		if count == 0 {
			count = 25
		}
		body["collectionPeriodCount"] = count
	}

	encodedBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding CrUX request: %w", err)
	}

	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("parsing CrUX endpoint: %w", err)
	}
	query := endpointURL.Query()
	query.Set("key", c.apiKey)
	endpointURL.RawQuery = query.Encode()

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpointURL.String(),
		bytes.NewReader(encodedBody),
	)
	if err != nil {
		return fmt.Errorf("building CrUX request: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("executing CrUX request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading CrUX response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"CrUX API returned HTTP %d: %s",
			response.StatusCode,
			truncate(string(responseBody), 500),
		)
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("parsing CrUX response: %w", err)
	}
	return nil
}

func truncate(value string, maxLength int) string {
	if len(value) <= maxLength {
		return value
	}
	return value[:maxLength] + "..."
}
