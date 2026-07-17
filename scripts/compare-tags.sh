#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRAPH="${1:-$ROOT/artifacts/hatfield-uk.aegis}"
OUT="${2:-$ROOT/artifacts/tag-comparison}"
OLD_TAG="${OLD_TAG:-v0.1.0-experimental}"
QUERIES="${QUERIES:-100}"
REPEATS="${REPEATS:-5}"
BATCH="${BATCH:-8}"
mkdir -p "$OUT"
GRAPH="$(realpath "$GRAPH")"
TMP="$(mktemp -d)"
trap 'git -C "$ROOT" worktree remove --force "$TMP/old" >/dev/null 2>&1 || true; rm -rf "$TMP"' EXIT

git -C "$ROOT" worktree add --detach "$TMP/old" "$OLD_TAG" >/dev/null
go build -trimpath -o "$TMP/current-aegis" "$ROOT/cmd/aegis"
(
  cd "$TMP/old"
  go build -trimpath -o "$TMP/old-aegis" ./cmd/aegis
)

run_one() {
  local binary="$1" output="$2"
  GOMAXPROCS=1 "$binary" benchmark \
    --graph "$GRAPH" --queries "$QUERIES" --repeats "$REPEATS" --batch "$BATCH" \
    --suite mixed --pair-mode strongly-connected --seed 1010 \
    --algorithms dijkstra,bidijkstra,astar,aegis \
    --output "$output" --html "${output%.json}.html"
}
run_one "$TMP/old-aegis" "$OUT/old.json"
run_one "$TMP/current-aegis" "$OUT/current.json"

python3 - "$OUT/old.json" "$OUT/current.json" "$OUT/summary.md" <<'PY'
import json, sys
old=json.load(open(sys.argv[1])); cur=json.load(open(sys.argv[2]))
def row(d, name): return next(x for x in d['summary'] if x['algorithm']==name)
o=row(old,'aegis'); c=row(cur,'aegis')
def pct(a,b): return (a-b)*100/a if a else 0
text=f'''# ACBS tag comparison\n\n| Metric | {old["version"]} | {cur["version"]} | Change |\n|---|---:|---:|---:|\n| p50 | {o["medianNs"]/1e3:.2f} µs | {c["medianNs"]/1e3:.2f} µs | {pct(o["medianNs"],c["medianNs"]):+.1f}% faster |\n| p95 | {o["p95Ns"]/1e3:.2f} µs | {c["p95Ns"]/1e3:.2f} µs | {pct(o["p95Ns"],c["p95Ns"]):+.1f}% faster |\n| median relaxed | {o["medianRelaxed"]} | {c["medianRelaxed"]} | {pct(o["medianRelaxed"],c["medianRelaxed"]):+.1f}% lower |\n| correct | {o["correct"]}/{o["runs"]} | {c["correct"]}/{c["runs"]} | — |\n'''
open(sys.argv[3],'w').write(text)
print(text)
PY
