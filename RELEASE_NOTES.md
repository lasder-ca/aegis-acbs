# Aegis ACBS v0.8.0-experimental

This release adds query-level regret diagnosis without changing the v0.7 ACBS search policy.

## Added

- `aegis diagnose` for benchmark JSON reports.
- Absolute slowdown (`algorithm - fastest classical`) alongside runtime ratios.
- A default meaningful-slowdown rule: ratio >= 1.25 and absolute penalty >= 1 ms.
- Top queries ranked independently by absolute penalty and ratio.
- Query features for diagnosis: distance ratio, endpoint degrees, forward share, switch rate, first-upper-bound fraction, stale-pop rate, and efficiency imbalance.
- Pearson correlations between absolute slowdown and recorded query/search features.
- Self-contained diagnostic JSON, CSV, and HTML reports.
- Source/target degree, distance ratio, and graph diameter in new benchmark reports.
- Standard benchmark output now reports meaningful slowdown count and p50/p95/max absolute penalty.

## Deliberately not changed

The production `aegis` scheduler remains `edge-efficiency-v3`. A conservative time-metric tail guard was prototyped, but it made synthetic time-road p50 and p95 slower, so it was rejected rather than shipped.

## Interpretation

A high ratio on a microsecond query is not necessarily a practical regression. The diagnostic report separates ratio-only noise from slowdowns with material absolute latency.
