// Package pagespeed provides types for the Google PageSpeed Insights API v5.
package pagespeed

import (
	"math"
	"time"
)

// Result is a parsed PageSpeed Insights analysis for a single URL and strategy.
type Result struct {
	URL         string    `json:"url"`
	Strategy    string    `json:"strategy"`
	AnalyzedAt  time.Time `json:"analyzedAt"`
	Scores      Scores    `json:"scores"`
	CoreWebVitals CoreWebVitals `json:"coreWebVitals"`
	Opportunities []Opportunity `json:"opportunities"`
	FailingAudits []Audit       `json:"failingAudits"`
	PassedAuditIDs []string     `json:"passedAuditIds"`
}

// Scores holds category scores in the range 0-100.
type Scores struct {
	Performance   int `json:"performance"`
	SEO           int `json:"seo"`
	Accessibility int `json:"accessibility"`
	BestPractices int `json:"bestPractices"`
}

// CoreWebVitals holds the key performance metrics.
type CoreWebVitals struct {
	FCP        MetricValue `json:"fcp"`
	LCP        MetricValue `json:"lcp"`
	CLS        MetricValue `json:"cls"`
	TBT        MetricValue `json:"tbt"`
	TTFB       MetricValue `json:"ttfb"`
	SpeedIndex MetricValue `json:"speedIndex"`
}

// MetricValue holds a single metric reading with its rating.
type MetricValue struct {
	Value  float64 `json:"value"`
	Unit   string  `json:"unit,omitempty"`
	Rating string  `json:"rating"` // "good", "needs-improvement", or "poor"
}

// Opportunity is a performance improvement suggestion.
type Opportunity struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Savings     string `json:"savings,omitempty"`
	Impact      string `json:"impact"` // "high", "medium", or "low"
}

// Audit is a single Lighthouse audit result.
type Audit struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Score        float64 `json:"score"`
	DisplayValue string  `json:"displayValue,omitempty"`
}

// --- PSI API raw response types (for JSON unmarshalling) ---

type apiResponse struct {
	LighthouseResult *lighthouseResult `json:"lighthouseResult"`
}

type lighthouseResult struct {
	Categories map[string]*category `json:"categories"`
	Audits     map[string]*audit    `json:"audits"`
}

type category struct {
	Score float64 `json:"score"`
}

type audit struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Score        *float64 `json:"score"`
	NumericValue *float64 `json:"numericValue"`
	DisplayValue string   `json:"displayValue"`
	Details      *auditDetails `json:"details"`
}

type auditDetails struct {
	Type string `json:"type"`
}

func parseResult(targetURL, strategy string, raw *apiResponse) *Result {
	r := &Result{
		URL:        targetURL,
		Strategy:   strategy,
		AnalyzedAt: time.Now().UTC(),
	}

	if raw.LighthouseResult == nil {
		return r
	}

	lhr := raw.LighthouseResult
	r.Scores = parseScores(lhr.Categories)
	r.CoreWebVitals = parseCoreWebVitals(lhr.Audits)

	for id, a := range lhr.Audits {
		if a.Score == nil {
			continue
		}
		score := *a.Score
		if score >= 0.9 {
			r.PassedAuditIDs = append(r.PassedAuditIDs, id)
			continue
		}
		parsed := Audit{
			ID:           id,
			Title:        a.Title,
			Description:  a.Description,
			Score:        score,
			DisplayValue: a.DisplayValue,
		}
		if a.Details != nil && a.Details.Type == "opportunity" {
			r.Opportunities = append(r.Opportunities, Opportunity{
				ID:          id,
				Title:       a.Title,
				Description: a.Description,
				Savings:     a.DisplayValue,
				Impact:      impactFromScore(score),
			})
		} else {
			r.FailingAudits = append(r.FailingAudits, parsed)
		}
	}

	return r
}

func parseScores(cats map[string]*category) Scores {
	s := Scores{}
	if c, ok := cats["performance"]; ok {
		s.Performance = int(math.Round(c.Score * 100))
	}
	if c, ok := cats["seo"]; ok {
		s.SEO = int(math.Round(c.Score * 100))
	}
	if c, ok := cats["accessibility"]; ok {
		s.Accessibility = int(math.Round(c.Score * 100))
	}
	if c, ok := cats["best-practices"]; ok {
		s.BestPractices = int(math.Round(c.Score * 100))
	}
	return s
}

func parseCoreWebVitals(audits map[string]*audit) CoreWebVitals {
	return CoreWebVitals{
		FCP:        parseTimeMetric(audits, "first-contentful-paint", fcpRating),
		LCP:        parseTimeMetric(audits, "largest-contentful-paint", lcpRating),
		CLS:        parseCLSMetric(audits),
		TBT:        parseTimeMetric(audits, "total-blocking-time", tbtRating),
		TTFB:       parseTimeMetric(audits, "server-response-time", ttfbRating),
		SpeedIndex: parseTimeMetric(audits, "speed-index", siRating),
	}
}

func parseTimeMetric(audits map[string]*audit, id string, ratingFn func(float64) string) MetricValue {
	a, ok := audits[id]
	if !ok || a.NumericValue == nil {
		return MetricValue{}
	}
	secs := *a.NumericValue / 1000
	return MetricValue{Value: round2(secs), Unit: "s", Rating: ratingFn(secs)}
}

func parseCLSMetric(audits map[string]*audit) MetricValue {
	a, ok := audits["cumulative-layout-shift"]
	if !ok || a.NumericValue == nil {
		return MetricValue{}
	}
	v := round2(*a.NumericValue)
	return MetricValue{Value: v, Rating: clsRating(v)}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func impactFromScore(score float64) string {
	switch {
	case score < 0.5:
		return "high"
	case score < 0.75:
		return "medium"
	default:
		return "low"
	}
}

// CWV rating functions follow Google's standard thresholds.

func fcpRating(s float64) string {
	switch {
	case s < 1.8:
		return "good"
	case s < 3.0:
		return "needs-improvement"
	default:
		return "poor"
	}
}

func lcpRating(s float64) string {
	switch {
	case s < 2.5:
		return "good"
	case s < 4.0:
		return "needs-improvement"
	default:
		return "poor"
	}
}

func clsRating(v float64) string {
	switch {
	case v < 0.1:
		return "good"
	case v < 0.25:
		return "needs-improvement"
	default:
		return "poor"
	}
}

func tbtRating(ms float64) string {
	// TBT is stored in ms (we receive ms from the API and divide by 1000 in parseTimeMetric,
	// but for TBT thresholds we compare in seconds: 0.2s = 200ms, 0.6s = 600ms).
	switch {
	case ms < 0.2:
		return "good"
	case ms < 0.6:
		return "needs-improvement"
	default:
		return "poor"
	}
}

func ttfbRating(s float64) string {
	switch {
	case s < 0.8:
		return "good"
	case s < 1.8:
		return "needs-improvement"
	default:
		return "poor"
	}
}

func siRating(s float64) string {
	switch {
	case s < 3.4:
		return "good"
	case s < 5.8:
		return "needs-improvement"
	default:
		return "poor"
	}
}
