# Aegis ACBS v0.4.0-experimental

v0.4.0 keeps the v0.2/v0.3 ACBS search policy fixed and strengthens experimental methodology. It does not claim a new speed improvement caused by an algorithm change.

## Main changes

- Interleave all algorithms inside each repeated measurement. The order is deterministically shuffled from the benchmark seed, query index, and repeat index.
- Report arithmetic mean, best, worst, p50, p95, and p99 query latency.
- Add `--measure-memory` for a separate untimed allocation pass, so allocation instrumentation does not contaminate the primary latency sample.
- Record median allocation bytes/objects per algorithm, process peak RSS, Go heap allocation, heap reservation, cumulative allocation, and GC count.
- Replace the ambiguous meeting counter with three explicit values: `connectionChecks`, `finiteMeetings`, and `upperBoundUpdates`. The v0.3 `meetingChecks` JSON field is retained as an alias for finite overlaps.
- Add `scripts/validate-research.sh` for 1,000 queries over ten seeds and four Japanese urban regions.
- Add `scripts/memory-profile.sh` for ACBS-only `/usr/bin/time -v` validation.

## Interpretation

Peak RSS is process-wide and includes the graph, benchmark harness, Go runtime, and selected algorithms. Per-query `allocBytes` and `allocObjects` are measured in a separate untimed pass. For an ACBS-only process footprint, run `scripts/memory-profile.sh`.

## Validation

- Differential correctness against Dijkstra.
- Deterministic interleaving tests.
- Descriptive-statistics ordering checks.
- Connection accounting invariant: `connectionChecks >= finiteMeetings >= upperBoundUpdates`.
- JSON, CSV, benchmark HTML, and matrix HTML generation.
- `go test ./...`, `go vet ./...`, and race detector checks.

## Research status

Novelty remains unconfirmed. MM, NBS, DVCBS, BAE*, lower-bound propagation, and recent exact bidirectional termination criteria still require independent comparison.
