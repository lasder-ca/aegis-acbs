# Aegis ACBS v0.5.0-experimental

v0.5.0 keeps the exact ACBS search policy and termination condition unchanged. It removes the dominant steady-state allocation source from all priority-queue-based search algorithms and strengthens memory reproducibility.

## Main changes

- Replace `container/heap` with a specialized binary min-heap over the existing queue item type.
- Retain forward, backward, and single-frontier queue backing arrays inside pooled search workspaces.
- Reconstruct one-way and bidirectional paths with one exact-sized result allocation instead of repeatedly growing and copying slices.
- Add a heap-ordering regression test and a steady-state ACBS allocation ceiling test.
- Add `BenchmarkACBSLargeGrid` for repeatable search-core allocation measurements.
- Add `scripts/compare-allocations.sh`, which checks out v0.4 and v0.5 in isolated worktrees and generates JSON/Markdown allocation comparisons from the same generated road grid.
- Preserve the v0.4 interleaved methodology, memory telemetry, exactness checks, and Japan multi-city scripts.

## Allocation interpretation

The pooled queue capacity is retained by the process so subsequent queries can reuse it. This reduces cumulative allocation and GC pressure, but retained capacity contributes to process RSS. The remaining steady-state search-core allocation is normally the returned route path. Use `scripts/memory-profile.sh` with `--algorithms aegis` for an ACBS-only process footprint and the normal research matrix for end-to-end comparison.

## Correctness

The change is an implementation optimization, not a new search rule. Exhaustive small directed graphs, random directed time-road graphs, path validation, optimality certificates, race tests, and heap-ordering tests remain required before release.

## Research status

Novelty remains unconfirmed. Performance claims must continue to be limited to named datasets, hardware, compiler versions, query generators, seeds, and measurement settings.
