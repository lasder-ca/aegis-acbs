#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
mkdir -p bin artifacts

go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w" -o bin/aegis ./cmd/aegis
bin/aegis import-osm --input benchdata/hatfield-uk.osm --output artifacts/hatfield-uk.aegis --name hatfield-uk --profile car --metric distance
bin/aegis benchmark --graph artifacts/hatfield-uk.aegis --queries 100 --repeats 5 --batch 8 --research --suite mixed --seed 1010 --output artifacts/hatfield-uk-benchmark.json --html artifacts/hatfield-uk-benchmark.html
