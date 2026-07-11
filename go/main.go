// Command google-psi-mcp is an MCP server that exposes Google PageSpeed Insights
// as tools for AI assistants. It communicates via STDIO using the MCP protocol.
//
// Usage:
//
//	google-psi-mcp [--api-key <key>]
//
// API key resolution order: --api-key flag, GOOGLE_PSI_API_KEY env var, .env file.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/config"
	"github.com/ncosentino/google-psi-mcp/go/internal/crux"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

var version = "dev"

func main() {
	apiKeyFlag := flag.String("api-key", "", "Google PageSpeed Insights API key")
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

	slog.Info("google-psi-mcp starting", "version", version, "transport", "stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// newServer builds the MCP server independently of its transport.
func newServer(client *pagespeed.Client, cruxClient *crux.Client) *mcp.Server {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-psi-mcp",
		Version: version,
	}, nil)

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
	client *pagespeed.Client,
	urls []string,
	strategy string,
	categories []string,
	locale string,
) (*mcp.CallToolResult, any, error) {
	strategies, err := pagespeed.ResolveStrategies(strategy)
	if err != nil {
		return nil, nil, err
	}

	type analysisError struct {
		InputURL string `json:"inputUrl"`
		Strategy string `json:"strategy"`
		Message  string `json:"message"`
	}

	type analysisResponse struct {
		Results []*pagespeed.AnalysisResult `json:"results"`
		Errors  []analysisError             `json:"errors"`
	}

	response := analysisResponse{}
	for _, u := range urls {
		for _, s := range strategies {
			analysisRequest, err := pagespeed.NewAnalysisRequest(u, s, categories, locale)
			if err != nil {
				return nil, nil, err
			}

			r, err := client.Analyze(ctx, analysisRequest)
			if err != nil {
				slog.Warn("PSI analysis failed", "url", u, "strategy", s, "err", err)
				response.Errors = append(response.Errors, analysisError{
					InputURL: u,
					Strategy: s,
					Message:  fmt.Sprintf("error analyzing %s (%s): %v", u, s, err),
				})
				continue
			}
			response.Results = append(response.Results, r)
		}
	}

	return jsonToolResult(response)
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
