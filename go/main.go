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

	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-psi-mcp",
		Version: version,
	}, nil)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_page",
			Description: "Analyze a single URL using Google PageSpeed Insights. Returns Core Web Vitals scores, category scores (performance, SEO, accessibility, best-practices), and actionable audit findings.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePageInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, []string{input.URL}, input.Strategy)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_pages",
			Description: "Analyze multiple URLs using Google PageSpeed Insights in a single call. Returns an array of results, one per URL.",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePagesInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, input.URLs, input.Strategy)
		},
	)

	slog.Info("google-psi-mcp starting", "version", version, "transport", "stdio")
	if err := srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

// analyzePageInput is the input schema for the analyze_page tool.
type analyzePageInput struct {
	URL      string `json:"url"`
	Strategy string `json:"strategy"`
}

// analyzePagesInput is the input schema for the analyze_pages tool.
type analyzePagesInput struct {
	URLs     []string `json:"urls"`
	Strategy string   `json:"strategy"`
}

// analyzePages runs PSI analysis for the given URLs and strategy, returning a JSON tool result.
func analyzePages(ctx context.Context, client *pagespeed.Client, urls []string, strategy string) (*mcp.CallToolResult, any, error) {
	if strategy == "" {
		strategy = "both"
	}

	strategies := resolveStrategies(strategy)

	type entry struct {
		URL      string            `json:"url"`
		Results  []*pagespeed.Result `json:"results"`
		Error    string            `json:"error,omitempty"`
	}

	entries := make([]entry, 0, len(urls))

	for _, u := range urls {
		var results []*pagespeed.Result
		var errMsg string

		for _, s := range strategies {
			r, err := client.Analyze(ctx, u, s)
			if err != nil {
				errMsg = fmt.Sprintf("error analyzing %s (%s): %v", u, s, err)
				slog.Warn("PSI analysis failed", "url", u, "strategy", s, "err", err)
				break
			}
			results = append(results, r)
		}

		entries = append(entries, entry{URL: u, Results: results, Error: errMsg})
	}

	b, err := json.Marshal(entries)
	if err != nil {
		return nil, nil, fmt.Errorf("marshalling results: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(b)},
		},
	}, nil, nil
}

func resolveStrategies(strategy string) []string {
	if strategy == "both" {
		return []string{"mobile", "desktop"}
	}
	return []string{strategy}
}
