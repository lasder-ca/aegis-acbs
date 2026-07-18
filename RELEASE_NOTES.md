# Aegis ACBS v0.11.1-experimental

This is a candidate-evaluation release, not the first public GitHub release.

## Added

- `aegis-connect-32`: balanced scheduling from chunk 32 until the first upper bound.
- `aegis-connect-40`: balanced scheduling from chunk 40 until the first upper bound.
- `aegis-connect-32x16`: balanced scheduling from chunk 32 for at most 16 chunks.
- Replay reports now compare all three candidates against normal ACBS, static ACBS, and the classical baselines.
- `validate-v0111-release.sh` runs the reproduced-tail replay, a 10,000-query benchmark, and an automatic candidate selector.

## Publication policy

No candidate is promoted and no GitHub repository is published unless the release gate selects exactly one candidate. The normal `aegis` algorithm remains unchanged in this build.
