# Benchmarking ACBS

## Principles

Dijkstra is the correctness reference and is run outside timed samples. Algorithms are interleaved inside every repetition. A deterministic shuffle derived from the seed, query index, and repeat index changes the order while preserving reproducibility.

```text
for each query:
  for each repeat:
    shuffle(all algorithms)
    measure every algorithm once in that order
  retain the median repetition for each algorithm
```

Use `--order rotated` only when comparing directly with the earlier rotated-order methodology.

## Latency

Reports contain query-level:

- `meanNs`
- `minNs` (best)
- `maxNs` (worst)
- `medianNs` (p50)
- `p95Ns`
- `p99Ns`

Each query-level value is the median of its repeated inner measurements. Aggregate percentiles are then computed across queries.

Very small synthetic fixtures can approach the resolution of the platform timer. Publication-scale latency claims should use realistic graphs, repeated measurements, and absolute-time thresholds rather than ratios alone.

## Memory

`--measure-memory` adds a separate untimed execution for every query and algorithm and records:

- `allocBytes`: delta of Go `TotalAlloc`.
- `allocObjects`: delta of Go `Mallocs`.

The untimed pass prevents memory instrumentation from contaminating latency. Allocation values include query-context and harness allocations surrounding the algorithm call.

The report also records process-wide:

- `peakRssBytes`
- `goHeapAllocBytes`
- `goHeapSysBytes`
- `goTotalAllocBytes`
- `goMallocs`
- `goNumGc`

Peak RSS includes the graph, all algorithms used by the process, the Go runtime, and report data. Linux and macOS report peak RSS directly; unsupported platforms currently record `peakRssBytes` as `0` while the Go heap and allocation counters remain available. For an ACBS-only process measurement:

```bash
scripts/memory-profile.sh path/to/graph.aegis
```

## Speedup definitions

```text
ratio of medians = median(Dijkstra query times) / median(candidate query times)
per-query speedup(q) = Dijkstra time(q) / candidate time(q)
median per-query speedup = median(per-query speedup)
geometric-mean speedup = exp(mean(log(per-query speedup)))
```

## Fastest classical-baseline comparison

```text
fastest classical baseline = min(Dijkstra, bidirectional Dijkstra, A*)
runtime ratio = ACBS time / fastest classical baseline time
classical oracle regret = max(1, runtime ratio)
```

## Work and connection counters

- `expanded`: adjacency lists processed.
- `relaxed`: edges examined.
- `queuePushes`, `queuePops`, `stalePops`: priority-queue activity.
- `prunedAtPop`, `prunedAtRelax`, `boundPruned`: incumbent-bound pruning.
- `connectionChecks`: every attempted forward/backward connection check.
- `finiteMeetings`: checks where both directional labels are finite.
- `meetingChecks`: compatibility alias for `finiteMeetings`.
- `upperBoundUpdates`: finite meetings that improved the incumbent.

The accounting invariant is:

```text
connectionChecks >= finiteMeetings >= upperBoundUpdates
boundPruned = prunedAtPop + prunedAtRelax
```

## Publication-scale validation

The research matrix supports Tokyo, Yokohama, Osaka, and Nagoya, with distance and travel-time graphs, multiple seeds, and independently stored reports:

```bash
scripts/validate-research.sh
```

Recommended environment controls:

```bash
GOMAXPROCS=1
AEGIS_ORDER=interleaved
AEGIS_QUERIES=1000
AEGIS_REPEATS=3
AEGIS_MEASURE_MEMORY=1
```

Record CPU model, governor, temperature policy, OS, Go version, graph checksum, import options, raw JSON, and command line. For p99 claims, increase the query count beyond 1,000 and report confidence intervals.

## Steady-state allocation regression

The search-core benchmark warms the pooled workspace before measuring ACBS:

```bash
go test ./internal/search -run '^$' -bench '^BenchmarkACBSLargeGrid$' -benchmem
```

For a direct comparison between two revisions using identical temporary benchmark code and isolated Git worktrees:

```bash
scripts/compare-allocations.sh
```

The generated grid fixture isolates queue and path allocation behavior. It does not replace the real OSM-derived city matrix for algorithm-performance claims. Queue backing arrays are retained after warm-up, so lower `B/op` and `allocs/op` should be evaluated together with ACBS-only peak RSS.

## Experimental variants

The default research set isolates the scheduler. Other mechanisms can be added explicitly:

```bash
aegis benchmark \
  --graph city.aegis \
  --algorithms dijkstra,bidijkstra,astar,aegis-static,aegis,aegis-prune,aegis-projection \
  --order interleaved \
  --measure-memory
```

Publication reports list each ACBS variant separately and exclude selector results from the main comparison. A projection speedup accompanied by substantially higher expansion counts is an implementation tradeoff rather than evidence of a stronger heuristic.

## Concurrent and soak validation

Use the in-process stress runner to exercise pooled workspaces under goroutine concurrency:

```bash
GOMAXPROCS=8 aegis stress \
  --graph city.aegis --queries 10000 --workers 8 \
  --verify-every 100 --output stress.json
```

`verify-every=1` verifies every query against Dijkstra. Larger values reduce reference-search cost while retaining deterministic sampling. A zero value disables Dijkstra verification but still validates returned path continuity.

Worker scaling and repeated-process soak helpers:

```bash
scripts/stress-matrix.sh city.aegis artifacts/stress
scripts/soak.sh city.aegis artifacts/soak
```

Report throughput together with p95, p99, and peak RSS. A throughput increase accompanied by substantially worse p99 or RSS should be treated as a tradeoff rather than a clean scaling improvement.

## Regret diagnosis

Use `aegis diagnose` after a benchmark containing `dijkstra`, `bidijkstra`, `astar`, and `aegis` samples:

```bash
aegis diagnose \
  --input benchmark.json \
  --ratio-threshold 1.25 \
  --penalty-floor 1ms \
  --output regret.json \
  --csv regret.csv \
  --html regret.html
```

The ratio threshold alone is insufficient on very short queries. By default a query is marked meaningful only when ACBS is at least 1.25 times slower than the fastest classical baseline and loses at least 1 ms in absolute time.

## Multi-seed meaningful-slowdown validation

Run resumable validation over a graph:

```bash
AEGIS_QUERIES=1000 \
AEGIS_SEEDS="1010 20260717 424242 8675309 123456789 314159265 271828182 161803398 141421356 173205080" \
scripts/validate-tail.sh city-time.aegis artifacts/tail-validation
```

The workflow writes one benchmark report per seed, then runs:

```bash
aegis validate-regret \
  --input-dir artifacts/tail-validation \
  --ratio-threshold 1.25 \
  --penalty-floor 1ms \
  --min-queries 10000 \
  --max-meaningful-rate 0
```

Report both the observed event rate and its 95% interval. When zero events are observed, also report the exact one-sided upper bound `1 - 0.05^(1/N)`. At `N=10,000` this is approximately `0.02995%`, not proof that the true rate is zero.

## Isolated tail replay

Large validation matrices can retain a small set of meaningful slowdowns in `regret-validation.json`. A retained timed sample is first replayed in isolation before it is used to propose a scheduler change:

```bash
aegis replay-regret \
  --graph city-time.aegis \
  --validation validation/regret-validation.json \
  --input-root validation \
  --runs 31 \
  --warmup 5 \
  --output validation/regret-replay.json \
  --csv validation/regret-replay.csv \
  --html validation/regret-replay.html
```

Each case is measured with Dijkstra, bidirectional Dijkstra, A*, static ACBS, and adaptive ACBS in a rotated interleaved order. Timed runs do not record traces. A separate untimed ACBS run records one event per scheduler chunk.

Replay classifications are evidence for analysis:

- `not-reproduced`: the validation outlier did not survive repeated isolated measurement.
- `adaptive-scheduler-tail`: static ACBS materially beat adaptive ACBS under the configured absolute floor.
- `persistent-classical-tail`: a classical method remained faster, but static scheduling did not explain the difference.

Repeated adaptive-scheduler tails with a shared trace pattern can support a narrowly defined scheduler experiment. Any proposed change is then evaluated against a predefined whole-suite gate.
