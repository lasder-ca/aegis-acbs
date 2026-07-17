#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
VERSION="$(cat VERSION)"

[[ "${AEGIS_PUBLISH_CONFIRM:-}" == "YES" ]] || {
  echo "refusing to publish: set AEGIS_PUBLISH_CONFIRM=YES after reviewing the release" >&2
  exit 2
}
command -v gh >/dev/null || { echo "GitHub CLI (gh) is required" >&2; exit 1; }
gh auth status
git diff --quiet && git diff --cached --quiet || { echo "working tree must be clean" >&2; exit 1; }

if [[ "$VERSION" == "0.11.0-experimental" ]]; then
  : "${AEGIS_REPLAY_JSON:?set AEGIS_REPLAY_JSON to the v0.11 replay report}"
  : "${AEGIS_BENCHMARK_JSON:?set AEGIS_BENCHMARK_JSON to the v0.11 10k benchmark report}"
  scripts/check-v011-release-gate.py "$AEGIS_REPLAY_JSON" "$AEGIS_BENCHMARK_JSON"
fi

scripts/build-release.sh
if ! gh repo view lasder-ca/aegis-acbs >/dev/null 2>&1; then
  gh repo create lasder-ca/aegis-acbs --public --source=. --remote=origin \
    --description "Experimental exact coupled-bound shortest-path search for real road networks"
fi

git push -u origin main
git push origin "v$VERSION"
release_args=("v$VERSION" dist/* --title "Aegis ACBS v$VERSION" --notes-file RELEASE_NOTES.md)
case "$VERSION" in
  *experimental*|*alpha*|*beta*|*rc*) release_args+=(--prerelease) ;;
esac
gh release create "${release_args[@]}"
