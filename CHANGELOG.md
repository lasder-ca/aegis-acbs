# Changelog

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
