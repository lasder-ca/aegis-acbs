#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
GRAPH="${1:-}"
VALIDATION_ROOT="${2:-}"
OUT="${3:-$ROOT/artifacts/v011-release-gate}"
QUERIES="${AEGIS_QUERIES:-10000}"
REPEATS="${AEGIS_REPEATS:-3}"
SEED="${AEGIS_SEED:-1010}"
TIMEOUT="${AEGIS_TIMEOUT:-30s}"

[[ -n "$GRAPH" && -f "$GRAPH" ]] || { echo "usage: $0 GRAPH.aegis VALIDATION_ROOT [OUTPUT_DIR]" >&2; exit 2; }
[[ -n "$VALIDATION_ROOT" && -f "$VALIDATION_ROOT/regret-validation.json" ]] || {
  echo "validation root must contain regret-validation.json: $VALIDATION_ROOT" >&2
  exit 2
}
mkdir -p "$OUT" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)

GOMAXPROCS=1 "$BIN" replay-regret \
  --graph "$GRAPH" \
  --validation "$VALIDATION_ROOT/regret-validation.json" \
  --input-root "$VALIDATION_ROOT" \
  --runs 31 --warmup 5 --timeout "$TIMEOUT" \
  --ratio-threshold 1.25 --penalty-floor 1ms --top 100 \
  --output "$OUT/regret-replay-v011.json" \
  --csv "$OUT/regret-replay-v011.csv" \
  --html "$OUT/regret-replay-v011.html"

GOMAXPROCS=1 "$BIN" benchmark \
  --graph "$GRAPH" \
  --queries "$QUERIES" --repeats "$REPEATS" --batch 1 \
  --order interleaved \
  --algorithms aegis,aegis-late-guard \
  --suite mixed --pair-mode strongly-connected \
  --seed "$SEED" --timeout "$TIMEOUT" \
  --output "$OUT/benchmark-v011.json" \
  --html "$OUT/benchmark-v011.html"

"$ROOT/scripts/check-v011-release-gate.py" \
  "$OUT/regret-replay-v011.json" \
  "$OUT/benchmark-v011.json" | tee "$OUT/release-gate.txt"

echo "release gate artifacts: $OUT"
