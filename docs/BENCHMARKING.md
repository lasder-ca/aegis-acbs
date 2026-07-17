# Benchmark methodology

## Fairness rules

- No contraction hierarchy, landmarks, routing tables, or algorithm-specific preprocessing.
- Every algorithm receives the same immutable graph and deterministic query list.
- Import, largest strongly connected component selection, and nearest-node lookup are outside query timings.
- Dijkstra correctness-reference runs are unmeasured.
- Algorithm order rotates per query.
- `aegis-race` is excluded from default comparisons because it uses two CPU cores.

## Measurements

```text
queries × repeats × batch size
```

Each query/algorithm pair is measured an odd number of times. Each measurement may execute a batch, and the per-execution average is recorded. The median repeated measurement becomes the query sample.

Automatic batch sizes:

| Graph size | Batch |
|---:|---:|
| < 1,000 | 64 |
| < 10,000 | 16 |
| < 100,000 | 4 |
| ≥ 100,000 | 1 |

## Query suites

- `local`: nearest candidate among a deterministic sample
- `random`: uniform pair
- `regional`: farthest candidate among a deterministic sample
- `mixed`: local, random, regional in rotation

The default `strongly-connected` pair mode selects the largest strongly connected component before timing.

## Raw ACBS metrics

- wall-clock duration
- expanded vertices
- relaxed edges
- queue pushes
- forward expanded vertices
- backward expanded vertices
- direction switches
- scheduler chunks
- expansion count when the first complete upper bound was found
- termination lower bound
- forward/backward bound-progress efficiency
- exact path distance and path length

## Summary metrics

- p50 and p95 latency
- median expanded vertices and relaxed edges
- speedup versus Dijkstra
- relaxed-edge reduction versus Dijkstra
- local/random/regional summaries
- ACBS forward/backward balance
- direction balance by distance class
- runtime regret versus the fastest established baseline
- first-upper-bound fraction
- termination-bound tightness

```text
runtime regret = ACBS runtime / min(Dijkstra, bidirectional Dijkstra, A*) runtime
```

Runtime regret is diagnostic, not an oracle used by ACBS.

## Visual report

The generated HTML is self-contained and works offline. It includes Japanese, English, Simplified Chinese, Korean, and French.

```bash
aegis benchmark \
  --graph city.aegis \
  --queries 300 \
  --repeats 7 \
  --output artifacts/city.json \
  --html artifacts/city.html
```

## Reproducibility

Record the commit, version, OS/architecture, Go version, CPU count, source-data SHA-256, graph dimensions, metric, profile, query suite, pair mode, seed, repeats, batch size, and algorithm list. OSM extracts with different hashes are different datasets.
