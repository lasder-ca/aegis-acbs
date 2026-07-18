# Aegis ACBS

[![CI](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml/badge.svg)](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Aegis Coupled-Bound Search (ACBS)** is an exact point-to-point shortest-path implementation for weighted directed road graphs.

ACBS advances forward and reverse frontiers under one shared proof of optimality. A balanced admissible potential guides both directions, while an adaptive scheduler shifts edge-work toward the frontier that is making more lower-bound progress. The search returns only after the proven lower bound reaches the best complete path found so far.

> **Research status:** ACBS is a reproducible research prototype. Its relationship to prior bidirectional-search work is documented, but academic novelty and performance generalization have not been independently established.

日本語: [README.ja.md](README.ja.md)

## Highlights

- **Exact routing:** returns a shortest path on finite, non-negative weighted directed graphs.
- **Proof-oriented output:** reports the incumbent upper bound, termination lower bound, and optimality gap.
- **Adaptive bidirectional work:** allocates edge-processing chunks using observed lower-bound progress.
- **Road-graph tooling:** imports OSM XML and DIMACS data, then stores a compact binary graph format.
- **Reproducible evaluation:** emits JSON, CSV, and self-contained HTML reports for benchmarks, tail replay, and trigger profiling.
- **Portable CLI:** tested on Linux, Windows, and macOS.

## Current evidence

The first public release includes a user-run Tokyo travel-time graph experiment from July 18, 2026. The graph contained 611,846 nodes and 1,235,323 directed edges.

| Check | Observed result |
|---|---:|
| Shortest-path distance matched Dijkstra | **10,000 / 10,000** |
| Initially detected meaningful slowdowns | 11 / 10,000 |
| Slowdowns reproduced in isolated replay | 2 / 11 |
| Reproduced adaptive-scheduler tail | 1 / 10,000 |
| Reproduced persistent classical tail | 1 / 10,000 |
| Guard candidates accepted by the predefined gate | **0 / 3** |
| In-sample diagnostic trigger | checkpoint 48, one match |

These results describe one graph, one workload design, and one machine environment. They are evidence for that experiment, not a universal speed claim. Raw reports and the acceptance criteria are documented in [Tokyo evidence](docs/TOKYO_EVIDENCE.md).

## Quick start

Requirements: Go 1.23 or newer.

```bash
git clone https://github.com/lasder-ca/aegis-acbs.git
cd aegis-acbs

go test ./...
go build -o bin/aegis ./cmd/aegis
```

Import the bundled OSM fixture:

```bash
bin/aegis import-osm \
  --input benchdata/hatfield-uk.osm \
  --output /tmp/hatfield-distance.aegis \
  --profile car \
  --metric distance
```

Run an interleaved benchmark:

```bash
bin/aegis benchmark \
  --graph /tmp/hatfield-distance.aegis \
  --queries 1000 \
  --repeats 9 \
  --order interleaved \
  --measure-memory \
  --suite mixed \
  --seed 1010 \
  --output /tmp/hatfield.json \
  --html /tmp/hatfield.html
```

## How the search is organized

```text
source  ->  forward frontier  ->  candidate connection
                                            ^
target  <-  reverse frontier  <-  candidate connection

shared state: admissible lower bound L, incumbent path cost U
termination: L >= U
```

The scheduler changes which frontier receives the next edge-work chunk. It does not change the admissible potential, the incumbent path, the coupled lower bound, or the exact stopping condition.

A formal description is available in [Algorithm](docs/ALGORITHM.md), with the correctness argument separated into [Correctness](docs/CORRECTNESS.md).

## Main commands

| Command | Purpose |
|---|---|
| `import-osm` | Import OSM XML into an Aegis graph |
| `import-dimacs` | Import DIMACS shortest-path data |
| `route` | Compute one route |
| `benchmark` | Compare algorithms with interleaved repeated timing |
| `stress` | Run concurrent routing with sampled Dijkstra verification |
| `diagnose` | Find meaningful per-query performance tails |
| `replay-regret` | Remeasure retained tails in isolation |
| `profile-trigger` | Record deterministic scheduler features at checkpoints |
| `aggregate` | Build multi-seed benchmark matrices |

The normal benchmark set includes Dijkstra, bidirectional Dijkstra, geographic A*, static ACBS, and adaptive ACBS. Rejected experiments remain available only for reproducibility and are described in the changelog and research documents.

## Reproducing the research workflow

```bash
# Multi-seed tail validation
scripts/validate-tail.sh path/to/time-graph.aegis artifacts/tail

# Isolated replay of retained cases
bin/aegis replay-regret \
  --graph path/to/time-graph.aegis \
  --validation artifacts/tail/regret-validation.json \
  --input-root artifacts/tail \
  --runs 31 \
  --warmup 5 \
  --output artifacts/replay.json \
  --csv artifacts/replay.csv \
  --html artifacts/replay.html

# Whole-suite scheduler-feature profiling
bin/aegis profile-trigger \
  --graph path/to/time-graph.aegis \
  --validation artifacts/tail/regret-validation.json \
  --replay artifacts/replay.json \
  --input-root artifacts/tail \
  --checkpoints 24,32,40,48 \
  --max-matches 5 \
  --output artifacts/trigger-profile.json \
  --csv artifacts/trigger-profile.csv \
  --html artifacts/trigger-profile.html
```

## Documentation

- [Algorithm](docs/ALGORITHM.md)
- [Correctness](docs/CORRECTNESS.md)
- [Benchmark methodology](docs/BENCHMARKING.md)
- [Tokyo evidence](docs/TOKYO_EVIDENCE.md)
- [Related work](docs/RELATED_WORK.md)
- [Data formats](docs/DATA.md)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## Limitations

- Performance varies by graph, metric, route length, and hardware.
- The public large-graph evidence currently centers on one Tokyo travel-time graph.
- The checkpoint-48 trigger was discovered and evaluated on the same suite, so it remains diagnostic only.
- ACBS does not currently use contraction hierarchies, landmarks, or graph-specific preprocessing.
- Independent novelty review and broader third-party reproduction are still needed.

## Release status

`v0.1.0` is the first public research release. Public semantic versioning starts there; earlier version numbers in the changelog refer to private research iterations.

## License

MIT. See [LICENSE](LICENSE).
