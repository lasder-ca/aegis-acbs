#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
VERSION="$(cat VERSION)"

if [[ "$VERSION" == "0.11.2-experimental" ]]; then
  echo "v0.11.2 publishes the repository only; use scripts/publish-research-repo.sh after the 10,000-query trigger profile" >&2
  exit 2
fi

[[ "${AEGIS_PUBLISH_CONFIRM:-}" == "YES" ]] || {
  echo "refusing to publish: set AEGIS_PUBLISH_CONFIRM=YES after reviewing the release" >&2
  exit 2
}
command -v gh >/dev/null || { echo "GitHub CLI (gh) is required" >&2; exit 1; }
gh auth status
git diff --quiet && git diff --cached --quiet || { echo "working tree must be clean" >&2; exit 1; }

if [[ "$VERSION" == "0.11.1-experimental" ]]; then
  : "${AEGIS_RELEASE_GATE_JSON:?set AEGIS_RELEASE_GATE_JSON to the v0.11.1 release-gate.json}"
  python3 - "$AEGIS_RELEASE_GATE_JSON" <<'PY_GATE'
import json, sys
r=json.load(open(sys.argv[1], encoding="utf-8"))
if not r.get("pass") or not r.get("selected"):
    raise SystemExit("refusing to publish: v0.11.1 gate did not select a passing candidate")
print("release gate selected:", r["selected"])
PY_GATE
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
