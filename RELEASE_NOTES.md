# Aegis ACBS v0.11.0-experimental

This release adds an opt-in, narrowly scoped late-upper-bound guard and the release gates needed to decide whether it should ever replace the default scheduler. The default `aegis` algorithm remains unchanged.

## Added

- `aegis-late-guard`, an experimental ACBS variant.
- The guard is eligible only for travel-time routes.
- It can trigger only after 48 completed chunks, while no upper bound exists, when direction switching is frequent and the two measured efficiency scores are within 25%.
- Once triggered, it uses the static lower-key direction rule and base edge budget for at most eight chunks.
- Search statistics now record guard activations, guarded chunks, and the first guarded chunk.
- ACBS traces mark the trigger and active guard chunks.
- `replay-regret` now measures Dijkstra, bidirectional Dijkstra, A*, static ACBS, guarded ACBS, and normal ACBS.
- Replay JSON/CSV/HTML reports show guarded latency, improvement, regression, and an acceptance verdict.
- `scripts/check-v011-release-gate.py` enforces the Tokyo release criteria.
- `docs/RELEASE_PLAN.md` defines the public GitHub pre-release, alpha, and stable gates.

## Acceptance criteria before public GitHub pre-release

1. The reproduced scheduler-tail case improves by at least 0.5 ms.
2. The reproduced persistent-classical-tail case regresses by less than 1 ms.
3. A 10,000-query Tokyo time run is 100% correct.
4. Mean, median, and p95 do not regress by more than 1% versus normal ACBS.
5. CI and the race detector pass.

## Deliberately not changed

The default scheduler and exact stopping condition are unchanged. The guarded variant is not promoted until the external Tokyo graph validates the acceptance criteria.
