#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REPO="${AEGIS_GITHUB_REPO:-lasder-ca/aegis-acbs}"
VERSION="$(cat VERSION)"
TAG="v$VERSION"

[[ "$VERSION" == "0.12.0-research-preview" ]] || {
  echo "refusing publication from unexpected version: $VERSION" >&2
  exit 1
}
[[ "${AEGIS_PUBLISH_CONFIRM:-}" == "YES" ]] || {
  echo "refusing publication: set AEGIS_PUBLISH_CONFIRM=YES after reviewing the evidence" >&2
  exit 2
}
command -v gh >/dev/null || { echo "GitHub CLI (gh) is required" >&2; exit 1; }
gh auth status

git diff --quiet && git diff --cached --quiet || {
  echo "working tree must be clean" >&2
  exit 1
}

scripts/validate-v012-release.sh
scripts/build-release.sh

current="$(git branch --show-current)"
if [[ "$current" != "main" ]]; then
  git branch -M main
fi

if ! gh repo view "$REPO" >/dev/null 2>&1; then
  gh repo create "$REPO" --public --source=. --remote=origin \
    --description "Experimental exact coupled-bound shortest-path search for real road networks"
elif ! git remote get-url origin >/dev/null 2>&1; then
  git remote add origin "https://github.com/$REPO.git"
fi

if ! git rev-parse "$TAG" >/dev/null 2>&1; then
  git tag -a "$TAG" -m "Aegis ACBS $VERSION"
fi

git push -u origin main
git push origin "$TAG"

if gh release view "$TAG" >/dev/null 2>&1; then
  echo "release already exists: $TAG"
else
  gh release create "$TAG" dist/* \
    --title "Aegis ACBS $VERSION" \
    --notes-file RELEASE_NOTES.md \
    --prerelease
fi

echo "published repository and prerelease: https://github.com/$REPO"
