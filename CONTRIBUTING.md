# Contributing

Aegis ACBS is a research prototype. Changes should preserve exactness, make measurements reproducible, and keep failed experiments visible.

## Local checks

```bash
gofmt -w cmd internal
go test ./...
go vet ./...
go test -race ./internal/search ./internal/graph ./internal/bench ./internal/server
bash -n scripts/*.sh
python3 -m py_compile scripts/*.py
```

## Algorithm changes

Every algorithm or scheduler change must include:

- randomized or exhaustive comparison against Dijkstra,
- a deterministic regression fixture when fixing a specific query,
- unchanged lower-bound/upper-bound exactness unless a new proof is documented,
- and a predeclared performance gate before large benchmark results are inspected.

Do not relax an acceptance threshold after seeing a result.

## Benchmark changes

- Use the same graph and query pairs for all compared algorithms.
- Interleave or deterministically rotate measurement order.
- Separate runtime, search work, allocation, and correctness claims.
- Do not add preprocessing to only one algorithm without reporting it.
- Keep raw JSON and the command line needed to reproduce it.

## Research claims

Do not claim academic novelty or universal superiority without independent evidence. Related-work corrections and negative results are welcome.
