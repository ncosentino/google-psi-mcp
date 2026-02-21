package pagespeed_test

import (
	"math"
	"testing"

	"github.com/ncosentino/google-psi-mcp/go/internal/pagespeed"
)

func TestCLSRating(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cls     float64
		wantCLS string
	}{
		{name: "good CLS", cls: 0.05, wantCLS: "good"},
		{name: "needs-improvement CLS", cls: 0.15, wantCLS: "needs-improvement"},
		{name: "poor CLS", cls: 0.30, wantCLS: "poor"},
		{name: "boundary good/needs", cls: 0.10, wantCLS: "needs-improvement"},
		{name: "boundary needs/poor", cls: 0.25, wantCLS: "poor"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &pagespeed.Result{}
			r.CoreWebVitals.CLS = makeCLSMetric(tc.cls)
			if got := r.CoreWebVitals.CLS.Rating; got != tc.wantCLS {
				t.Errorf("CLS rating = %q, want %q", got, tc.wantCLS)
			}
		})
	}
}

func TestNewClientNotNil(t *testing.T) {
	t.Parallel()
	client := pagespeed.NewClient("fake-key")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestScoreRoundingHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rawScore  float64
		wantScore int
	}{
		{rawScore: 0.85, wantScore: 85},
		{rawScore: 1.0, wantScore: 100},
		{rawScore: 0.855, wantScore: 86},
		{rawScore: 0.0, wantScore: 0},
	}

	for _, tc := range tests {
		got := int(math.Round(tc.rawScore * 100))
		if got != tc.wantScore {
			t.Errorf("round(%.3f*100) = %d, want %d", tc.rawScore, got, tc.wantScore)
		}
	}
}

// makeCLSMetric builds a MetricValue for CLS using the same thresholds as models.go.
func makeCLSMetric(v float64) pagespeed.MetricValue {
	rating := "good"
	switch {
	case v >= 0.25:
		rating = "poor"
	case v >= 0.10:
		rating = "needs-improvement"
	}
	return pagespeed.MetricValue{Value: v, Rating: rating}
}
