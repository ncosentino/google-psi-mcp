package main

import (
	"context"
	"slices"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/crux"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

// TestNewServer_RegistersTools verifies that the MCP server can be created and all
// tools can be registered without panicking. This catches invalid struct tags or
// schema-generation failures at test time rather than at runtime.
func TestNewServer_RegistersTools(t *testing.T) {
	t.Parallel()

	srv := newServer(pagespeed.NewClient("test-key"), crux.NewClient("test-key"))
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	for _, name := range []string{
		"analyze_page",
		"analyze_pages",
		"get_crux_data",
		"get_crux_history",
	} {
		found := false
		for _, tool := range result.Tools {
			if tool.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("tool %q not registered", name)
		}
	}
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

func TestCruxInputs_OptionalControls_AreNotRequired(t *testing.T) {
	t.Parallel()

	dataSchema, err := jsonschema.For[cruxDataInput](nil)
	if err != nil {
		t.Fatalf("current CrUX schema inference failed: %v", err)
	}
	for _, field := range []string{"target_type", "form_factor", "metrics"} {
		if slices.Contains(dataSchema.Required, field) {
			t.Errorf("get_crux_data field %s must not be required", field)
		}
	}

	historySchema, err := jsonschema.For[cruxHistoryInput](nil)
	if err != nil {
		t.Fatalf("history CrUX schema inference failed: %v", err)
	}
	for _, field := range []string{
		"target_type",
		"form_factor",
		"metrics",
		"collection_period_count",
	} {
		if slices.Contains(historySchema.Required, field) {
			t.Errorf("get_crux_history field %s must not be required", field)
		}
	}
}
