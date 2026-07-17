# Aegis ACBS v0.10.0-experimental

This release turns the v0.9 large-sample validation result into deterministic, isolated outlier replays. The production ACBS search policy is unchanged.

## Added

- `aegis replay-regret` reads the meaningful cases retained in a `regret-validation.json` report.
- Every case is remeasured with Dijkstra, bidirectional Dijkstra, A*, static ACBS, and adaptive ACBS using repeated interleaved execution.
- Cases are classified as:
  - `not-reproduced`: the original tail does not survive isolated remeasurement;
  - `adaptive-scheduler-tail`: static ACBS materially beats adaptive ACBS;
  - `persistent-classical-tail`: a classical method remains materially faster but static scheduling does not explain it.
- One untimed ACBS run captures a chunk-level trace containing direction, edge budget, lower-bound gain, normalized work, frontier sizes, efficiency scores, and upper-bound discovery.
- JSON, CSV, and self-contained HTML reports are generated.

## Deliberately not changed

The production scheduler was not tuned against eleven observed Tokyo outliers before those cases were isolated and shown to reproduce. This avoids regressing the other 9,989 queries to fix timing noise or unrelated special cases.

## Recommended workflow

1. Run `validate-regret` over a large matrix.
2. Run `replay-regret` against its `regret-validation.json`.
3. Only modify the scheduler if multiple persistent cases are classified as `adaptive-scheduler-tail` and share the same trace pattern.
