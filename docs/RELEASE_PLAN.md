# GitHub release plan

Aegis ACBS remains an experimental exact shortest-path implementation. Public release is for reproducibility and review, not a claim of research novelty or universal superiority.

## v0.11.1-experimental — connection-guard selection

Evaluate three candidates against the reproduced Tokyo tails and a 10,000-query time benchmark. A candidate passes only when all conditions hold:

1. Every reproduced `adaptive-scheduler-tail` improves by at least 0.5 ms.
2. Every `persistent-classical-tail` regresses by less than 1 ms.
3. All 10,000 routes match Dijkstra.
4. Mean, median, and p95 regress by no more than 1%.
5. Median relaxed edges and expanded nodes regress by no more than 1%.
6. Linux tests and race detector pass; Windows and macOS builds succeed.

Run:

```bash
scripts/validate-v0111-release.sh \
  path/to/tokyo-time.aegis \
  path/to/tokyo-time-tail-v09 \
  artifacts/v0111-release-gate
```

The gate prints `SELECTED: <algorithm>` and `RELEASE GATE: PASS`. If no candidate passes, keep the repository private and do not publish a GitHub Release.

## First public pre-release

Publish `lasder-ca/aegis-acbs` only after the v0.11.1 gate passes. Promote the selected candidate in a follow-up tagged build, then publish it as a GitHub pre-release with source, binaries, checksums, SBOM, replay JSON/CSV/HTML, benchmark JSON/HTML, and documented limitations.

## v0.12.0-alpha

Require Tokyo, Yokohama, Osaka, and Nagoya distance/time matrices; at least 10 seeds and 10,000 queries per metric; 100% Dijkstra agreement; no city with more than 2% median or p95 regression; and written comparison against MM, NBS, DVCBS, BAE*, and MEET.

## v1.0.0

Require independent reproduction, API and graph-format compatibility policy, fuzzing, long-running parallel stress, security policy, and a separately reviewed novelty claim.
