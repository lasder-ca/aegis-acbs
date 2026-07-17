# Aegis ACBS v0.9.0-experimental

This release turns the v0.8 query-level diagnosis into a resumable, multi-seed validation workflow. The ACBS search policy is unchanged.

## Added

- `aegis validate-regret` recursively reads benchmark JSON reports and aggregates meaningful slowdowns.
- Per-run and global query counts, slowdown rates, p50/p95/max ratios, and absolute penalties.
- Wilson 95% confidence intervals for the observed meaningful-slowdown rate.
- An exact one-sided 95% upper bound when no meaningful slowdown is observed. For 10,000 zero-event queries this upper bound is approximately 0.03%.
- Acceptance criteria through `--min-queries` and `--max-meaningful-rate`.
- Self-contained JSON, CSV, and HTML validation reports.
- `scripts/validate-tail.sh`, which resumes completed seeds and validates the combined result.

## Deliberately not changed

The production `aegis` scheduler, potential, radix queues, CSR graph representation, and exact termination condition are identical to v0.8. This release improves evidence quality rather than changing route selection.

## Recommended first stage

Run 1,000 queries over ten deterministic seeds. If zero meaningful slowdowns are observed, the exact one-sided 95% upper bound is about 0.03%. A second 10,000-query stage can then target fewer seeds for deeper tail coverage.
