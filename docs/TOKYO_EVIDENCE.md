# Tokyo travel-time graph evidence

This document records the large-graph experiment included with the first public Aegis ACBS release.

## Dataset and environment

- Graph: Tokyo road network with travel-time weights.
- Size: 611,846 nodes and 1,235,323 directed edges.
- Environment: WSL Ubuntu; benchmark runs were single-threaded unless noted otherwise.
- Date: July 18, 2026.
- Query suite: deterministic mixed source-target pairs restricted to a strongly connected pool.
- Correctness reference: shortest-path distance returned by Dijkstra.
- Meaningful slowdown threshold: at least `1.25×` the fastest classical baseline and at least `1 ms` of absolute penalty.

Raw JSON, CSV, and HTML reports are stored in `research/tokyo-time-2026-07-18/`. The release importer validates their checksums and expected summary values before publication.

## 10,000-query validation

| Measure | Observed value |
|---|---:|
| Correct shortest-path distances | **10,000 / 10,000** |
| Meaningful slowdowns | **11 / 10,000 (0.110000%)** |
| Wilson 95% interval | 0.061435% to 0.196880% |
| Overall p95 absolute penalty | 0.075 ms |
| Maximum observed absolute penalty | 14.469 ms |

The maximum absolute penalty did not necessarily satisfy the ratio component of the meaningful-slowdown threshold.

## Isolated replay

The 11 retained queries were warmed up and repeatedly measured outside the original matrix.

| Classification | Count |
|---|---:|
| Not reproduced | 9 |
| Reproduced adaptive-scheduler tail | 1 |
| Reproduced persistent classical tail | 1 |
| Correctness failures | 0 |

The adaptive-scheduler case was seed `1010`, query `877`, class `random`. The persistent classical case was seed `271828182`, query `279`, class `local`.

This replay separates stable behavior from timing outliers. Nine of the eleven initial events did not survive isolation.

## Guard experiments

Before the final measurements, the acceptance gate required all of the following:

- at least `0.5 ms` improvement on the reproduced scheduler tail,
- less than `1 ms` regression on the persistent classical tail,
- complete shortest-path agreement,
- no more than `1%` regression in global mean, median, p95, relaxed edges, or expanded nodes.

| Candidate | Scheduler-tail gain | Mean ratio | Median ratio | p95 ratio | Relaxed ratio | Expanded ratio | Decision |
|---|---:|---:|---:|---:|---:|---:|---|
| `aegis-connect-32` | 0.778 ms | 1.1595× | 1.0430× | 1.1498× | 1.0492× | 1.0472× | Rejected |
| `aegis-connect-40` | 0.732 ms | 1.1498× | 1.0210× | 1.1552× | 1.0252× | 1.0294× | Rejected |
| `aegis-connect-32x16` | 0.283 ms | 0.9996× | 0.9948× | 0.9934× | 1.0123× | 1.0246× | Rejected |

The gate was kept unchanged after the results were known. The default adaptive scheduler was retained.

## Whole-suite trigger profiling

All 10,000 queries were traced at chunks 24, 32, 40, and 48. The replay-confirmed adaptive-scheduler case was used as the positive label.

The highest-ranked same-suite rule was:

```text
checkpoint = 48
switchRate <= 0.458333333
```

| Measure | Observed value |
|---|---:|
| Total matches | 1 |
| Positive matches | 1 |
| False positives | 0 |
| Trace errors | 0 |
| Unstable replay labels | 0 |

The same dataset was used to discover and evaluate this rule. It may represent a useful separator, a property unique to one query, or overfitting. The rule remains diagnostic and is not part of the default scheduler.

## Interpretation

The experiment supports these limited conclusions:

- ACBS matched Dijkstra's shortest-path distance on all 10,000 measured queries.
- Most initially detected latency tails were not stable under isolated replay.
- One stable scheduler-related tail and one stable classical-method advantage remained.
- Broad scheduler guards were not justified by the predefined acceptance gate.

The experiment does not establish universal correctness, universal speed superiority, academic novelty, or transfer of the checkpoint rule to other graphs and workloads.
