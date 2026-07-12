package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestAllowedHostsMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		host       string
		wantStatus int
		wantCalled bool
	}{
		{name: "allowed", host: "127.0.0.1:8080", wantStatus: http.StatusOK, wantCalled: true},
		{name: "rejected", host: "evil.example", wantStatus: http.StatusForbidden, wantCalled: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			called := false
			handler := allowedHostsMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
				}),
				[]string{"127.0.0.1"},
			)
			request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/", nil)
			request.Host = test.host
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if called != test.wantCalled {
				t.Errorf("handler called = %v, want %v", called, test.wantCalled)
			}
		})
	}
}

func TestHTTPTransport_ServesRealSession(t *testing.T) {
	t.Parallel()

	srv := newServer(&trackingAnalyzer{}, fakeCruxQuerier{})
	httpServer := httptest.NewServer(buildHTTPHandler(srv, []string{"127.0.0.1"}))
	defer httpServer.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	session, err := client.Connect(
		ctx,
		&mcp.StreamableClientTransport{Endpoint: httpServer.URL},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools.Tools) != 4 {
		t.Errorf("tools = %d, want 4", len(tools.Tools))
	}

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name: "analyze_page",
		Arguments: map[string]any{
			"url":      "https://example.test",
			"strategy": "mobile",
		},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if result.IsError {
		t.Errorf("CallTool returned error content: %+v", result.Content)
	}
}

func TestHTTPTransport_RejectsForgedCrossSiteOrigin(t *testing.T) {
	t.Parallel()

	srv := newServer(&trackingAnalyzer{}, fakeCruxQuerier{})
	httpServer := httptest.NewServer(buildHTTPHandler(srv, []string{"127.0.0.1"}))
	defer httpServer.Close()

	request, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		httpServer.URL,
		nil,
	)
	if err != nil {
		t.Fatalf("NewRequestWithContext: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://evil.example")
	request.Header.Set("Sec-Fetch-Site", "cross-site")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d, want 403", response.StatusCode)
	}
}

func TestResolveHTTPPort(t *testing.T) {
	t.Setenv("PORT", "9999")
	if got := resolveHTTPPort(); got != "9999" {
		t.Errorf("port = %q, want 9999", got)
	}

	t.Setenv("PORT", "")
	if got := resolveHTTPPort(); got != "8080" {
		t.Errorf("port = %q, want 8080", got)
	}
}

func TestSplitAndTrim(t *testing.T) {
	t.Parallel()

	got := splitAndTrim("localhost, 127.0.0.1 ,, [::1]")
	want := []string{"localhost", "127.0.0.1", "[::1]"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("got[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
