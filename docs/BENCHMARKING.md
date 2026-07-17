# Benchmarking ACBS

## Principles

The benchmark harness treats Dijkstra as the correctness reference and measures each candidate on the same source/target pairs. Correctness reference runs are outside timed samples.

For each query and algorithm, the harness executes an odd number of repeated measurements and retains the median. Algorithm order rotates by query index to reduce systematic cache and thermal bias. Large graphs default to batch size one.

## Latency

Reports contain:

- `medianNs` (p50)
- `p95Ns`
- `p99Ns`

These are percentiles across query-level median measurements, not percentiles of every inner repetition.

## Speedup definitions

The following values are intentionally separate:

```text
ratio of medians = median(Dijkstra query times) / median(candidate query times)

per-query speedup(q) = Dijkstra time(q) / candidate time(q)

median per-query speedup = median(per-query speedup)

geometric-mean speedup = exp(mean(log(per-query speedup)))
```

They can differ and must not share one label.

## Fastest-baseline comparison

For each query:

```text
fastest baseline = min(Dijkstra, bidirectional Dijkstra, A*)

relative runtime = ACBS time / fastest baseline time

oracle regret = max(1, relative runtime)
```

Relative runtime can be below 1 when ACBS is faster than every baseline. Oracle regret cannot be below 1.

## Work counters

- `expanded`: states whose outgoing or incoming adjacency was processed.
- `relaxed`: edges examined, including edges rejected by overflow or an incumbent bound.
- `queuePushes`: priority-queue insertions.
- `queuePops`: valid and stale heap removals.
- `stalePops`: removed entries superseded by a better distance or already settled.
- `prunedAtPop`: states rejected by `g+h >= U` after leaving the queue.
- `prunedAtRelax`: candidate labels rejected by `g+h >= U` before queue insertion.
- `boundPruned`: `prunedAtPop + prunedAtRelax`.
- `meetingChecks`: finite forward/backward labels considered for an incumbent.
- `upperBoundUpdates`: meeting checks that improved the incumbent.

`relaxed` can remain unchanged while pruning improves runtime because relaxation-time pruning prevents queue insertion and later expansion. The split counters make that effect visible.

## Query classes

The mixed suite rotates through:

- `local`: nearby candidate selected from a sampled pool.
- `random`: uniform source/target selection inside the query pool.
- `regional`: distant candidate selected from a sampled pool.

The default pool is the largest strongly connected component.

## Multi-seed matrix

Run all `.aegis` graphs in a directory over five seeds:

```bash
AEGIS_GRAPH_DIR=.data/regional-graphs \
AEGIS_REPORT_DIR=artifacts/matrix \
scripts/benchmark-matrix.sh
```

The script creates individual JSON/HTML reports and then runs:

```bash
bin/aegis aggregate \
  --input-dir artifacts/matrix \
  --output artifacts/matrix/benchmark-matrix.json \
  --csv artifacts/matrix/benchmark-matrix.csv \
  --html artifacts/matrix/benchmark-matrix.html
```

The aggregate report includes median and worst p95 values across seeds. No single seed is sufficient evidence of a stable advantage.

## Reproducibility controls

Recommended settings for large road graphs:

```bash
GOMAXPROCS=1
--queries 50
--repeats 3
--batch 1
--pair-mode strongly-connected
--suite mixed
--research
```

Increase query and repeat counts for publication-grade results. Record CPU model, governor, temperature policy, OS, Go version, graph checksum, graph import options, and raw JSON files.
