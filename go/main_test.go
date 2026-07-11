package main

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ncosentino/google-psi-mcp/go/internal/apihttp"
	"github.com/ncosentino/google-psi-mcp/go/internal/crux"
	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

type trackingAnalyzer struct {
	active    atomic.Int32
	maxActive atomic.Int32
	calls     atomic.Int32
}

type fakeCruxQuerier struct{}

func (fakeCruxQuerier) QueryCurrent(
	_ context.Context,
	request crux.QueryRequest,
) (*crux.Result, error) {
	return &crux.Result{
		Target:     request.Target,
		TargetType: request.TargetType,
		Metrics:    map[string]crux.Metric{},
	}, nil
}

func (fakeCruxQuerier) QueryHistory(
	_ context.Context,
	request crux.QueryRequest,
) (*crux.HistoryResult, error) {
	return &crux.HistoryResult{
		Target:     request.Target,
		TargetType: request.TargetType,
		Metrics:    map[string]crux.HistoryMetric{},
	}, nil
}

func (a *trackingAnalyzer) Analyze(
	_ context.Context,
	request pagespeed.AnalysisRequest,
) (*pagespeed.AnalysisResult, error) {
	a.calls.Add(1)
	active := a.active.Add(1)
	for {
		currentMax := a.maxActive.Load()
		if active <= currentMax || a.maxActive.CompareAndSwap(currentMax, active) {
			break
		}
	}
	defer a.active.Add(-1)
	time.Sleep(10 * time.Millisecond)
	return &pagespeed.AnalysisResult{
		Metadata: pagespeed.AnalysisMetadata{
			InputURL: request.URL,
			Strategy: request.Strategy,
		},
	}, nil
}

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

func TestAnalyzePages_BoundsConcurrencyAndPreservesAllResults(t *testing.T) {
	t.Parallel()

	analyzer := &trackingAnalyzer{}
	result, _, err := analyzePages(
		context.Background(),
		analyzer,
		[]string{
			"https://example.test/one",
			"https://example.test/two",
			"https://example.test/three",
		},
		"both",
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("analyzePages: %v", err)
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content type = %T, want *mcp.TextContent", result.Content[0])
	}
	var response analysisResponse
	if err := json.Unmarshal([]byte(text.Text), &response); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(response.Results) != 6 {
		t.Errorf("results = %d, want 6", len(response.Results))
	}
	if len(response.Errors) != 0 {
		t.Errorf("errors = %v, want none", response.Errors)
	}
	if max := analyzer.maxActive.Load(); max > maxConcurrentAnalyses {
		t.Errorf("max concurrency = %d, want at most %d", max, maxConcurrentAnalyses)
	}
	if max := analyzer.maxActive.Load(); max < 2 {
		t.Errorf("max concurrency = %d, expected concurrent execution", max)
	}
}

func TestAnalyzePages_RejectsOversizedBatchBeforeCallingAPI(t *testing.T) {
	t.Parallel()

	analyzer := &trackingAnalyzer{}
	urls := make([]string, maxBatchURLs+1)
	for index := range urls {
		urls[index] = "https://example.test"
	}

	if _, _, err := analyzePages(
		context.Background(),
		analyzer,
		urls,
		"mobile",
		nil,
		"",
	); err == nil {
		t.Fatal("analyzePages returned nil error")
	}
	if calls := analyzer.calls.Load(); calls != 0 {
		t.Errorf("API calls = %d, want 0", calls)
	}
}

func TestClassifyAnalysisFailure_UsesStructuredRetryability(t *testing.T) {
	t.Parallel()

	request, err := pagespeed.NewAnalysisRequest(
		"https://example.test",
		"mobile",
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("NewAnalysisRequest: %v", err)
	}

	rateLimit := classifyAnalysisFailure(request, &apihttp.StatusError{
		Service:    "PSI API",
		StatusCode: http.StatusTooManyRequests,
	})
	if rateLimit.Code != "rate_limited" || !rateLimit.Retryable {
		t.Errorf("rate-limit failure = %+v", rateLimit)
	}

	badRequest := classifyAnalysisFailure(request, &apihttp.StatusError{
		Service:    "PSI API",
		StatusCode: http.StatusBadRequest,
	})
	if badRequest.Code != "upstream_rejected" || badRequest.Retryable {
		t.Errorf("bad-request failure = %+v", badRequest)
	}
}
