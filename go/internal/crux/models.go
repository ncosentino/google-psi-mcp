// Package crux provides normalized current and historical Chrome UX Report data.
package crux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// Result contains current Chrome UX Report data for one URL or origin.
type Result struct {
	// Target is the normalized URL or origin represented by the record.
	Target string `json:"target"`
	// TargetType is url or origin.
	TargetType string `json:"targetType"`
	// FormFactor is phone, tablet, desktop, or empty for all form factors.
	FormFactor string `json:"formFactor,omitempty"`
	// CollectionPeriod is the 28-day aggregation window.
	CollectionPeriod CollectionPeriod `json:"collectionPeriod"`
	// Metrics contains every metric returned by CrUX.
	Metrics map[string]Metric `json:"metrics"`
	// URLNormalization describes URL normalization performed by CrUX.
	URLNormalization *URLNormalization `json:"urlNormalization,omitempty"`
}

// HistoryResult contains historical Chrome UX Report timeseries data.
type HistoryResult struct {
	// Target is the normalized URL or origin represented by the record.
	Target string `json:"target"`
	// TargetType is url or origin.
	TargetType string `json:"targetType"`
	// FormFactor is phone, tablet, desktop, or empty for all form factors.
	FormFactor string `json:"formFactor,omitempty"`
	// CollectionPeriods contains the ordered 28-day aggregation windows.
	CollectionPeriods []CollectionPeriod `json:"collectionPeriods"`
	// Metrics contains every metric timeseries returned by CrUX.
	Metrics map[string]HistoryMetric `json:"metrics"`
}

// CollectionPeriod identifies one CrUX aggregation window.
type CollectionPeriod struct {
	// FirstDate is the first day in the aggregation window.
	FirstDate Date `json:"firstDate"`
	// LastDate is the final day in the aggregation window.
	LastDate Date `json:"lastDate"`
}

// Date is a calendar date returned by CrUX.
type Date struct {
	// Year is the four-digit year.
	Year int `json:"year"`
	// Month is the one-based month.
	Month int `json:"month"`
	// Day is the one-based day of month.
	Day int `json:"day"`
}

// URLNormalization describes how CrUX normalized a requested URL.
type URLNormalization struct {
	// OriginalURL is the URL supplied in the request.
	OriginalURL string `json:"originalUrl"`
	// NormalizedURL is the URL used for the lookup.
	NormalizedURL string `json:"normalizedUrl"`
}

// Metric contains the available current statistical aggregations for one metric.
type Metric struct {
	// Histogram contains metric distribution buckets.
	Histogram []HistogramBin `json:"histogram"`
	// P75 is the 75th-percentile value when the metric supports percentiles.
	P75 *float64 `json:"p75,omitempty"`
	// Fractions contains labeled proportions for categorical metrics.
	Fractions map[string]float64 `json:"fractions,omitempty"`
}

// HistogramBin contains one current CrUX histogram bucket.
type HistogramBin struct {
	// Start is the inclusive lower boundary when supplied by CrUX.
	Start *float64 `json:"start,omitempty"`
	// End is the exclusive upper boundary when supplied by CrUX.
	End *float64 `json:"end,omitempty"`
	// Density is the fraction of experiences in the bucket.
	Density float64 `json:"density"`
}

// HistoryMetric contains historical statistical aggregations for one metric.
type HistoryMetric struct {
	// Histogram contains metric distribution timeseries.
	Histogram []HistoryHistogramBin `json:"histogram"`
	// P75 contains ordered 75th-percentile values; unavailable periods are null.
	P75 []*float64 `json:"p75,omitempty"`
	// Fractions contains ordered labeled proportions; unavailable periods are null.
	Fractions map[string][]*float64 `json:"fractions,omitempty"`
}

// HistoryHistogramBin contains one CrUX histogram bucket across collection periods.
type HistoryHistogramBin struct {
	// Start is the inclusive lower boundary when supplied by CrUX.
	Start *float64 `json:"start,omitempty"`
	// End is the exclusive upper boundary when supplied by CrUX.
	End *float64 `json:"end,omitempty"`
	// Densities contains ordered proportions; unavailable periods are null.
	Densities []*float64 `json:"densities"`
}

type rawResponse struct {
	Record                  *rawRecord        `json:"record"`
	URLNormalizationDetails *URLNormalization `json:"urlNormalizationDetails"`
}

type rawRecord struct {
	Key              rawKey               `json:"key"`
	Metrics          map[string]rawMetric `json:"metrics"`
	CollectionPeriod CollectionPeriod     `json:"collectionPeriod"`
}

type rawHistoryResponse struct {
	Record *rawHistoryRecord `json:"record"`
}

type rawHistoryRecord struct {
	Key               rawKey                      `json:"key"`
	Metrics           map[string]rawHistoryMetric `json:"metrics"`
	CollectionPeriods []CollectionPeriod          `json:"collectionPeriods"`
}

type rawKey struct {
	URL        string `json:"url"`
	Origin     string `json:"origin"`
	FormFactor string `json:"formFactor"`
}

type rawMetric struct {
	Histogram   []rawHistogramBin  `json:"histogram"`
	Percentiles *rawPercentiles    `json:"percentiles"`
	Fractions   map[string]float64 `json:"fractions"`
}

type rawHistogramBin struct {
	Start   numberValue `json:"start"`
	End     numberValue `json:"end"`
	Density float64     `json:"density"`
}

type rawPercentiles struct {
	P75 numberValue `json:"p75"`
}

type rawHistoryMetric struct {
	HistogramTimeseries   []rawHistoryHistogramBin     `json:"histogramTimeseries"`
	PercentilesTimeseries *rawPercentilesTimeseries    `json:"percentilesTimeseries"`
	FractionTimeseries    map[string]rawFractionSeries `json:"fractionTimeseries"`
}

type rawHistoryHistogramBin struct {
	Start     numberValue   `json:"start"`
	End       numberValue   `json:"end"`
	Densities []numberValue `json:"densities"`
}

type rawPercentilesTimeseries struct {
	P75 []numberValue `json:"p75s"`
}

type rawFractionSeries struct {
	Fractions []numberValue `json:"fractions"`
}

type numberValue struct {
	value *float64
}

func (v *numberValue) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("null")) || bytes.Equal(data, []byte(`"NaN"`)) {
		v.value = nil
		return nil
	}

	var number float64
	if len(data) > 0 && data[0] == '"' {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return fmt.Errorf("parsing numeric string: %w", err)
		}
		parsed, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return fmt.Errorf("parsing numeric string %q: %w", text, err)
		}
		number = parsed
	} else if err := json.Unmarshal(data, &number); err != nil {
		return fmt.Errorf("parsing number: %w", err)
	}

	v.value = &number
	return nil
}

func parseCurrent(raw *rawResponse) *Result {
	if raw == nil || raw.Record == nil {
		return nil
	}

	target, targetType := raw.Record.Key.target()
	metrics := make(map[string]Metric, len(raw.Record.Metrics))
	for id, metric := range raw.Record.Metrics {
		histogram := make([]HistogramBin, 0, len(metric.Histogram))
		for _, bin := range metric.Histogram {
			histogram = append(histogram, HistogramBin{
				Start:   cloneNumber(bin.Start.value),
				End:     cloneNumber(bin.End.value),
				Density: bin.Density,
			})
		}

		var p75 *float64
		if metric.Percentiles != nil {
			p75 = cloneNumber(metric.Percentiles.P75.value)
		}
		metrics[id] = Metric{
			Histogram: histogram,
			P75:       p75,
			Fractions: metric.Fractions,
		}
	}

	return &Result{
		Target:           target,
		TargetType:       targetType,
		FormFactor:       normalizeFormFactor(raw.Record.Key.FormFactor),
		CollectionPeriod: raw.Record.CollectionPeriod,
		Metrics:          metrics,
		URLNormalization: raw.URLNormalizationDetails,
	}
}

func parseHistory(raw *rawHistoryResponse) *HistoryResult {
	if raw == nil || raw.Record == nil {
		return nil
	}

	target, targetType := raw.Record.Key.target()
	metrics := make(map[string]HistoryMetric, len(raw.Record.Metrics))
	for id, metric := range raw.Record.Metrics {
		histogram := make([]HistoryHistogramBin, 0, len(metric.HistogramTimeseries))
		for _, bin := range metric.HistogramTimeseries {
			histogram = append(histogram, HistoryHistogramBin{
				Start:     cloneNumber(bin.Start.value),
				End:       cloneNumber(bin.End.value),
				Densities: cloneNumbers(bin.Densities),
			})
		}

		var p75 []*float64
		if metric.PercentilesTimeseries != nil {
			p75 = cloneNumbers(metric.PercentilesTimeseries.P75)
		}
		fractions := make(map[string][]*float64, len(metric.FractionTimeseries))
		for label, series := range metric.FractionTimeseries {
			fractions[label] = cloneNumbers(series.Fractions)
		}
		if len(fractions) == 0 {
			fractions = nil
		}

		metrics[id] = HistoryMetric{
			Histogram: histogram,
			P75:       p75,
			Fractions: fractions,
		}
	}

	return &HistoryResult{
		Target:            target,
		TargetType:        targetType,
		FormFactor:        normalizeFormFactor(raw.Record.Key.FormFactor),
		CollectionPeriods: raw.Record.CollectionPeriods,
		Metrics:           metrics,
	}
}

func (k rawKey) target() (string, string) {
	if k.URL != "" {
		return k.URL, "url"
	}
	return k.Origin, "origin"
}

func normalizeFormFactor(value string) string {
	switch value {
	case "PHONE":
		return "phone"
	case "TABLET":
		return "tablet"
	case "DESKTOP":
		return "desktop"
	default:
		return ""
	}
}

func cloneNumber(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneNumbers(values []numberValue) []*float64 {
	if len(values) == 0 {
		return nil
	}
	result := make([]*float64, len(values))
	for index, value := range values {
		result[index] = cloneNumber(value.value)
	}
	return result
}
