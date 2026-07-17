#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
mkdir -p bin artifacts artifacts/repro-graphs

go test ./...
go vet ./...
go build -trimpath -ldflags "-s -w" -o bin/aegis ./cmd/aegis

bin/aegis import-osm --input benchdata/hatfield-uk.osm --output artifacts/repro-graphs/hatfield-distance.aegis --name hatfield-uk --profile car --metric distance
bin/aegis import-osm --input benchdata/hatfield-uk.osm --output artifacts/repro-graphs/hatfield-time.aegis --name hatfield-uk --profile car --metric time
cp artifacts/repro-graphs/hatfield-distance.aegis artifacts/hatfield-uk.aegis

GOMAXPROCS=1 bin/aegis benchmark --graph artifacts/repro-graphs/hatfield-distance.aegis --queries 100 --repeats 5 --batch 8 --research --suite mixed --seed 1010 --output artifacts/hatfield-uk-benchmark.json --html artifacts/hatfield-uk-benchmark.html

AEGIS_BIN="$ROOT/bin/aegis" \
AEGIS_GRAPH_DIR="$ROOT/artifacts/repro-graphs" \
AEGIS_REPORT_DIR="$ROOT/artifacts/matrix" \
AEGIS_QUERIES=30 AEGIS_REPEATS=3 AEGIS_BATCH=8 \
AEGIS_SEEDS="1010 20260717 424242" \
scripts/benchmark-matrix.sh
