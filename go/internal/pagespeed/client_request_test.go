package pagespeed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestAnalyze_DefaultCategories_AreSentToPSI(t *testing.T) {
	t.Parallel()

	query := make(chan url.Values, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query <- r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
		apiBaseURL: server.URL,
	}
	request, err := NewAnalysisRequest("https://example.test", "mobile", nil, "")
	if err != nil {
		t.Fatalf("NewAnalysisRequest: %v", err)
	}

	if _, err := client.Analyze(context.Background(), request); err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	got := <-query
	if categories := got["category"]; !reflect.DeepEqual(categories, defaultCategories) {
		t.Errorf("categories = %v, want %v", categories, defaultCategories)
	}
	if got.Get("locale") != "" {
		t.Errorf("locale = %q, want omitted", got.Get("locale"))
	}
}

func TestAnalyze_CustomCategoriesAndLocale_AreSentToPSI(t *testing.T) {
	t.Parallel()

	query := make(chan url.Values, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query <- r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"lighthouseResult":{"lighthouseVersion":"13.4.0"}}`))
	}))
	defer server.Close()

	client := &Client{
		apiKey:     "test-key",
		httpClient: server.Client(),
		apiBaseURL: server.URL,
	}
	request, err := NewAnalysisRequest(
		"https://example.test",
		"desktop",
		[]string{"agentic-browsing", "performance", "agentic-browsing"},
		"en-CA",
	)
	if err != nil {
		t.Fatalf("NewAnalysisRequest: %v", err)
	}

	if _, err := client.Analyze(context.Background(), request); err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	got := <-query
	wantCategories := []string{"agentic-browsing", "performance"}
	if categories := got["category"]; !reflect.DeepEqual(categories, wantCategories) {
		t.Errorf("categories = %v, want %v", categories, wantCategories)
	}
	if locale := got.Get("locale"); locale != "en-CA" {
		t.Errorf("locale = %q, want en-CA", locale)
	}
}

func TestNewAnalysisRequest_InvalidInput_IsRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		targetURL  string
		strategy   string
		categories []string
	}{
		{name: "relative URL", targetURL: "/relative", strategy: "mobile"},
		{name: "unsupported scheme", targetURL: "ftp://example.test", strategy: "mobile"},
		{name: "invalid strategy", targetURL: "https://example.test", strategy: "tablet"},
		{
			name:       "invalid category",
			targetURL:  "https://example.test",
			strategy:   "mobile",
			categories: []string{"pwa"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewAnalysisRequest(
				test.targetURL,
				test.strategy,
				test.categories,
				"",
			); err == nil {
				t.Fatal("NewAnalysisRequest returned nil error")
			}
		})
	}
}

func TestResolveStrategies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  []string
	}{
		{input: "", want: []string{"mobile", "desktop"}},
		{input: "both", want: []string{"mobile", "desktop"}},
		{input: "MOBILE", want: []string{"mobile"}},
		{input: "desktop", want: []string{"desktop"}},
	}

	for _, test := range tests {
		got, err := ResolveStrategies(test.input)
		if err != nil {
			t.Fatalf("ResolveStrategies(%q): %v", test.input, err)
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("ResolveStrategies(%q) = %v, want %v", test.input, got, test.want)
		}
	}

	if _, err := ResolveStrategies("tablet"); err == nil {
		t.Fatal("ResolveStrategies(tablet) returned nil error")
	}
}
