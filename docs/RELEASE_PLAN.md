# Release plan

## v0.1.0 — first public research release

The default `aegis` search remains unchanged from v0.11.2. The release publishes the algorithm, reproducibility tools, successful tests, rejected guard experiments, and explicit limitations.

Publication requirements:

- 10,000-query Tokyo validation imported as raw JSON/CSV/HTML.
- 10,000/10,000 distance agreement with Dijkstra.
- Isolated replay records one scheduler tail, one persistent classical tail, and nine non-reproduced cases.
- Rejected connection-guard gate remains failed with no selected candidate.
- Trigger profile contains 10,000 correct queries, zero trace errors, zero unstable labels, and the checkpoint-48 diagnostic rule.
- `go test`, `go vet`, race detector, formatting, shell/Python syntax, and cross-build checks pass.
- Linux, Windows, and macOS assets, source archives, Git bundle, checksums, and SBOM are generated.

The checkpoint-48 rule is diagnostic only. It must not change the default scheduler in this release.

## v0.2.0 — independent generalization

Target independent city and seed validation:

- Tokyo holdout query generation that was not used during rule discovery.
- Yokohama, Osaka, and Nagoya distance/time graphs.
- At least 10,000 queries per city/metric or a documented statistical power target.
- Comparison with the default ACBS, static scheduler, Dijkstra, bidirectional Dijkstra, and A*.
- Report trigger prevalence, precision, recall, performance impact, and failure cases.

The trigger may only become an experimental scheduler option if it is useful outside the discovery dataset and does not regress global latency or work beyond a predeclared threshold.

## v1.0.0

Not planned until:

- independent correctness and benchmark reproduction,
- fuzzing and long-running concurrent stress,
- stable graph/API compatibility policy,
- broader related-work and novelty review,
- and clear separation between academic and engineering claims.
