// Command google-psi-mcp is an MCP server that exposes Google PageSpeed Insights
// as tools for AI assistants. It supports STDIO and Streamable HTTP transports.
//
// Usage:
//
//	google-psi-mcp [--api-key <key>] [--transport stdio|http]
//	    [--listen-address <address>] [--port <port>]
//	    [--allowed-hosts <list>]
//
// API key resolution order: --api-key flag, GOOGLE_PSI_API_KEY env var, .env file.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/apihttp"
	"github.com/ncosentino/google-psi-mcp/go/internal/config"
	"github.com/ncosentino/google-psi-mcp/go/internal/crux"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

var version = "dev"

const (
	maxBatchURLs          = 10
	maxConcurrentAnalyses = 4
)

type pageAnalyzer interface {
	Analyze(context.Context, pagespeed.AnalysisRequest) (*pagespeed.AnalysisResult, error)
}

type cruxQuerier interface {
	QueryCurrent(context.Context, crux.QueryRequest) (*crux.Result, error)
	QueryHistory(context.Context, crux.QueryRequest) (*crux.HistoryResult, error)
}

type analysisFailure struct {
	InputURL  string `json:"inputUrl"`
	Strategy  string `json:"strategy"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type analysisResponse struct {
	Results []*pagespeed.AnalysisResult `json:"results"`
	Errors  []analysisFailure           `json:"errors"`
}

func main() {
	apiKeyFlag := flag.String("api-key", "", "Google API key for PageSpeed Insights and CrUX")
	transport := flag.String("transport", "stdio", "Transport mode: stdio or http")
	listenAddress := flag.String(
		"listen-address",
		"",
		"HTTP listen address (default MCP_LISTEN_ADDRESS or 127.0.0.1)",
	)
	port := flag.Int("port", 0, "HTTP listen port (default PORT or 8080)")
	allowedHosts := flag.String(
		"allowed-hosts",
		"localhost,127.0.0.1,[::1]",
		"Comma-separated Host header allow-list for HTTP transport",
	)
	flag.Parse()

	// All diagnostic output must go to stderr to avoid corrupting the MCP STDIO stream.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.Resolve(*apiKeyFlag)
	if cfg.APIKey == "" {
		slog.Error("no API key provided",
			"hint", "set --api-key flag, GOOGLE_PSI_API_KEY env var, or add it to a .env file")
		os.Exit(1)
	}

	client := pagespeed.NewClient(cfg.APIKey)
	cruxClient := crux.NewClient(cfg.APIKey)

	srv := newServer(client, cruxClient)

	switch *transport {
	case "stdio":
		slog.Info("google-psi-mcp starting", "version", version, "transport", "stdio")
		if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			slog.Error("server stopped with error", "err", err)
			os.Exit(1)
		}
	case "http":
		httpPort, err := resolveHTTPPort(*port)
		if err != nil {
			slog.Error("invalid HTTP port", "err", err)
			os.Exit(1)
		}
		ctx, stop := signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)
		defer stop()
		if err := runHTTP(ctx, srv, httpServerOptions{
			ListenAddress: resolveHTTPListenAddress(*listenAddress),
			Port:          httpPort,
			AllowedHosts:  splitAndTrim(*allowedHosts),
			ShutdownToken: strings.TrimSpace(os.Getenv("MCP_SHUTDOWN_TOKEN")),
		}); err != nil {
			slog.Error("server stopped with error", "err", err)
			os.Exit(1)
		}
	default:
		slog.Error("invalid transport", "transport", *transport, "expected", "stdio or http")
		os.Exit(1)
	}
}

// newServer builds the MCP server independently of its transport.
func newServer(client pageAnalyzer, cruxClient cruxQuerier) *mcp.Server {
	client = newLimitedPageAnalyzer(client, maxConcurrentAnalyses)
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-psi-mcp",
		Version: version,
	}, nil)
	srv.AddReceivingMiddleware(coerceStringifiedArrayArgs(toolArrayFields))

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_page",
			Description: "Analyze a single URL using Google PageSpeed Insights. Separates real-user CrUX field data from synthetic Lighthouse lab data and returns Lighthouse 13 insights with structured details. strategy defaults to both. categories defaults to performance, SEO, accessibility, and best-practices; agentic-browsing is experimental and must be requested explicitly.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePageInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, []string{input.URL}, input.Strategy, input.Categories, input.Locale)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_crux_data",
			Description: "Get current Chrome UX Report real-user data for a URL or origin. Supports all current CrUX metrics, including Core Web Vitals, LCP subparts, navigation types, RTT, resource types, and form-factor fractions. Requires the Chrome UX Report API to be enabled and allowed for the configured API key.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input cruxDataInput) (*mcp.CallToolResult, any, error) {
			request, err := crux.NewQueryRequest(
				input.Target,
				input.TargetType,
				input.FormFactor,
				input.Metrics,
				0,
			)
			if err != nil {
				return nil, nil, err
			}
			result, err := cruxClient.QueryCurrent(ctx, request)
			if err != nil {
				return nil, nil, err
			}
			return jsonToolResult(result)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "get_crux_history",
			Description: "Get up to 40 weekly Chrome UX Report collection periods for a URL or origin. Returns real-user metric timeseries with null values for unavailable periods. Requires the Chrome UX Report API to be enabled and allowed for the configured API key.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input cruxHistoryInput) (*mcp.CallToolResult, any, error) {
			request, err := crux.NewQueryRequest(
				input.Target,
				input.TargetType,
				input.FormFactor,
				input.Metrics,
				input.CollectionPeriodCount,
			)
			if err != nil {
				return nil, nil, err
			}
			result, err := cruxClient.QueryHistory(ctx, request)
			if err != nil {
				return nil, nil, err
			}
			return jsonToolResult(result)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_pages",
			Description: "Analyze multiple URLs using Google PageSpeed Insights. Returns separate real-user field data and Lighthouse lab data for every URL and strategy. strategy defaults to both. categories defaults to performance, SEO, accessibility, and best-practices; agentic-browsing is experimental and must be requested explicitly.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePagesInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, input.URLs, input.Strategy, input.Categories, input.Locale)
		},
	)

	return srv
}

// analyzePageInput is the input schema for the analyze_page tool.
type analyzePageInput struct {
	URL        string   `json:"url"`
	Strategy   string   `json:"strategy,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Locale     string   `json:"locale,omitempty"`
}

// analyzePagesInput is the input schema for the analyze_pages tool.
type analyzePagesInput struct {
	URLs       []string `json:"urls"`
	Strategy   string   `json:"strategy,omitempty"`
	Categories []string `json:"categories,omitempty"`
	Locale     string   `json:"locale,omitempty"`
}

// cruxDataInput is the input schema for current Chrome UX Report data.
type cruxDataInput struct {
	Target     string   `json:"target"`
	TargetType string   `json:"target_type,omitempty"`
	FormFactor string   `json:"form_factor,omitempty"`
	Metrics    []string `json:"metrics,omitempty"`
}

// cruxHistoryInput is the input schema for historical Chrome UX Report data.
type cruxHistoryInput struct {
	Target                string   `json:"target"`
	TargetType            string   `json:"target_type,omitempty"`
	FormFactor            string   `json:"form_factor,omitempty"`
	Metrics               []string `json:"metrics,omitempty"`
	CollectionPeriodCount int      `json:"collection_period_count,omitempty"`
}

// analyzePages runs PSI analysis for the given URLs and strategy, returning a JSON tool result.
func analyzePages(
	ctx context.Context,
	client pageAnalyzer,
	urls []string,
	strategy string,
	categories []string,
	locale string,
) (*mcp.CallToolResult, any, error) {
	if len(urls) == 0 {
		return nil, nil, fmt.Errorf("at least one URL is required")
	}
	if len(urls) > maxBatchURLs {
		return nil, nil, fmt.Errorf("at most %d URLs may be analyzed per call", maxBatchURLs)
	}

	strategies, err := pagespeed.ResolveStrategies(strategy)
	if err != nil {
		return nil, nil, err
	}

	requests := make([]pagespeed.AnalysisRequest, 0, len(urls)*len(strategies))
	for _, inputURL := range urls {
		for _, selectedStrategy := range strategies {
			request, err := pagespeed.NewAnalysisRequest(
				inputURL,
				selectedStrategy,
				categories,
				locale,
			)
			if err != nil {
				return nil, nil, err
			}
			requests = append(requests, request)
		}
	}

	type analysisEntry struct {
		result  *pagespeed.AnalysisResult
		failure *analysisFailure
	}

	entries := make([]analysisEntry, len(requests))
	var waitGroup sync.WaitGroup
	for index, request := range requests {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			result, err := client.Analyze(ctx, request)
			if err != nil {
				slog.Warn(
					"PSI analysis failed",
					"url",
					request.URL,
					"strategy",
					request.Strategy,
					"err",
					err,
				)
				failure := classifyAnalysisFailure(request, err)
				entries[index].failure = &failure
				return
			}
			entries[index].result = result
		}()
	}
	waitGroup.Wait()

	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	response := analysisResponse{
		Results: make([]*pagespeed.AnalysisResult, 0, len(entries)),
		Errors:  make([]analysisFailure, 0),
	}
	for _, entry := range entries {
		if entry.result != nil {
			response.Results = append(response.Results, entry.result)
		}
		if entry.failure != nil {
			response.Errors = append(response.Errors, *entry.failure)
		}
	}

	return jsonToolResult(response)
}

func classifyAnalysisFailure(
	request pagespeed.AnalysisRequest,
	err error,
) analysisFailure {
	failure := analysisFailure{
		InputURL: request.URL,
		Strategy: request.Strategy,
		Code:     "request_failed",
		Message:  err.Error(),
	}

	var statusError *apihttp.StatusError
	if errors.As(err, &statusError) {
		failure.Retryable = statusError.Retryable()
		switch {
		case statusError.StatusCode == 429:
			failure.Code = "rate_limited"
		case statusError.StatusCode >= 500:
			failure.Code = "upstream_unavailable"
		default:
			failure.Code = "upstream_rejected"
		}
		return failure
	}

	var networkError net.Error
	if errors.Is(err, context.DeadlineExceeded) ||
		(errors.As(err, &networkError) && networkError.Timeout()) {
		failure.Code = "timeout"
		failure.Retryable = true
	}
	return failure
}

func jsonToolResult(value any) (*mcp.CallToolResult, any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling tool result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(encoded)},
		},
	}, nil, nil
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			result = append(result, part)
		}
	}
	return result
}
