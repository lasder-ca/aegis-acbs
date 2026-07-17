#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
GRAPH="${1:-}"
OUT="${2:-$ROOT/artifacts/memory-profile}"
QUERIES="${AEGIS_QUERIES:-1000}"
SEED="${AEGIS_SEED:-1010}"

[[ -n "$GRAPH" && -f "$GRAPH" ]] || { echo "usage: $0 GRAPH.aegis [OUTPUT_DIR]" >&2; exit 2; }
command -v /usr/bin/time >/dev/null || { echo "/usr/bin/time is required" >&2; exit 1; }
mkdir -p "$OUT" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)

/usr/bin/time -v "$BIN" benchmark \
  --graph "$GRAPH" --queries "$QUERIES" --repeats 3 --batch 1 \
  --order interleaved --algorithms aegis --measure-memory \
  --suite mixed --pair-mode strongly-connected --seed "$SEED" --timeout 30s \
  --output "$OUT/report.json" --html "$OUT/report.html" \
  2>&1 | tee "$OUT/process-memory.txt"
