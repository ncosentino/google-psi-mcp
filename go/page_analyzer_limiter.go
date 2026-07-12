package main

import (
	"context"

	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

type limitedPageAnalyzer struct {
	analyzer pageAnalyzer
	slots    chan struct{}
}

func newLimitedPageAnalyzer(analyzer pageAnalyzer, maxConcurrency int) *limitedPageAnalyzer {
	if maxConcurrency < 1 {
		panic("maxConcurrency must be positive")
	}
	return &limitedPageAnalyzer{
		analyzer: analyzer,
		slots:    make(chan struct{}, maxConcurrency),
	}
}

func (a *limitedPageAnalyzer) Analyze(
	ctx context.Context,
	request pagespeed.AnalysisRequest,
) (*pagespeed.AnalysisResult, error) {
	select {
	case a.slots <- struct{}{}:
		defer func() { <-a.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return a.analyzer.Analyze(ctx, request)
}
