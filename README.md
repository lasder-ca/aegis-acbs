# Aegis ACBS

[![CI](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml/badge.svg)](https://github.com/lasder-ca/aegis-acbs/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Aegis Coupled-Bound Search (ACBS)** is an experimental exact shortest-path search for weighted directed road graphs. It maintains forward and backward frontiers in one search, uses a shared admissible lower bound for termination, and adaptively allocates edge-work between directions.

ACBS is not a portfolio that chooses among A*, bidirectional Dijkstra, and Dijkstra. The default `aegis` algorithm performs one coupled bidirectional search and returns a shortest path only after its lower bound reaches a valid incumbent upper bound.

> **Research preview.** Independent novelty review is incomplete. This repository does not claim that ACBS is academically novel or fastest for every query.

日本語版: [README.ja.md](README.ja.md)

## Current evidence

Tokyo time-weighted road graph, user-run on July 18, 2026:

| Check | Result |
|---|---:|
| Exactness against Dijkstra | **10,000 / 10,000** |
| Initially detected meaningful slowdowns | 11 / 10,000 |
| Reproduced in isolated replay | 2 / 11 |
| Reproduced adaptive-scheduler tail | 1 / 10,000 |
| Reproduced persistent classical tail | 1 / 10,000 |
| Guard candidates accepted | **0 / 3** |
| Narrow diagnostic trigger | checkpoint 48, one match |

The narrow trigger was discovered and evaluated on the same Tokyo suite. It is retained as a diagnostic hypothesis, not promoted into the default scheduler. See [Tokyo evidence and limitations](docs/TOKYO_EVIDENCE.md).

## Algorithm outline

1. Maintain forward and reverse priority queues.
2. Construct a balanced potential from an admissible geographic lower bound.
3. Search both reduced-cost graphs with nonnegative edge weights.
4. Maintain a coupled lower bound from the two frontier minima.
5. Maintain the best complete path as an upper bound.
6. Adapt edge-work chunks using observed lower-bound progress and frontier work.
7. Stop only when the proven lower bound reaches the incumbent upper bound.

The adaptive scheduler changes exploration order, not the exactness criterion.

## Quick start

Requirements: Go 1.23 or newer.

```bash
go test ./...
go build -o bin/aegis ./cmd/aegis

bin/aegis import-osm \
  --input benchdata/hatfield-uk.osm \
  --output /tmp/hatfield-distance.aegis \
  --profile car \
  --metric distance

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

## Algorithms

Normal comparisons:

- `dijkstra`: correctness oracle.
- `bidijkstra`: bidirectional Dijkstra.
- `astar`: geographic A* when an admissible cost-per-meter bound exists.
- `aegis`: default ACBS with adaptive scheduling.
- `aegis-static`: ACBS scheduler ablation.

Rejected or diagnostic variants remain available under explicit experimental use so failed experiments stay reproducible. They are not recommended production choices.

## Reproducibility commands

```bash
# Multi-seed meaningful-tail validation
scripts/validate-tail.sh path/to/time-graph.aegis artifacts/tail

# Isolated replay of detected tails
bin/aegis replay-regret \
  --graph path/to/time-graph.aegis \
  --validation artifacts/tail/regret-validation.json \
  --input-root artifacts/tail \
  --runs 31 --warmup 5 \
  --output artifacts/replay.json \
  --csv artifacts/replay.csv \
  --html artifacts/replay.html

# Whole-suite checkpoint profiling
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

See [Benchmarking](docs/BENCHMARKING.md), [Correctness](docs/CORRECTNESS.md), [Algorithm](docs/ALGORITHM.md), and [Related work](docs/RELATED_WORK.md).

## Known limitations

- No guarantee of being fastest for every source-target pair.
- The Tokyo trigger rule has no independent city or seed validation yet.
- The public evidence currently emphasizes one large Tokyo graph and a small bundled Hatfield fixture.
- No contraction hierarchies, landmarks, or graph-specific preprocessing are used in the baseline comparison.
- Research novelty remains unverified by independent reviewers.

## Release status

`v0.12.0-research-preview` is the first public research preview. It keeps the v0.11.2 default search unchanged and publishes successful and failed experiments together.

## License

MIT. See [LICENSE](LICENSE).
