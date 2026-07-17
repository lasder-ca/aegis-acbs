#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
GRAPH="${1:-${AEGIS_GRAPH:-}}"
OUT="${2:-$ROOT/artifacts/stress-matrix}"
WORKERS_LIST="${AEGIS_WORKERS:-1 2 4 8}"
QUERIES="${AEGIS_STRESS_QUERIES:-5000}"
VERIFY_EVERY="${AEGIS_VERIFY_EVERY:-100}"
SEED="${AEGIS_SEED:-7070}"
TIMEOUT="${AEGIS_TIMEOUT:-30s}"

[[ -n "$GRAPH" ]] || { echo "usage: $0 GRAPH [OUTPUT_DIR]" >&2; exit 2; }
[[ -f "$GRAPH" ]] || { echo "graph not found: $GRAPH" >&2; exit 2; }
mkdir -p "$OUT" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)

for workers in $WORKERS_LIST; do
  echo "[stress workers=$workers queries=$QUERIES]"
  GOMAXPROCS="$workers" "$BIN" stress \
    --graph "$GRAPH" --algorithm aegis --queries "$QUERIES" --workers "$workers" \
    --verify-every "$VERIFY_EVERY" --seed "$((SEED + workers))" --timeout "$TIMEOUT" \
    --suite mixed --pair-mode strongly-connected \
    --output "$OUT/workers-$workers.json"
done

python3 - "$OUT" <<'PY'
import csv, json, pathlib, sys
out=pathlib.Path(sys.argv[1])
rows=[]
for path in sorted(out.glob('workers-*.json')):
    data=json.loads(path.read_text())
    rows.append({
      'workers':data['config']['workers'], 'queries':data['config']['queries'],
      'completed':data['completed'], 'verified':data['verified'], 'correct':data['correct'],
      'errors':data['errors'], 'throughput_qps':data['throughputQps'],
      'mean_ms':data['meanNs']/1e6, 'p50_ms':data['medianNs']/1e6,
      'p95_ms':data['p95Ns']/1e6, 'p99_ms':data['p99Ns']/1e6,
      'worst_ms':data['maxNs']/1e6, 'peak_rss_mib':data['memory']['peakRssBytes']/1048576,
      'go_heap_mib':data['memory']['goHeapAllocBytes']/1048576,
      'all_verified_correct':data['allVerifiedCorrect'],
    })
(out/'stress-matrix.json').write_text(json.dumps(rows,indent=2)+'\n')
with (out/'stress-matrix.csv').open('w',newline='') as f:
    w=csv.DictWriter(f,fieldnames=rows[0].keys()); w.writeheader(); w.writerows(rows)
print('\nworkers  qps       p50_ms   p95_ms   peak_rss_mib  correct')
for r in rows:
    print(f"{r['workers']:>7}  {r['throughput_qps']:>8.2f}  {r['p50_ms']:>8.3f}  {r['p95_ms']:>8.3f}  {r['peak_rss_mib']:>12.2f}  {r['all_verified_correct']}")
PY

echo "stress matrix: $OUT/stress-matrix.csv"
