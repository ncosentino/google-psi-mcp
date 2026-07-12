package crux

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseCurrentFixture_PreservesHistogramsPercentilesAndFractions(t *testing.T) {
	t.Parallel()

	var raw rawResponse
	loadFixture(t, "crux-current.json", &raw)
	result := parseCurrent(&raw)
	if result == nil {
		t.Fatal("parseCurrent returned nil")
	}

	if result.Target != "https://example.test/page" || result.TargetType != "url" {
		t.Errorf("target = %q (%s), want fixture URL", result.Target, result.TargetType)
	}
	if result.FormFactor != "phone" {
		t.Errorf("form factor = %q, want phone", result.FormFactor)
	}
	if got := result.Metrics["largest_contentful_paint"].P75; got == nil || *got != 3100 {
		t.Errorf("LCP p75 = %v, want 3100", got)
	}
	if got := result.Metrics["cumulative_layout_shift"].P75; got == nil || *got != 0.08 {
		t.Errorf("CLS p75 = %v, want 0.08", got)
	}
	if got := result.Metrics["navigation_types"].Fractions["back_forward_cache"]; got != 0.15 {
		t.Errorf("bfcache fraction = %v, want 0.15", got)
	}
	if result.URLNormalization == nil ||
		result.URLNormalization.NormalizedURL != "https://example.test/page" {
		t.Fatal("URL normalization details were not preserved")
	}
}

func TestParseHistoryFixture_ConvertsUnavailableValuesToNull(t *testing.T) {
	t.Parallel()

	var raw rawHistoryResponse
	loadFixture(t, "crux-history.json", &raw)
	result := parseHistory(&raw)
	if result == nil {
		t.Fatal("parseHistory returned nil")
	}

	if len(result.CollectionPeriods) != 3 {
		t.Fatalf("collection periods = %d, want 3", len(result.CollectionPeriods))
	}
	lcp := result.Metrics["largest_contentful_paint"]
	if len(lcp.P75) != 3 || lcp.P75[1] != nil {
		t.Errorf("LCP p75 = %v, want unavailable middle period", lcp.P75)
	}
	if got := lcp.Histogram[0].Densities[1]; got != nil {
		t.Errorf("middle histogram density = %v, want nil", got)
	}
	cls := result.Metrics["cumulative_layout_shift"]
	if got := cls.P75[0]; got == nil || *got != 0.09 {
		t.Errorf("CLS first p75 = %v, want 0.09", got)
	}
	navigation := result.Metrics["navigation_types"]
	if got := navigation.Fractions["navigate"][2]; got == nil || *got != 0.66 {
		t.Errorf("navigate fraction = %v, want 0.66", got)
	}
}

func loadFixture(t *testing.T, name string, output any) {
	t.Helper()

	path := filepath.Join("..", "..", "..", "testdata", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if err := json.Unmarshal(data, output); err != nil {
		t.Fatalf("unmarshal %s: %v", name, err)
	}
}
