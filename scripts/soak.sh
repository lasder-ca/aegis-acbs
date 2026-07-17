#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
GRAPH="${1:-${AEGIS_GRAPH:-}}"
OUT="${2:-$ROOT/artifacts/soak}"
ROUNDS="${AEGIS_SOAK_ROUNDS:-10}"
QUERIES="${AEGIS_SOAK_QUERIES:-5000}"
WORKERS="${AEGIS_WORKERS:-4}"
VERIFY_EVERY="${AEGIS_VERIFY_EVERY:-250}"
TIMEOUT="${AEGIS_TIMEOUT:-30s}"

[[ -n "$GRAPH" ]] || { echo "usage: $0 GRAPH [OUTPUT_DIR]" >&2; exit 2; }
[[ -f "$GRAPH" ]] || { echo "graph not found: $GRAPH" >&2; exit 2; }
mkdir -p "$OUT" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)

for ((round=1; round<=ROUNDS; round++)); do
  echo "[soak round=$round/$ROUNDS]"
  GOMAXPROCS="$WORKERS" "$BIN" stress \
    --graph "$GRAPH" --algorithm aegis --queries "$QUERIES" --workers "$WORKERS" \
    --verify-every "$VERIFY_EVERY" --seed "$((7000 + round))" --timeout "$TIMEOUT" \
    --suite mixed --pair-mode strongly-connected \
    --output "$OUT/round-$round.json"
done

python3 - "$OUT" <<'PY'
import json, pathlib, statistics, sys
out=pathlib.Path(sys.argv[1]); reports=[json.loads(p.read_text()) for p in sorted(out.glob('round-*.json'))]
summary={
 'rounds':len(reports), 'completed':sum(r['completed'] for r in reports),
 'verified':sum(r['verified'] for r in reports), 'correct':sum(r['correct'] for r in reports),
 'errors':sum(r['errors'] for r in reports),
 'throughputQpsMedian':statistics.median(r['throughputQps'] for r in reports),
 'p50MsMedian':statistics.median(r['medianNs']/1e6 for r in reports),
 'p95MsMedian':statistics.median(r['p95Ns']/1e6 for r in reports),
 'p99MsWorst':max(r['p99Ns']/1e6 for r in reports),
 'peakRssMiBWorst':max(r['memory']['peakRssBytes']/1048576 for r in reports),
 'goHeapMiBWorst':max(r['memory']['goHeapAllocBytes']/1048576 for r in reports),
 'allVerifiedCorrect':all(r['allVerifiedCorrect'] for r in reports),
}
(out/'soak-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps(summary,indent=2))
PY

echo "soak summary: $OUT/soak-summary.json"
