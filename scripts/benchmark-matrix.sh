#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
GRAPH_DIR="${AEGIS_GRAPH_DIR:-$ROOT/.data/regional-graphs}"
OUT="${AEGIS_REPORT_DIR:-$ROOT/artifacts/matrix}"
SEEDS="${AEGIS_SEEDS:-1010 20260717 424242 8675309 123456789}"
QUERIES="${AEGIS_QUERIES:-50}"
REPEATS="${AEGIS_REPEATS:-3}"
BATCH="${AEGIS_BATCH:-1}"
TIMEOUT="${AEGIS_TIMEOUT:-30s}"
ORDER="${AEGIS_ORDER:-interleaved}"
MEASURE_MEMORY="${AEGIS_MEASURE_MEMORY:-0}"
GOMAXPROCS="${GOMAXPROCS:-1}"
export GOMAXPROCS

mkdir -p "$OUT" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)
shopt -s nullglob
graphs=("$GRAPH_DIR"/*.aegis)
(( ${#graphs[@]} > 0 )) || { echo "no .aegis graphs in $GRAPH_DIR" >&2; exit 1; }

for graph in "${graphs[@]}"; do
  name="$(basename "$graph" .aegis)"
  for seed in $SEEDS; do
    run_dir="$OUT/$name/seed-$seed"
    mkdir -p "$run_dir"
    echo "[$name seed=$seed]"
    memory_flag=()
    [[ "$MEASURE_MEMORY" == "1" ]] && memory_flag+=(--measure-memory)
    "$BIN" benchmark \
      --graph "$graph" \
      --queries "$QUERIES" --repeats "$REPEATS" --batch "$BATCH" \
      --order "$ORDER" "${memory_flag[@]}" \
      --suite mixed --pair-mode strongly-connected --seed "$seed" --timeout "$TIMEOUT" \
      --research \
      --output "$run_dir/report.json" --html "$run_dir/report.html"
  done
done

"$BIN" aggregate --input-dir "$OUT" \
  --output "$OUT/benchmark-matrix.json" \
  --csv "$OUT/benchmark-matrix.csv" \
  --html "$OUT/benchmark-matrix.html"

echo "matrix: $OUT/benchmark-matrix.html"
