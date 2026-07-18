# Aegis ACBS v0.11.2-experimental

v0.11.2 does not change the default `aegis` search. The three v0.11.1 connection-guard candidates failed the Tokyo 10,000-query release gate: the unbounded variants improved one reproduced scheduler tail but regressed global mean and p95 by about 15%, while the bounded variant preserved latency but did not meet the required tail improvement and increased search work.

This release adds whole-suite trigger profiling:

- `profile-trigger` traces every query referenced by a multi-seed validation report once.
- Only deterministic checkpoint features at chunks 24, 32, 40, and 48 are retained.
- Replay-confirmed scheduler and persistent tails are used as labels.
- Interpretable one- and two-condition rules are ranked by recall, total matches, false positives, and the configured maximum trigger population.
- An eligible rule must cover every confirmed scheduler tail and match no more than five queries by default.
- Extended opt-in trace telemetry records queue priorities, stale-pop deltas, finite-meeting deltas, connection checks, and cumulative directional work. Normal routing and benchmark runs remain untraced.

The purpose is to determine whether the reproduced Tokyo query 877 tail can be distinguished from the other 9,999 queries without broad guard activation. No trigger is promoted into the default scheduler by this release.

Known Tokyo evidence retained from v0.11.1:

- 10,000/10,000 shortest paths agreed with Dijkstra.
- One reproducible adaptive-scheduler tail and one reproducible persistent classical tail were observed.
- `aegis-connect-32` and `aegis-connect-40` were rejected for approximately 15% global mean/p95 regression.
- `aegis-connect-32x16` was rejected because its 0.283 ms scheduler-tail gain was below the predeclared 0.500 ms threshold and its expanded-node regression exceeded 1%.

Research novelty and universal performance superiority are not claimed.
