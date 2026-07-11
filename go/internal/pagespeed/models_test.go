package pagespeed

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParseResult_Lighthouse134Fixture_SeparatesFieldAndLabData(t *testing.T) {
	t.Parallel()

	raw := loadPSIFixture(t)
	result := parseResult("https://example.test/page", "mobile", raw)

	if result.Metadata.LighthouseVersion != "13.4.0" {
		t.Fatalf("LighthouseVersion = %q, want 13.4.0", result.Metadata.LighthouseVersion)
	}
	if result.Metadata.FinalURL != "https://example.test/final" {
		t.Fatalf("FinalURL = %q, want fixture final URL", result.Metadata.FinalURL)
	}
	if len(result.Metadata.RunWarnings) != 1 {
		t.Fatalf("RunWarnings count = %d, want 1", len(result.Metadata.RunWarnings))
	}

	if result.FieldData == nil || result.FieldData.Page == nil || result.FieldData.Origin == nil {
		t.Fatal("field data must contain page and origin experiences")
	}
	pageMetrics := result.FieldData.Page.Metrics
	if got := pageMetrics["lcp"].Value; got != 3100 {
		t.Errorf("page LCP = %v, want 3100 ms", got)
	}
	if got := pageMetrics["cls"].Value; got != 0.08 {
		t.Errorf("page CLS = %v, want 0.08", got)
	}
	if got := pageMetrics["inp"].Rating; got != "needs-improvement" {
		t.Errorf("page INP rating = %q, want needs-improvement", got)
	}
	if got := result.FieldData.Origin.Metrics["lcp"].Rating; got != "good" {
		t.Errorf("origin LCP rating = %q, want good", got)
	}

	if result.LabData == nil {
		t.Fatal("lab data must be present")
	}
	if got := result.LabData.Categories["agentic-browsing"].ScoreDisplayMode; got != "fraction" {
		t.Errorf("agentic score display mode = %q, want fraction", got)
	}
	if got := result.LabData.Metrics["serverResponseTime"].Value; got == nil || *got != 420 {
		t.Errorf("server response time = %v, want 420 ms", got)
	}
}

func TestParseResult_Lighthouse134Fixture_PreservesInsightsAndAuditDetails(t *testing.T) {
	t.Parallel()

	raw := loadPSIFixture(t)
	result := parseResult("https://example.test/page", "mobile", raw)
	lab := result.LabData
	if lab == nil {
		t.Fatal("lab data must be present")
	}

	insight := findAudit(t, lab.Insights, "render-blocking-insight")
	if got := insight.MetricSavings["LCP"]; got != 710 {
		t.Errorf("LCP metric savings = %v, want 710", got)
	}

	var details struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(insight.Details, &details); err != nil {
		t.Fatalf("unmarshal insight details: %v", err)
	}
	if details.Type != "table" {
		t.Errorf("insight detail type = %q, want table", details.Type)
	}

	findAudit(t, lab.Diagnostics, "uses-text-compression")
	assertContains(t, lab.PassedAuditIDs, "llms-txt")
	assertContains(t, lab.ManualAuditIDs, "manual-audit")

	encoded, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal normalized result: %v", err)
	}
	if !json.Valid(encoded) {
		t.Fatal("normalized result must be valid JSON")
	}
	var output map[string]any
	if err := json.Unmarshal(encoded, &output); err != nil {
		t.Fatalf("unmarshal normalized output: %v", err)
	}
}

func TestParseResult_WithoutLighthouseOrFieldData_PreservesRequestMetadata(t *testing.T) {
	t.Parallel()

	result := parseResult("https://example.test", "desktop", &apiResponse{})

	if result.Metadata.InputURL != "https://example.test" {
		t.Errorf("InputURL = %q, want request URL", result.Metadata.InputURL)
	}
	if result.Metadata.Strategy != "desktop" {
		t.Errorf("Strategy = %q, want desktop", result.Metadata.Strategy)
	}
	if result.FieldData != nil {
		t.Error("FieldData must be nil when PSI returns no field data")
	}
	if result.LabData != nil {
		t.Error("LabData must be nil when PSI returns no Lighthouse result")
	}
}

func TestNewClientNotNil(t *testing.T) {
	t.Parallel()

	if client := NewClient("fake-key"); client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func loadPSIFixture(t *testing.T) *apiResponse {
	t.Helper()

	path := filepath.Join("..", "..", "..", "testdata", "psi-lighthouse-13.4.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read PSI fixture: %v", err)
	}

	var raw apiResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal PSI fixture: %v", err)
	}
	return &raw
}

func findAudit(t *testing.T, audits []LighthouseAudit, id string) LighthouseAudit {
	t.Helper()

	for _, audit := range audits {
		if audit.ID == id {
			return audit
		}
	}
	t.Fatalf("audit %q not found", id)
	return LighthouseAudit{}
}

func assertContains(t *testing.T, values []string, want string) {
	t.Helper()

	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Errorf("%q not found in %v", want, values)
}
