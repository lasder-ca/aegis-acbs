#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
VERSION="$(cat VERSION)"
command -v gh >/dev/null || { echo "GitHub CLI (gh) is required" >&2; exit 1; }
gh auth status
scripts/build-release.sh
if ! gh repo view lasder-ca/aegis-acbs >/dev/null 2>&1; then
  gh repo create lasder-ca/aegis-acbs --public --source=. --remote=origin --description "Experimental exact coupled-bound shortest-path search for real road networks"
fi
git push -u origin main
git push origin "v$VERSION"
gh release create "v$VERSION" dist/* --title "Aegis ACBS v$VERSION" --notes-file RELEASE_NOTES.md
