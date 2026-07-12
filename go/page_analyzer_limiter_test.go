package main

import (
	"context"
	"sync"
	"testing"

	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

func TestLimitedPageAnalyzer_BoundsConcurrentCalls(t *testing.T) {
	t.Parallel()

	analyzer := &trackingAnalyzer{}
	limited := newLimitedPageAnalyzer(analyzer, 2)
	request, err := pagespeed.NewAnalysisRequest(
		"https://example.test",
		"mobile",
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("NewAnalysisRequest: %v", err)
	}

	var waitGroup sync.WaitGroup
	for range 6 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if _, err := limited.Analyze(context.Background(), request); err != nil {
				t.Errorf("Analyze: %v", err)
			}
		}()
	}
	waitGroup.Wait()

	if max := analyzer.maxActive.Load(); max > 2 {
		t.Errorf("max concurrency = %d, want at most 2", max)
	}
}

func TestLimitedPageAnalyzer_RespectsCancellationWhileWaiting(t *testing.T) {
	t.Parallel()

	analyzer := &trackingAnalyzer{}
	limited := newLimitedPageAnalyzer(analyzer, 1)
	limited.slots <- struct{}{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request, err := pagespeed.NewAnalysisRequest(
		"https://example.test",
		"mobile",
		nil,
		"",
	)
	if err != nil {
		t.Fatalf("NewAnalysisRequest: %v", err)
	}

	if _, err := limited.Analyze(ctx, request); err == nil {
		t.Fatal("Analyze returned nil error")
	}
}
