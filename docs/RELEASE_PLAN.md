# GitHub release plan

Aegis ACBS is still an experimental research implementation. Public release is useful for reproducibility and review, but the repository must not claim that ACBS is novel or universally faster until the related-work and multi-city gates are complete.

## v0.11.0-experimental — first public pre-release

Publish the repository and mark the GitHub Release as a **pre-release** after all of these gates pass:

1. The reproduced `adaptive-scheduler-tail` case improves by at least 0.5 ms with `aegis-late-guard`.
2. The reproduced `persistent-classical-tail` case regresses by less than 1 ms.
3. A 10,000-query Tokyo time benchmark remains 100% correct.
4. `aegis-late-guard` does not regress mean, median, or p95 by more than 1% against the normal `aegis` implementation.
5. Linux, Windows, and macOS CI pass; the race detector passes on Linux.
6. Release assets include source archives, binaries, SHA-256 checksums, SBOM, benchmark JSON, and self-contained HTML reports.

Run the complete Tokyo replay, 10,000-query comparison, and local gate with:

```bash
scripts/validate-v011-release.sh \
  path/to/tokyo-time.aegis \
  path/to/tokyo-time-tail-v09 \
  artifacts/v011-release-gate
```

To re-check already generated reports only:

```bash
scripts/check-v011-release-gate.py \
  path/to/regret-replay-v011.json \
  path/to/benchmark-tokyo-time-v011.json
```

When the gates pass, publish within the same development session rather than waiting for a later feature release. The release title must include `experimental`, and GitHub must mark it as a pre-release.

## v0.12.0-alpha — promoted research preview

Promote to `alpha` only after:

- Tokyo, Yokohama, Osaka, and Nagoya distance/time matrices complete;
- at least 10 seeds and 10,000 queries per metric are retained;
- all routes match Dijkstra;
- no city shows more than a 2% median or p95 regression from the best accepted ACBS version;
- the exact stopping condition is compared in writing against MM, NBS, DVCBS, BAE*, and MEET;
- a clean machine can reproduce the published benchmark from documented commands.

## v1.0.0 — stable

Do not publish a stable release until:

- an independent person reproduces correctness and at least one large-road benchmark;
- the API and graph file format have a compatibility policy;
- fuzzing and long-running parallel stress tests pass;
- security reporting and supported-version policies are defined;
- novelty claims, if any, are reviewed separately from engineering performance claims.

Open-sourcing the code does not require claiming research novelty. The README should consistently describe ACBS as an experimental exact bidirectional shortest-path implementation until the novelty review is complete.
