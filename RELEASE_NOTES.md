# Aegis ACBS v0.3.0-experimental

v0.3.0 freezes the v0.2 ACBS search policy and focuses on measurement integrity and reproducibility. The algorithm still runs one coupled bidirectional search; this release does not replace it with a portfolio selector.

## Main changes

- Separate `queuePushes`, `queuePops`, and `stalePops`.
- Separate upper-bound pruning at node pop from pruning during edge relaxation.
- Record meeting checks independently from successful upper-bound updates.
- Add p99 latency to JSON, CLI, HTML, and the browser UI.
- Replace ambiguous speedup output with:
  - ratio of medians versus Dijkstra,
  - median per-query speedup,
  - geometric-mean per-query speedup.
- Replace the incorrectly named sub-1.0 “runtime regret” with relative runtime to the fastest baseline.
- Add true oracle regret, clamped to at least 1.0.
- Add `aegis aggregate` for multi-seed, multi-graph JSON/CSV/HTML reports.
- Add reproducibility scripts for five seeds, distance/time metrics, and regional Japanese road graphs.
- Add Tokyo, Yokohama, Osaka, and Nagoya extraction definitions using Geofabrik PBF data.

## Validation

- Exhaustive directed graphs through four vertices.
- Random directed distance and time road graphs.
- Dijkstra distance, reachability, and path-continuity differential checks.
- Pruning counter identity: `boundPruned = prunedAtPop + prunedAtRelax`.
- Multi-seed report aggregation tests.
- Standalone benchmark and matrix HTML generation tests.
- `go test ./...`, `go vet ./...`, and race detector validation.

## Research status

Research novelty remains unconfirmed. MM, NBS, DVCBS, BAE*, lower-bound propagation, and newer exact bidirectional termination criteria must still be implemented or independently compared before making a novelty or state-of-the-art claim.
