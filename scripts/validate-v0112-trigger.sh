#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GRAPH="${1:?usage: validate-v0112-trigger.sh GRAPH VALIDATION_ROOT REPLAY_JSON OUT_DIR}"
VALIDATION_ROOT="${2:?usage: validate-v0112-trigger.sh GRAPH VALIDATION_ROOT REPLAY_JSON OUT_DIR}"
REPLAY_JSON="${3:?usage: validate-v0112-trigger.sh GRAPH VALIDATION_ROOT REPLAY_JSON OUT_DIR}"
OUT="${4:?usage: validate-v0112-trigger.sh GRAPH VALIDATION_ROOT REPLAY_JSON OUT_DIR}"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
VALIDATION_JSON="${AEGIS_VALIDATION_JSON:-$VALIDATION_ROOT/regret-validation.json}"

mkdir -p "$OUT"

GOMAXPROCS="${GOMAXPROCS:-1}" "$BIN" profile-trigger \
  --graph "$GRAPH" \
  --validation "$VALIDATION_JSON" \
  --replay "$REPLAY_JSON" \
  --input-root "$VALIDATION_ROOT" \
  --checkpoints "${AEGIS_CHECKPOINTS:-24,32,40,48}" \
  --max-matches "${AEGIS_MAX_MATCHES:-5}" \
  --top-rules "${AEGIS_TOP_RULES:-50}" \
  --label-repeats "${AEGIS_LABEL_REPEATS:-3}" \
  --timeout "${AEGIS_TIMEOUT:-30s}" \
  --output "$OUT/trigger-profile.json" \
  --csv "$OUT/trigger-profile.csv" \
  --html "$OUT/trigger-profile.html"

python3 "$ROOT/scripts/check-v0112-trigger-profile.py" \
  "$OUT/trigger-profile.json" \
  | tee "$OUT/trigger-profile-gate.txt"
