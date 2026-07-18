# GitHub release plan

Aegis ACBS remains an experimental exact shortest-path implementation. Public release is for reproducibility and review, not a claim of research novelty or universal superiority.

## Completed Tokyo gates

The v0.11.1 connection-guard gate rejected every candidate without changing the predeclared thresholds:

- `aegis-connect-32`: scheduler-tail gain 0.778 ms, but global mean +15.95% and p95 +14.98%.
- `aegis-connect-40`: scheduler-tail gain 0.732 ms, but global mean +14.98% and p95 +15.52%.
- `aegis-connect-32x16`: global latency was neutral or slightly better, but scheduler-tail gain was only 0.283 ms and expanded nodes increased 2.46%.

The default `aegis` scheduler remains unchanged. These negative results are part of the research record and must not be hidden or reclassified by relaxing the gate after measurement.

## v0.11.2-experimental — narrow-trigger diagnosis

Run `profile-trigger` over all 10,000 Tokyo validation queries. It records deterministic features at chunks 24, 32, 40, and 48, then ranks interpretable threshold rules using the replay-confirmed scheduler tail as the positive label.

A trigger is only eligible for later experimentation when:

1. Every reproduced adaptive-scheduler tail is detected.
2. The trigger matches no more than five of the 10,000 queries.
3. Every traced route remains correct.
4. The selected rule remains stable when the positive query is re-profiled.
5. The rule is validated on at least one additional seed set before scheduler integration.

No rule is promoted automatically by v0.11.2.

## First public repository

Publish `lasder-ca/aegis-acbs` after the v0.11.2 trigger profile has been produced, whether or not a narrow rule is found. The public repository must state:

- 10,000/10,000 Tokyo time routes agreed with Dijkstra.
- One reproducible adaptive-scheduler tail and one reproducible persistent classical tail were observed in 10,000 queries.
- No algorithm is guaranteed to be fastest for every query.
- The v0.11.1 guard candidates failed the declared release gate.
- Research novelty has not been independently verified.

The repository itself may be public before a GitHub Release. The first tagged downloadable release is `v0.12.0-research-preview`, based on the unchanged default ACBS plus diagnostic tooling. Failed guard variants remain available only as documented experimental ablations.

## v0.12.0-research-preview

Required before tagging:

1. Tokyo trigger profile JSON/CSV/HTML.
2. Tokyo distance and time correctness reports.
3. Linux tests, `go vet`, race detector, Windows/macOS cross-builds.
4. SHA-256 and SBOM.
5. Reproduction scripts that do not require private source code.
6. README, LICENSE, SECURITY, CONTRIBUTING, and limitations.
7. Public benchmark claims limited to the measured hardware, graph, metric, seeds, and command line.

## v0.13.0-alpha

Require Tokyo, Yokohama, Osaka, and Nagoya distance/time matrices; at least 10 seeds and 10,000 queries per metric; 100% Dijkstra agreement; no city with more than 2% median or p95 regression; and written comparison against MM, NBS, DVCBS, BAE*, and MEET.

## v1.0.0

Require independent reproduction, API and graph-format compatibility policy, fuzzing, long-running parallel stress, security policy, and a separately reviewed novelty claim.
