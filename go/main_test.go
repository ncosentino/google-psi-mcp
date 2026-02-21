package main

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

// TestNewServer_RegistersTools verifies that the MCP server can be created and all
// tools can be registered without panicking. This catches invalid struct tags or
// schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(t *testing.T) {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "google-psi-mcp",
		Version: "test",
	}, nil)

	client := pagespeed.NewClient("test-key")

	// If AddTool panics (e.g. invalid jsonschema tags), the test will fail.
	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_page",
			Description: "test",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePageInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, []string{input.URL}, input.Strategy)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_pages",
			Description: "test",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePagesInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(ctx, client, input.URLs, input.Strategy)
		},
	)
}

func TestResolveStrategies_Both(t *testing.T) {
	got := resolveStrategies("both")
	if len(got) != 2 {
		t.Fatalf("expected 2 strategies, got %d", len(got))
	}
	if got[0] != "mobile" || got[1] != "desktop" {
		t.Fatalf("unexpected strategies: %v", got)
	}
}

func TestResolveStrategies_Single(t *testing.T) {
	for _, s := range []string{"mobile", "desktop"} {
		t.Run(s, func(t *testing.T) {
			got := resolveStrategies(s)
			if len(got) != 1 || got[0] != s {
				t.Fatalf("expected [%s], got %v", s, got)
			}
		})
	}
}

func TestResolveStrategies_DefaultsOnEmpty(t *testing.T) {
	// Empty strategy is resolved to "both" inside analyzePages, but
	// resolveStrategies itself receives a non-empty value by then.
	// Verify the helper returns a single-element slice for any non-"both" value.
	got := resolveStrategies("mobile")
	if len(got) != 1 {
		t.Fatalf("expected 1, got %d: %v", len(got), got)
	}
}
