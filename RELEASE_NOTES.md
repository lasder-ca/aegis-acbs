# Aegis ACBS v0.7.0-experimental

v0.7.0 is a production-path cleanup and validation release. It does not change the exact coupled-bound termination rule, balanced chord potential, radix queues, or CSR graph representation.

## Main changes

- Disable the inactive incumbent-bound pruning pass in the default `aegis` path.
- Add `aegis-prune` as an explicit experimental variant.
- Make `aegis-static` differ from `aegis` only in direction scheduling.
- Keep `aegis-no-prune` as a compatibility alias for older command lines.
- Move the weak linear-projection variant behind `--experimental`.
- Change `--research` to include only the scheduler ablation.
- Rename CLI wording to `runtime-vs-fastest-classical` and `classical-oracle-regret`.
- Label ACBS work counters as medians in CLI output.
- Add `aegis stress` for concurrent in-process validation with sampled Dijkstra checking.
- Add worker-scaling and repeated soak scripts.

## Interpretation

The incumbent-pruning proof remains valid, but the feature did not activate on the Tokyo v0.6 runs. Removing it from the production path simplifies the algorithm claim and prevents unnecessary potential-bound evaluations. This is not a claim that pruning can never help on another domain; the exact experiment remains available as `aegis-prune`.

## Research status

Academic novelty remains unconfirmed. Publication claims still require same-codebase comparisons with MM/MMe, NBS, DVCBS, BAE*, and MEET-style termination methods, plus independent multi-city reproduction.
