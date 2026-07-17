# Changelog

## 0.3.0-experimental

- Split queue work into pushes, pops, and stale pops.
- Split incumbent pruning into pop-time and relaxation-time counters.
- Add meeting-check and p99 latency metrics.
- Define ratio-of-medians, median per-query, and geometric-mean speedups separately.
- Rename sub-1.0 baseline comparison to relative runtime and add clamped oracle regret.
- Add `aegis aggregate` with self-contained JSON, CSV, and HTML matrix reports.
- Add five-seed multi-graph benchmark automation.
- Add regional Japan preparation for Tokyo, Yokohama, Osaka, and Nagoya, each with distance and time graphs.
- Add matrix aggregation and pruning-counter consistency tests.

## 0.2.0-experimental

- Replace per-node great-circle evaluation with precomputed unit vectors and a consistent chord-distance potential.
- Change adaptive chunks from node-count budgets to edge-work budgets.
- Add incumbent-bound pruning using exact admissible forward/backward bounds.
- Add exactness certificate fields: upper bound, lower bound, and optimality gap.
- Add scheduler, potential, pruning, and upper-bound diagnostics.
- Add `aegis-static` and `aegis-no-prune` research ablations plus `benchmark --research`.
- Add 10,000 random directed time-road differential queries and chord admissibility checks.
- Add a formal novelty checklist and expanded correctness notes.

## 0.1.0-experimental

- Introduce Aegis Coupled-Bound Search.
- Replace the default Aegis portfolio selector with one exact coupled bidirectional algorithm.
- Add balanced-potential reduced costs, adaptive bound-progress scheduling, proof-oriented metrics, and ACBS visual reports.
