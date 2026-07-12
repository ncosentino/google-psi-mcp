package crux

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestQueryCurrent_SendsValidatedCrUXRequest(t *testing.T) {
	t.Parallel()

	fixture := readFixture(t, "crux-current.json")
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := &Client{
		apiKey:        "test-key",
		httpClient:    server.Client(),
		currentAPIURL: server.URL,
		historyAPIURL: server.URL,
	}
	request, err := NewQueryRequest(
		"https://example.test/page",
		"url",
		"phone",
		[]string{"largest_contentful_paint", "largest_contentful_paint"},
		0,
	)
	if err != nil {
		t.Fatalf("NewQueryRequest: %v", err)
	}

	result, err := client.QueryCurrent(context.Background(), request)
	if err != nil {
		t.Fatalf("QueryCurrent: %v", err)
	}

	if result.Target != "https://example.test/page" {
		t.Errorf("result target = %q, want fixture target", result.Target)
	}
	if requestBody["formFactor"] != "PHONE" {
		t.Errorf("formFactor = %v, want PHONE", requestBody["formFactor"])
	}
	if got := requestBody["metrics"]; !reflect.DeepEqual(got, []any{"largest_contentful_paint"}) {
		t.Errorf("metrics = %v, want one deduplicated metric", got)
	}
}

func TestQueryHistory_SendsCollectionPeriodCount(t *testing.T) {
	t.Parallel()

	fixture := readFixture(t, "crux-history.json")
	var requestBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer server.Close()

	client := &Client{
		apiKey:        "test-key",
		httpClient:    server.Client(),
		currentAPIURL: server.URL,
		historyAPIURL: server.URL,
	}
	request, err := NewQueryRequest(
		"https://example.test",
		"origin",
		"desktop",
		nil,
		3,
	)
	if err != nil {
		t.Fatalf("NewQueryRequest: %v", err)
	}

	if _, err := client.QueryHistory(context.Background(), request); err != nil {
		t.Fatalf("QueryHistory: %v", err)
	}
	if requestBody["collectionPeriodCount"] != float64(3) {
		t.Errorf("collectionPeriodCount = %v, want 3", requestBody["collectionPeriodCount"])
	}
	if _, ok := requestBody["metrics"]; ok {
		t.Error("metrics must be omitted when none are requested")
	}
}

func TestNewQueryRequest_InvalidInput_IsRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		target       string
		targetType   string
		formFactor   string
		metrics      []string
		historyCount int
	}{
		{name: "relative target", target: "/relative", targetType: "url"},
		{name: "origin with path", target: "https://example.test/path", targetType: "origin"},
		{name: "invalid target type", target: "https://example.test", targetType: "site"},
		{name: "invalid form factor", target: "https://example.test", targetType: "origin", formFactor: "watch"},
		{
			name:       "invalid metric",
			target:     "https://example.test",
			targetType: "origin",
			metrics:    []string{"LCP!"},
		},
		{
			name:         "too many periods",
			target:       "https://example.test",
			targetType:   "origin",
			historyCount: 41,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewQueryRequest(
				test.target,
				test.targetType,
				test.formFactor,
				test.metrics,
				test.historyCount,
			); err == nil {
				t.Fatal("NewQueryRequest returned nil error")
			}
		})
	}
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()

	path := filepath.Join("..", "..", "..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}
