package apihttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestDo_RetriesTransientStatus(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	response, err := Do(context.Background(), server.Client(), func() (*http.Request, error) {
		return http.NewRequestWithContext(
			context.Background(),
			http.MethodGet,
			server.URL,
			nil,
		)
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", response.StatusCode)
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("request count = %d, want 2", got)
	}
}

func TestStatusError_Retryable(t *testing.T) {
	t.Parallel()

	if !(&StatusError{StatusCode: http.StatusTooManyRequests}).Retryable() {
		t.Error("429 must be retryable")
	}
	if !(&StatusError{StatusCode: http.StatusBadGateway}).Retryable() {
		t.Error("502 must be retryable")
	}
	if (&StatusError{StatusCode: http.StatusBadRequest}).Retryable() {
		t.Error("400 must not be retryable")
	}
}
