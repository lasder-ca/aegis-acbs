#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRAPHS="${AEGIS_GRAPH_DIR:-$ROOT/.data/graphs}"
OUT="${AEGIS_REPORT_DIR:-$ROOT/artifacts/benchmarks}"
QUERIES="${AEGIS_QUERIES:-300}"
REPEATS="${AEGIS_REPEATS:-9}"
mkdir -p "$OUT"
cd "$ROOT"
[[ -x bin/aegis ]] || go build -trimpath -o bin/aegis ./cmd/aegis
for graph in "$GRAPHS"/*.aegis; do
  [[ -f "$graph" ]] || continue
  name="$(basename "$graph" .aegis)"
  bin/aegis benchmark --graph "$graph" --queries "$QUERIES" --repeats "$REPEATS" --seed 1010 --suite mixed --output "$OUT/$name.json" --html "$OUT/$name.html"
done
