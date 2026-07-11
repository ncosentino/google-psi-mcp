package main

import (
	"context"
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

// TestNewServer_RegistersTools verifies that the MCP server can be created and all
// tools can be registered without panicking. This catches invalid struct tags or
// schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(_ *testing.T) {
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
			return analyzePages(
				ctx,
				client,
				[]string{input.URL},
				input.Strategy,
				input.Categories,
				input.Locale,
			)
		},
	)

	mcp.AddTool(srv,
		&mcp.Tool{
			Name:        "analyze_pages",
			Description: "test",
		},
		func(ctx context.Context, _ *mcp.CallToolRequest, input analyzePagesInput) (*mcp.CallToolResult, any, error) {
			return analyzePages(
				ctx,
				client,
				input.URLs,
				input.Strategy,
				input.Categories,
				input.Locale,
			)
		},
	)
}

func TestAnalyzeInputs_OptionalControls_AreNotRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		infer func() (*jsonschema.Schema, error)
	}{
		{
			name: "analyze_page",
			infer: func() (*jsonschema.Schema, error) {
				return jsonschema.For[analyzePageInput](nil)
			},
		},
		{
			name: "analyze_pages",
			infer: func() (*jsonschema.Schema, error) {
				return jsonschema.For[analyzePagesInput](nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			schema, err := test.infer()
			if err != nil {
				t.Fatalf("schema inference failed: %v", err)
			}
			for _, field := range []string{"strategy", "categories", "locale"} {
				if slices.Contains(schema.Required, field) {
					t.Errorf("%s must not be required (got %v)", field, schema.Required)
				}
			}
		})
	}
}
