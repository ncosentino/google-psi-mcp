// Package pagespeed provides a client and normalized response models for the
// Google PageSpeed Insights API v5.
package pagespeed

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

// AnalysisResult contains one PageSpeed Insights analysis for a URL and strategy.
type AnalysisResult struct {
	// Metadata identifies the request and the Lighthouse execution that produced it.
	Metadata AnalysisMetadata `json:"metadata"`
	// FieldData contains real-user Chrome UX Report data when it is available.
	FieldData *FieldData `json:"fieldData,omitempty"`
	// LabData contains the synthetic Lighthouse result when it is available.
	LabData *LabData `json:"labData,omitempty"`
}

// AnalysisMetadata describes the source and timing of a PageSpeed Insights result.
type AnalysisMetadata struct {
	// InputURL is the URL supplied by the MCP caller.
	InputURL string `json:"inputUrl"`
	// Strategy is the requested Lighthouse form factor.
	Strategy string `json:"strategy"`
	// AnalysisTimestamp is the PSI analysis timestamp.
	AnalysisTimestamp *time.Time `json:"analysisTimestamp,omitempty"`
	// FetchTime is the time at which Lighthouse fetched the page.
	FetchTime *time.Time `json:"fetchTime,omitempty"`
	// LighthouseVersion is the Lighthouse engine version used by PSI.
	LighthouseVersion string `json:"lighthouseVersion,omitempty"`
	// RequestedURL is the URL Lighthouse received after PSI request processing.
	RequestedURL string `json:"requestedUrl,omitempty"`
	// FinalURL is the final resolved URL after redirects.
	FinalURL string `json:"finalUrl,omitempty"`
	// FinalDisplayedURL is the URL Lighthouse selected for report display.
	FinalDisplayedURL string `json:"finalDisplayedUrl,omitempty"`
	// MainDocumentURL is the final main-document request URL.
	MainDocumentURL string `json:"mainDocumentUrl,omitempty"`
	// RunWarnings contains warnings emitted by Lighthouse during the run.
	RunWarnings []string `json:"runWarnings"`
	// RuntimeError identifies a fatal Lighthouse runtime failure.
	RuntimeError *RuntimeError `json:"runtimeError,omitempty"`
}

// RuntimeError describes a Lighthouse failure that can invalidate the lab result.
type RuntimeError struct {
	// Code is the stable Lighthouse error identifier.
	Code string `json:"code"`
	// Message is the human-readable failure description.
	Message string `json:"message"`
}

// FieldData contains page-level and origin-level real-user measurements.
type FieldData struct {
	// Page contains URL-level measurements or an explicitly marked origin fallback.
	Page *FieldExperience `json:"page,omitempty"`
	// Origin contains measurements aggregated across the resolved origin.
	Origin *FieldExperience `json:"origin,omitempty"`
}

// FieldExperience describes one PSI Chrome UX Report loading experience.
type FieldExperience struct {
	// ID is the URL or origin represented by the measurements.
	ID string `json:"id"`
	// InitialURL is the URL originally requested from the field-data service.
	InitialURL string `json:"initialUrl,omitempty"`
	// OverallRating is the normalized overall field-data rating.
	OverallRating string `json:"overallRating,omitempty"`
	// OriginFallback reports whether page data was replaced with origin data.
	OriginFallback bool `json:"originFallback"`
	// Metrics contains available field metrics keyed by stable friendly names.
	Metrics map[string]FieldMetric `json:"metrics"`
}

// FieldMetric contains a p75 real-user metric and its distribution.
type FieldMetric struct {
	// ID is the upstream PSI metric identifier.
	ID string `json:"id"`
	// Percentile identifies the percentile represented by Value.
	Percentile int `json:"percentile"`
	// Value is the normalized p75 metric value.
	Value float64 `json:"value"`
	// Unit identifies the unit used by Value and distribution boundaries.
	Unit string `json:"unit,omitempty"`
	// Rating is good, needs-improvement, poor, or unavailable.
	Rating string `json:"rating"`
	// Distributions contains the proportions in the upstream rating buckets.
	Distributions []FieldDistribution `json:"distributions"`
}

// FieldDistribution contains one real-user metric histogram bucket.
type FieldDistribution struct {
	// Min is the inclusive lower bucket boundary when supplied by PSI.
	Min *float64 `json:"min,omitempty"`
	// Max is the exclusive upper bucket boundary when supplied by PSI.
	Max *float64 `json:"max,omitempty"`
	// Proportion is the fraction of experiences in the bucket.
	Proportion float64 `json:"proportion"`
}

// LabData contains normalized Lighthouse category, metric, and audit results.
type LabData struct {
	// Categories contains every Lighthouse category returned by PSI.
	Categories map[string]CategoryResult `json:"categories"`
	// Metrics contains the primary Lighthouse lab metrics.
	Metrics map[string]LabMetric `json:"metrics"`
	// Insights contains actionable Lighthouse 13 insight audits.
	Insights []LighthouseAudit `json:"insights"`
	// Diagnostics contains failed non-insight audits.
	Diagnostics []LighthouseAudit `json:"diagnostics"`
	// UnscoredAudits contains informative audits without a numeric score.
	UnscoredAudits []LighthouseAudit `json:"unscoredAudits"`
	// PassedAuditIDs contains passed non-metric audit identifiers.
	PassedAuditIDs []string `json:"passedAuditIds"`
	// NotApplicableAuditIDs contains audits that did not apply to the page.
	NotApplicableAuditIDs []string `json:"notApplicableAuditIds"`
	// ManualAuditIDs contains audits requiring human verification.
	ManualAuditIDs []string `json:"manualAuditIds"`
	// Entities contains Lighthouse first-party and third-party classifications.
	Entities []Entity `json:"entities"`
}

// CategoryResult contains one Lighthouse category result.
type CategoryResult struct {
	// ID is the Lighthouse category identifier.
	ID string `json:"id"`
	// Title is the human-readable category title.
	Title string `json:"title"`
	// Description explains the category's purpose.
	Description string `json:"description,omitempty"`
	// Score is the category score when Lighthouse provides one.
	Score *float64 `json:"score,omitempty"`
	// ScoreDisplayMode identifies gauge or fractional score rendering.
	ScoreDisplayMode string `json:"scoreDisplayMode,omitempty"`
}

// LabMetric contains one synthetic Lighthouse metric result.
type LabMetric struct {
	// ID is the Lighthouse audit identifier backing the metric.
	ID string `json:"id"`
	// Title is the human-readable metric title.
	Title string `json:"title"`
	// Description explains the metric.
	Description string `json:"description,omitempty"`
	// Score is the Lighthouse metric score when available.
	Score *float64 `json:"score,omitempty"`
	// Value is the raw Lighthouse numeric value.
	Value *float64 `json:"value,omitempty"`
	// Unit identifies the upstream numeric value unit.
	Unit string `json:"unit,omitempty"`
	// DisplayValue is the localized value rendered by Lighthouse.
	DisplayValue string `json:"displayValue,omitempty"`
}

// LighthouseAudit contains an actionable or informative Lighthouse audit.
type LighthouseAudit struct {
	// ID is the stable Lighthouse audit identifier.
	ID string `json:"id"`
	// Title is the human-readable audit title.
	Title string `json:"title"`
	// Description explains what the audit measures.
	Description string `json:"description,omitempty"`
	// Score is the audit score when Lighthouse supplies one.
	Score *float64 `json:"score,omitempty"`
	// ScoreDisplayMode describes how Lighthouse renders the score.
	ScoreDisplayMode string `json:"scoreDisplayMode,omitempty"`
	// DisplayValue is the localized summary rendered by Lighthouse.
	DisplayValue string `json:"displayValue,omitempty"`
	// Explanation provides additional failure context.
	Explanation string `json:"explanation,omitempty"`
	// ErrorMessage contains an audit execution error.
	ErrorMessage string `json:"errorMessage,omitempty"`
	// Warnings contains audit-specific warnings.
	Warnings []string `json:"warnings"`
	// NumericValue is the audit's raw numeric value when available.
	NumericValue *float64 `json:"numericValue,omitempty"`
	// NumericUnit identifies the unit of NumericValue.
	NumericUnit string `json:"numericUnit,omitempty"`
	// MetricSavings contains estimated improvements to Lighthouse metrics.
	MetricSavings map[string]float64 `json:"metricSavings,omitempty"`
	// Details preserves the structured Lighthouse audit details object.
	Details json.RawMessage `json:"details,omitempty"`
}

// Entity describes a first-party or third-party entity identified by Lighthouse.
type Entity struct {
	// Name is the entity's display name.
	Name string `json:"name"`
	// Category is the entity classification when available.
	Category string `json:"category,omitempty"`
	// Homepage is the entity homepage when available.
	Homepage string `json:"homepage,omitempty"`
	// Origins contains origins associated with the entity.
	Origins []string `json:"origins"`
	// IsFirstParty reports whether Lighthouse classified the entity as first party.
	IsFirstParty bool `json:"isFirstParty"`
	// IsUnrecognized reports whether Lighthouse could not identify the entity.
	IsUnrecognized bool `json:"isUnrecognized"`
}

type apiResponse struct {
	ID                      string                `json:"id"`
	AnalysisUTCTimestamp    string                `json:"analysisUTCTimestamp"`
	LoadingExperience       *rawLoadingExperience `json:"loadingExperience"`
	OriginLoadingExperience *rawLoadingExperience `json:"originLoadingExperience"`
	LighthouseResult        *rawLighthouseResult  `json:"lighthouseResult"`
}

type rawLoadingExperience struct {
	ID              string                    `json:"id"`
	InitialURL      string                    `json:"initial_url"`
	OverallCategory string                    `json:"overall_category"`
	OriginFallback  bool                      `json:"origin_fallback"`
	Metrics         map[string]rawFieldMetric `json:"metrics"`
}

type rawFieldMetric struct {
	Percentile    float64                `json:"percentile"`
	Distributions []rawFieldDistribution `json:"distributions"`
	Category      string                 `json:"category"`
}

type rawFieldDistribution struct {
	Min        *float64 `json:"min"`
	Max        *float64 `json:"max"`
	Proportion float64  `json:"proportion"`
}

type rawLighthouseResult struct {
	RequestedURL      string                  `json:"requestedUrl"`
	FinalURL          string                  `json:"finalUrl"`
	FinalDisplayedURL string                  `json:"finalDisplayedUrl"`
	MainDocumentURL   string                  `json:"mainDocumentUrl"`
	LighthouseVersion string                  `json:"lighthouseVersion"`
	FetchTime         string                  `json:"fetchTime"`
	RunWarnings       []json.RawMessage       `json:"runWarnings"`
	RuntimeError      *RuntimeError           `json:"runtimeError"`
	Categories        map[string]*rawCategory `json:"categories"`
	Audits            map[string]*rawAudit    `json:"audits"`
	Entities          []Entity                `json:"entities"`
}

type rawCategory struct {
	ID                       string        `json:"id"`
	Title                    string        `json:"title"`
	Description              string        `json:"description"`
	Score                    *float64      `json:"score"`
	CategoryScoreDisplayMode string        `json:"categoryScoreDisplayMode"`
	AuditRefs                []rawAuditRef `json:"auditRefs"`
}

type rawAuditRef struct {
	ID    string `json:"id"`
	Group string `json:"group"`
}

type rawAudit struct {
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	Description      string             `json:"description"`
	Score            *float64           `json:"score"`
	ScoreDisplayMode string             `json:"scoreDisplayMode"`
	DisplayValue     string             `json:"displayValue"`
	Explanation      string             `json:"explanation"`
	ErrorMessage     string             `json:"errorMessage"`
	Warnings         []json.RawMessage  `json:"warnings"`
	NumericValue     *float64           `json:"numericValue"`
	NumericUnit      string             `json:"numericUnit"`
	MetricSavings    map[string]float64 `json:"metricSavings"`
	Details          json.RawMessage    `json:"details"`
}

type fieldMetricDefinition struct {
	name  string
	unit  string
	scale float64
}

var fieldMetricDefinitions = map[string]fieldMetricDefinition{
	"CUMULATIVE_LAYOUT_SHIFT_SCORE":   {name: "cls", scale: 0.01},
	"EXPERIMENTAL_TIME_TO_FIRST_BYTE": {name: "ttfb", unit: "ms", scale: 1},
	"FIRST_CONTENTFUL_PAINT_MS":       {name: "fcp", unit: "ms", scale: 1},
	"INTERACTION_TO_NEXT_PAINT":       {name: "inp", unit: "ms", scale: 1},
	"LARGEST_CONTENTFUL_PAINT_MS":     {name: "lcp", unit: "ms", scale: 1},
}

var labMetricNames = map[string]string{
	"first-contentful-paint":   "fcp",
	"largest-contentful-paint": "lcp",
	"cumulative-layout-shift":  "cls",
	"total-blocking-time":      "tbt",
	"server-response-time":     "serverResponseTime",
	"speed-index":              "speedIndex",
}

func parseResult(targetURL, strategy string, raw *apiResponse) *AnalysisResult {
	result := &AnalysisResult{
		Metadata: AnalysisMetadata{
			InputURL:          targetURL,
			Strategy:          strategy,
			AnalysisTimestamp: parseTimestamp(raw.AnalysisUTCTimestamp),
		},
		FieldData: parseFieldData(raw.LoadingExperience, raw.OriginLoadingExperience),
	}

	if raw.LighthouseResult == nil {
		return result
	}

	lhr := raw.LighthouseResult
	result.Metadata.FetchTime = parseTimestamp(lhr.FetchTime)
	result.Metadata.LighthouseVersion = lhr.LighthouseVersion
	result.Metadata.RequestedURL = lhr.RequestedURL
	result.Metadata.FinalURL = lhr.FinalURL
	result.Metadata.FinalDisplayedURL = lhr.FinalDisplayedURL
	result.Metadata.MainDocumentURL = lhr.MainDocumentURL
	result.Metadata.RunWarnings = normalizeJSONMessages(lhr.RunWarnings)
	result.Metadata.RuntimeError = lhr.RuntimeError
	result.LabData = parseLabData(lhr)
	return result
}

func parseTimestamp(value string) *time.Time {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil
	}
	return &parsed
}

func parseFieldData(page, origin *rawLoadingExperience) *FieldData {
	if page == nil && origin == nil {
		return nil
	}
	return &FieldData{
		Page:   parseFieldExperience(page),
		Origin: parseFieldExperience(origin),
	}
}

func parseFieldExperience(raw *rawLoadingExperience) *FieldExperience {
	if raw == nil {
		return nil
	}

	metrics := make(map[string]FieldMetric, len(raw.Metrics))
	for id, metric := range raw.Metrics {
		definition, ok := fieldMetricDefinitions[id]
		if !ok {
			definition = fieldMetricDefinition{
				name:  strings.ToLower(id),
				scale: 1,
			}
		}

		distributions := make([]FieldDistribution, 0, len(metric.Distributions))
		for _, distribution := range metric.Distributions {
			distributions = append(distributions, FieldDistribution{
				Min:        scaledPointer(distribution.Min, definition.scale),
				Max:        scaledPointer(distribution.Max, definition.scale),
				Proportion: distribution.Proportion,
			})
		}

		metrics[definition.name] = FieldMetric{
			ID:            id,
			Percentile:    75,
			Value:         metric.Percentile * definition.scale,
			Unit:          definition.unit,
			Rating:        normalizeRating(metric.Category),
			Distributions: distributions,
		}
	}

	return &FieldExperience{
		ID:             raw.ID,
		InitialURL:     raw.InitialURL,
		OverallRating:  normalizeRating(raw.OverallCategory),
		OriginFallback: raw.OriginFallback,
		Metrics:        metrics,
	}
}

func scaledPointer(value *float64, scale float64) *float64 {
	if value == nil {
		return nil
	}
	scaled := *value * scale
	return &scaled
}

func normalizeRating(category string) string {
	switch strings.ToUpper(category) {
	case "FAST":
		return "good"
	case "AVERAGE":
		return "needs-improvement"
	case "SLOW":
		return "poor"
	default:
		return "unavailable"
	}
}

func parseLabData(raw *rawLighthouseResult) *LabData {
	data := &LabData{
		Categories:            make(map[string]CategoryResult, len(raw.Categories)),
		Metrics:               make(map[string]LabMetric, len(labMetricNames)),
		Insights:              []LighthouseAudit{},
		Diagnostics:           []LighthouseAudit{},
		UnscoredAudits:        []LighthouseAudit{},
		PassedAuditIDs:        []string{},
		NotApplicableAuditIDs: []string{},
		ManualAuditIDs:        []string{},
		Entities:              append([]Entity(nil), raw.Entities...),
	}
	if data.Entities == nil {
		data.Entities = []Entity{}
	}

	insightIDs := make(map[string]struct{})
	for id, category := range raw.Categories {
		if category == nil {
			continue
		}
		categoryID := category.ID
		if categoryID == "" {
			categoryID = id
		}
		data.Categories[id] = CategoryResult{
			ID:               categoryID,
			Title:            category.Title,
			Description:      category.Description,
			Score:            category.Score,
			ScoreDisplayMode: normalizeCategoryScoreDisplayMode(category.CategoryScoreDisplayMode),
		}
		for _, ref := range category.AuditRefs {
			if strings.EqualFold(ref.Group, "insights") {
				insightIDs[ref.ID] = struct{}{}
			}
		}
	}

	for id, friendlyName := range labMetricNames {
		audit, ok := raw.Audits[id]
		if !ok || audit == nil {
			continue
		}
		data.Metrics[friendlyName] = LabMetric{
			ID:           id,
			Title:        audit.Title,
			Description:  audit.Description,
			Score:        audit.Score,
			Value:        audit.NumericValue,
			Unit:         audit.NumericUnit,
			DisplayValue: audit.DisplayValue,
		}
	}

	for id, rawAudit := range raw.Audits {
		if rawAudit == nil {
			continue
		}
		if _, isMetric := labMetricNames[id]; isMetric {
			continue
		}

		displayMode := strings.ToLower(rawAudit.ScoreDisplayMode)
		switch displayMode {
		case "notapplicable":
			data.NotApplicableAuditIDs = append(data.NotApplicableAuditIDs, id)
			continue
		case "manual":
			data.ManualAuditIDs = append(data.ManualAuditIDs, id)
			continue
		}

		if rawAudit.Score != nil && *rawAudit.Score >= 0.9 {
			data.PassedAuditIDs = append(data.PassedAuditIDs, id)
			continue
		}

		audit := normalizeAudit(id, rawAudit)
		_, groupedAsInsight := insightIDs[id]
		if groupedAsInsight || strings.HasSuffix(id, "-insight") {
			data.Insights = append(data.Insights, audit)
			continue
		}
		if rawAudit.Score == nil {
			data.UnscoredAudits = append(data.UnscoredAudits, audit)
			continue
		}
		data.Diagnostics = append(data.Diagnostics, audit)
	}

	sort.Strings(data.PassedAuditIDs)
	sort.Strings(data.NotApplicableAuditIDs)
	sort.Strings(data.ManualAuditIDs)
	sort.Slice(data.Insights, func(i, j int) bool { return data.Insights[i].ID < data.Insights[j].ID })
	sort.Slice(data.Diagnostics, func(i, j int) bool { return data.Diagnostics[i].ID < data.Diagnostics[j].ID })
	sort.Slice(data.UnscoredAudits, func(i, j int) bool { return data.UnscoredAudits[i].ID < data.UnscoredAudits[j].ID })
	return data
}

func normalizeCategoryScoreDisplayMode(value string) string {
	value = strings.TrimPrefix(value, "CATEGORY_SCORE_DISPLAY_MODE_")
	return strings.ToLower(value)
}

func normalizeAudit(id string, raw *rawAudit) LighthouseAudit {
	auditID := raw.ID
	if auditID == "" {
		auditID = id
	}

	var details json.RawMessage
	trimmedDetails := bytes.TrimSpace(raw.Details)
	if len(trimmedDetails) > 0 && !bytes.Equal(trimmedDetails, []byte("null")) {
		details = append(json.RawMessage(nil), trimmedDetails...)
	}

	return LighthouseAudit{
		ID:               auditID,
		Title:            raw.Title,
		Description:      raw.Description,
		Score:            raw.Score,
		ScoreDisplayMode: raw.ScoreDisplayMode,
		DisplayValue:     raw.DisplayValue,
		Explanation:      raw.Explanation,
		ErrorMessage:     raw.ErrorMessage,
		Warnings:         normalizeJSONMessages(raw.Warnings),
		NumericValue:     raw.NumericValue,
		NumericUnit:      raw.NumericUnit,
		MetricSavings:    raw.MetricSavings,
		Details:          details,
	}
}

func normalizeJSONMessages(values []json.RawMessage) []string {
	if len(values) == 0 {
		return []string{}
	}

	messages := make([]string, 0, len(values))
	for _, value := range values {
		var text string
		if err := json.Unmarshal(value, &text); err == nil {
			messages = append(messages, text)
			continue
		}
		messages = append(messages, string(value))
	}
	return messages
}
