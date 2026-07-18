# Tokyo time-graph evidence

This document records the user-run Tokyo time-weighted road-graph experiments used for the first public research preview. The graph contained 611,846 nodes and 1,235,323 directed edges.

## Scope and provenance

- Environment: WSL Ubuntu, single-thread benchmark runs unless otherwise stated.
- Date: July 18, 2026.
- Query suite: deterministic mixed, strongly connected source-target pairs.
- Correctness oracle: Dijkstra distance equality.
- Raw JSON/CSV/HTML files are imported before public release by `scripts/import-tokyo-evidence.sh`.
- `research/tokyo-time-2026-07-18/observed-summary.json` is a transcription of console results, not a replacement for raw artifacts.

## 10,000-query tail validation

- Correct shortest paths: **10,000 / 10,000**.
- Meaningful slowdowns under the predeclared threshold: **11 / 10,000 (0.110000%)**.
- Wilson 95% interval: **0.061435% to 0.196880%**.
- Overall p95 absolute penalty: **0.075 ms**.
- Maximum observed absolute penalty: **14.469 ms**; this maximum did not necessarily satisfy the meaningful ratio threshold.

A meaningful slowdown required both:

- runtime at least 1.25 times the fastest classical baseline, and
- at least 1 ms of absolute penalty.

## Isolated replay

The 11 detected cases were warmed up and repeatedly measured in isolation.

- Not reproduced: **9**.
- Reproduced adaptive-scheduler tail: **1**.
- Reproduced persistent classical tail: **1**.
- Correctness failures: **0**.

The scheduler-tail case was seed `1010`, query `877`, class `random`. The persistent classical case was seed `271828182`, query `279`, class `local`.

## Rejected guard experiments

The acceptance gate was declared before the final measurements: at least 0.5 ms scheduler-tail improvement, less than 1 ms persistent-tail regression, 100% correctness, and no more than 1% regression in global latency or search work.

| Candidate | Scheduler-tail gain | Mean ratio | Median ratio | p95 ratio | Relaxed ratio | Expanded ratio | Result |
|---|---:|---:|---:|---:|---:|---:|---|
| `aegis-connect-32` | 0.778 ms | 1.1595x | 1.0430x | 1.1498x | 1.0492x | 1.0472x | Rejected |
| `aegis-connect-40` | 0.732 ms | 1.1498x | 1.0210x | 1.1552x | 1.0252x | 1.0294x | Rejected |
| `aegis-connect-32x16` | 0.283 ms | 0.9996x | 0.9948x | 0.9934x | 1.0123x | 1.0246x | Rejected |

The gate was not relaxed after seeing the results. The default `aegis` scheduler was retained.

## Whole-suite trigger profiling

All 10,000 queries were traced at chunks 24, 32, 40, and 48. The replay-confirmed scheduler tail was the positive label.

Selected diagnostic rule:

```text
checkpoint = 48
switchRate <= 0.458333333
```

Observed on the same Tokyo suite:

- Matches: **1**.
- Positive matches: **1**.
- False positives: **0**.
- Trace errors: **0**.
- Unstable replay labels: **0**.

This is an in-sample result. It may be a real separator, an accidental property of one query, or a form of overfitting. It is not used by the default algorithm and must be evaluated on independent cities and seeds before any promotion.

## Claims supported by this evidence

Supported:

- ACBS returned the same shortest-path distances as Dijkstra for these 10,000 Tokyo queries.
- Most initially detected tail events did not survive isolated replay.
- One reproducible adaptive-scheduler tail and one reproducible persistent classical tail remained.
- Broad connection guards were not justified by the predeclared gate.

Not supported:

- Universal correctness over all graphs.
- Universal speed superiority.
- Academic novelty.
- Generalization of the selected checkpoint rule.
