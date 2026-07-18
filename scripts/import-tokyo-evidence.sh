#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_ROOT="${1:?usage: import-tokyo-evidence.sh /path/to/tokyo-v12.1.0}"
DEST="$ROOT/research/tokyo-time-2026-07-18"

VALIDATION="$SOURCE_ROOT/tokyo-time-tail-v09"
RELEASE_GATE="$SOURCE_ROOT/v0111-release-gate"
PROFILE="$SOURCE_ROOT/v0112-trigger-profile"

required=(
  "$VALIDATION/regret-validation.json"
  "$VALIDATION/regret-validation.csv"
  "$VALIDATION/regret-validation.html"
  "$RELEASE_GATE/regret-replay-v0111.json"
  "$RELEASE_GATE/regret-replay-v0111.csv"
  "$RELEASE_GATE/regret-replay-v0111.html"
  "$RELEASE_GATE/benchmark-v0111.json"
  "$RELEASE_GATE/benchmark-v0111.html"
  "$RELEASE_GATE/release-gate.json"
  "$RELEASE_GATE/release-gate.txt"
  "$PROFILE/trigger-profile.json"
  "$PROFILE/trigger-profile.csv"
  "$PROFILE/trigger-profile.html"
  "$PROFILE/trigger-profile-gate.txt"
)

for file in "${required[@]}"; do
  [[ -f "$file" ]] || { echo "missing evidence: $file" >&2; exit 1; }
done

mkdir -p "$DEST"
cp "$VALIDATION/regret-validation.json" "$DEST/regret-validation.json"
cp "$VALIDATION/regret-validation.csv" "$DEST/regret-validation.csv"
cp "$VALIDATION/regret-validation.html" "$DEST/regret-validation.html"
cp "$RELEASE_GATE/regret-replay-v0111.json" "$DEST/regret-replay.json"
cp "$RELEASE_GATE/regret-replay-v0111.csv" "$DEST/regret-replay.csv"
cp "$RELEASE_GATE/regret-replay-v0111.html" "$DEST/regret-replay.html"
cp "$RELEASE_GATE/benchmark-v0111.json" "$DEST/rejected-guards-benchmark.json"
cp "$RELEASE_GATE/benchmark-v0111.html" "$DEST/rejected-guards-benchmark.html"
cp "$RELEASE_GATE/release-gate.json" "$DEST/rejected-guards-gate.json"
cp "$RELEASE_GATE/release-gate.txt" "$DEST/rejected-guards-gate.txt"
cp "$PROFILE/trigger-profile.json" "$DEST/trigger-profile.json"
cp "$PROFILE/trigger-profile.csv" "$DEST/trigger-profile.csv"
cp "$PROFILE/trigger-profile.html" "$DEST/trigger-profile.html"
cp "$PROFILE/trigger-profile-gate.txt" "$DEST/trigger-profile-gate.txt"

python3 "$ROOT/scripts/check-v012-release-evidence.py" "$DEST"
(
  cd "$DEST"
  find . -maxdepth 1 -type f ! -name MANIFEST.sha256 -printf '%f\n' \
    | sort \
    | xargs sha256sum > MANIFEST.sha256
)

echo "Tokyo evidence imported: $DEST"
