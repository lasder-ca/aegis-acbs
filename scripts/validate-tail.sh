#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"; GRAPH="${1:-${AEGIS_GRAPH:-}}"; OUT="${2:-$ROOT/artifacts/tail-validation}"
QUERIES="${AEGIS_QUERIES:-1000}"; REPEATS="${AEGIS_REPEATS:-3}"; BATCH="${AEGIS_BATCH:-1}"
SEEDS="${AEGIS_SEEDS:-1010 20260717 424242 8675309 123456789 314159265 271828182 161803398 141421356 173205080}"
RATIO="${AEGIS_RATIO_THRESHOLD:-1.25}"; PENALTY="${AEGIS_PENALTY_FLOOR:-1ms}"; MAX_RATE="${AEGIS_MAX_MEANINGFUL_RATE:-0}"; TIMEOUT="${AEGIS_TIMEOUT:-30s}"
[[ -n "$GRAPH" ]] || { echo "usage: $0 GRAPH [OUTPUT_DIR]" >&2; exit 2; }; [[ -f "$GRAPH" ]] || { echo "graph not found: $GRAPH" >&2; exit 2; }
mkdir -p "$OUT" "$ROOT/bin"; [[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)
valid_report(){ python3 - "$1" "$QUERIES" <<'PY'
import json,pathlib,sys
p=pathlib.Path(sys.argv[1]); expected=int(sys.argv[2])
try:d=json.loads(p.read_text()); ok=d.get('allCorrect') is True and d.get('config',{}).get('queries')==expected and len(d.get('samples',[]))>0
except Exception:ok=False
raise SystemExit(0 if ok else 1)
PY
}
runs=0
for seed in $SEEDS; do
  run="$OUT/seed-$seed"; report="$run/report.json"; mkdir -p "$run"
  if [[ -f "$report" ]] && valid_report "$report"; then echo "[resume seed=$seed] using $report"; else
    echo "[benchmark seed=$seed queries=$QUERIES]"; rm -f "$report.tmp" "$run/report.html.tmp"
    GOMAXPROCS="${GOMAXPROCS:-1}" "$BIN" benchmark --graph "$GRAPH" --queries "$QUERIES" --repeats "$REPEATS" --batch "$BATCH" --order interleaved --algorithms dijkstra,bidijkstra,astar,aegis --suite mixed --pair-mode strongly-connected --seed "$seed" --timeout "$TIMEOUT" --output "$report.tmp" --html "$run/report.html.tmp"
    mv "$report.tmp" "$report"; mv "$run/report.html.tmp" "$run/report.html"
  fi
  runs=$((runs+1))
done
minimum=$((QUERIES*runs))
"$BIN" validate-regret --input-dir "$OUT" --algorithm aegis --ratio-threshold "$RATIO" --penalty-floor "$PENALTY" --min-queries "$minimum" --max-meaningful-rate "$MAX_RATE" --top 100 --output "$OUT/regret-validation.json" --csv "$OUT/regret-validation.csv" --html "$OUT/regret-validation.html" --fail-on-violation=false
python3 - "$OUT/regret-validation.json" <<'PY'
import json,pathlib,sys
r=json.loads(pathlib.Path(sys.argv[1]).read_text())
print('\nsummary'); print(f"  runs: {r['files']}"); print(f"  queries: {r['totalQueries']}"); print(f"  meaningful: {r['totalMeaningful']} ({100*r['meaningfulRate']:.6f}%)"); print(f"  Wilson 95%: {100*r['meaningfulRateWilsonLow95']:.6f}% .. {100*r['meaningfulRateWilsonHigh95']:.6f}%")
if r['totalMeaningful']==0: print(f"  exact zero-event upper 95%: {100*r['zeroEventUpper95']:.6f}%")
print(f"  p95 penalty: {r['p95PenaltyNs']/1e6:.3f} ms"); print(f"  max penalty: {r['maxPenaltyNs']/1e6:.3f} ms"); print(f"  all correct: {r['allCorrect']}"); print(f"  pass: {r['passed']}")
raise SystemExit(0 if r['passed'] else 1)
PY
