# Benchmarking ACBS

## Principles

Dijkstra is the correctness reference and is run outside timed samples. Starting in v0.4, algorithms are interleaved inside every repetition. A deterministic shuffle derived from the seed, query index, and repeat index changes the order while preserving exact reproducibility.

```text
for each query:
  for each repeat:
    shuffle(all algorithms)
    measure every algorithm once in that order
  retain the median repetition for each algorithm
```

Use `--order rotated` only when comparing directly with v0.3 methodology.

## Latency

Reports contain query-level:

- `meanNs`
- `minNs` (best)
- `maxNs` (worst)
- `medianNs` (p50)
- `p95Ns`
- `p99Ns`

Each query-level value is the median of its repeated inner measurements. Aggregate percentiles are then computed across queries.

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

Peak RSS includes the graph, all algorithms used by the process, Go runtime, and report data. For an ACBS-only process measurement:

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

The default research suite uses Tokyo, Yokohama, Osaka, and Nagoya, both distance and time graphs, ten seeds, and 1,000 queries per seed:

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

v0.5 adds a search-core benchmark that warms the pooled workspace before measuring ACBS. Run:

```bash
go test ./internal/search -run '^$' -bench '^BenchmarkACBSLargeGrid$' -benchmem
```

For a direct v0.4/v0.5 comparison using identical temporary benchmark code and isolated Git worktrees:

```bash
scripts/compare-allocations.sh
```

The generated grid fixture isolates queue and path allocation behavior. It does not replace the real OSM-derived city matrix for algorithm-performance claims. Queue backing arrays are retained after warm-up, so lower `B/op` and `allocs/op` should be evaluated together with ACBS-only peak RSS.

## Experimental pruning and dual-potential comparison

The default research set only isolates the scheduler. Add experimental mechanisms explicitly:

```bash
aegis benchmark \
  --graph city.aegis \
  --algorithms dijkstra,bidijkstra,astar,aegis-static,aegis,aegis-prune,aegis-projection \
  --order interleaved \
  --measure-memory
```

Do not merge the two ACBS variants into a selector for publication results. Report each variant's latency and work counters separately. A projection speedup accompanied by substantially higher expansion counts should be described as an implementation tradeoff rather than a stronger heuristic.

## Concurrent and soak validation

Use the in-process stress runner to exercise pooled workspaces under real goroutine concurrency:

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

Report throughput together with p95/p99 and peak RSS. A throughput increase accompanied by exploding p99 or RSS is not considered a successful scaling result.

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

The ratio threshold alone is insufficient on very short queries. By default a query is marked meaningful only when ACBS is at least 1.25x slower than the fastest classical baseline **and** loses at least 1 ms in absolute time.


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

Report both the observed event rate and its 95% interval. When zero events are observed, also report the exact one-sided upper bound `1 - 0.05^(1/N)`. At N=10,000 this is approximately 0.02995%, not proof that the true rate is zero.
