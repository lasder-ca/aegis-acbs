# Aegis ACBS v0.1.0

`v0.1.0` is the first public release of Aegis Coupled-Bound Search, an exact bidirectional shortest-path research implementation for weighted directed road graphs.

## Included in this release

- The adaptive `aegis` search and the `aegis-static` scheduler ablation.
- OSM XML, DIMACS, and Aegis binary-graph import support.
- Route, benchmark, stress, diagnosis, isolated replay, trigger profiling, and matrix aggregation commands.
- JSON, CSV, and self-contained HTML reports.
- Linux amd64/arm64, Windows amd64, and macOS amd64 builds.
- SHA-256 checksums, source archives, a Git bundle, and a CycloneDX SBOM.

## Published evidence

The repository includes the raw artifacts from a 10,000-query Tokyo travel-time graph experiment:

- 10,000/10,000 shortest-path distances matched Dijkstra.
- 11 queries crossed the predefined meaningful-slowdown threshold in the initial suite.
- Isolated replay retained one adaptive-scheduler tail and one persistent classical tail; nine cases did not reproduce.
- Three connection-guard candidates were rejected by the predefined acceptance gate.
- Whole-suite profiling found one same-suite checkpoint rule matching the scheduler tail with no observed false positives.

The checkpoint rule remains diagnostic. It was discovered and evaluated on the same dataset and is not part of the default scheduler.

See [Tokyo evidence](docs/TOKYO_EVIDENCE.md) and `research/tokyo-time-2026-07-18/`.

## Validation and release controls

- CI runs on Linux, Windows, and macOS.
- Go 1.23 compatibility and the current Go toolchain are tested.
- Linux race detection, `go vet`, formatting, cross-build, shell syntax, and Python syntax checks are included.
- Release publication verifies the imported Tokyo artifacts before creating a tag or uploading assets.

## Scope

This release is intended for reproducible review and independent testing. It does not establish universal performance superiority, universal correctness over every possible graph, academic novelty, or generalization of the Tokyo diagnostic trigger.
