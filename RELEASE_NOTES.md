# Aegis ACBS v0.12.0-research-preview

This is the first public research preview of Aegis Coupled-Bound Search.

The default `aegis` search is unchanged from v0.11.2. No rejected guard or in-sample trigger rule has been promoted into the scheduler.

## Public evidence

The release publishes both successful and failed Tokyo time-graph experiments:

- 10,000/10,000 query distances matched Dijkstra.
- 11 queries initially met the predeclared meaningful-slowdown threshold.
- Isolated replay retained one adaptive-scheduler tail and one persistent classical tail; nine cases did not reproduce.
- Three connection-guard candidates were rejected by a predeclared release gate.
- Whole-suite profiling found an in-sample checkpoint-48 rule matching the one scheduler tail with zero same-suite false positives.
- The rule remains diagnostic because it has not been validated on an independent city or seed suite.

See `docs/TOKYO_EVIDENCE.md` and `research/tokyo-time-2026-07-18/`.

## Repository hardening

- English primary README and Japanese README.
- CI on Linux, Windows, and macOS using the current Go toolchain, plus Go 1.23 compatibility testing.
- Race detector, formatting, vet, cross-build, shell syntax, and Python syntax checks.
- Release creation uses GitHub CLI directly rather than a third-party release action.
- Release publication refuses to proceed without the imported raw Tokyo evidence and a clean worktree.
- Release assets include checksums, a CycloneDX SBOM, source archives, a Git bundle, binaries, and offline reports.

## Non-claims

This release does not claim:

- academic novelty,
- universal correctness over every graph,
- universal performance superiority,
- or that the selected Tokyo trigger generalizes.

The repository is intended for reproducible review and independent validation.
