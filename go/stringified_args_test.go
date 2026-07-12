package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCoerceStringifiedArray(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{name: "stringified array", raw: `"[""invalid quoting""]"`, ok: false},
		{name: "valid stringified array", raw: `"[\"performance\"]"`, ok: true},
		{name: "genuine array", raw: `["performance"]`, ok: false},
		{name: "plain string", raw: `"performance"`, ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			coerced, ok := coerceStringifiedArray(json.RawMessage(test.raw))
			if ok != test.ok {
				t.Fatalf("ok = %v, want %v", ok, test.ok)
			}
			if ok && !json.Valid(coerced) {
				t.Errorf("coerced value is not valid JSON: %s", coerced)
			}
		})
	}
}

func TestStringifiedCategories_AreCoercedBeforeValidation(t *testing.T) {
	t.Parallel()

	srv := newServer(&trackingAnalyzer{}, fakeCruxQuerier{})
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := srv.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	result, err := clientSession.CallTool(ctx, &mcp.CallToolParams{
		Name: "analyze_page",
		Arguments: map[string]any{
			"url":        "https://example.test",
			"strategy":   "mobile",
			"categories": `["performance"]`,
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("CallTool returned error content: %+v", result.Content)
	}
}
