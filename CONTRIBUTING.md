# Contributing

Before submitting changes:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go test -race ./...
```

Algorithm changes must include a randomized or exhaustive comparison against Dijkstra. Benchmark improvements must not add hidden preprocessing to only one algorithm.
