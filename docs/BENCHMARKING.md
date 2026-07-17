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

## Fastest-baseline comparison

```text
fastest baseline = min(Dijkstra, bidirectional Dijkstra, A*)
relative runtime = ACBS time / fastest baseline time
oracle regret = max(1, relative runtime)
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
