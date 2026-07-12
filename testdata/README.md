# API response fixtures

These sanitized fixtures represent the Google PageSpeed Insights API v5 and
Chrome UX Report API response shapes used by the Go and C# test suites.

- `psi-lighthouse-13.4.json` covers PSI field data, Lighthouse 13.4 metadata,
  categories, metrics, insights, diagnostics, and unavailable audits.
- `crux-current.json` covers current CrUX histograms, percentiles, fractions,
  collection periods, and URL normalization.
- `crux-history.json` covers CrUX timeseries values, including unavailable
  periods represented upstream as `"NaN"` or `null`.

The values and URLs are synthetic. No credentials or private response data are
stored in the repository.
