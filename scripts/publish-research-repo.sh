#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

PROFILE="${1:?usage: publish-research-repo.sh TRIGGER_PROFILE_JSON}"
REPO="${AEGIS_GITHUB_REPO:-lasder-ca/aegis-acbs}"

[[ "${AEGIS_PUBLISH_CONFIRM:-}" == "YES" ]] || {
  echo "refusing to publish: set AEGIS_PUBLISH_CONFIRM=YES after reviewing the repository and profile" >&2
  exit 2
}
command -v gh >/dev/null || { echo "GitHub CLI (gh) is required" >&2; exit 1; }
gh auth status

git diff --quiet && git diff --cached --quiet || {
  echo "working tree must be clean" >&2
  exit 1
}

python3 - "$PROFILE" <<'PY'
import json, sys
p=sys.argv[1]
r=json.load(open(p, encoding='utf-8'))
if not r.get('allCorrect'):
    raise SystemExit('refusing to publish: trigger profile contains an incorrect route')
if int(r.get('traceErrors', 0)) != 0:
    raise SystemExit('refusing to publish: trigger profile contains trace errors')
if int(r.get('unstableLabels', 0)) != 0:
    raise SystemExit('refusing to publish: replay-labelled traces were not deterministic')
if int(r.get('queries', 0)) < 10000:
    raise SystemExit('refusing to publish: fewer than 10,000 Tokyo queries were profiled')
print('profile accepted:', r['queries'], 'queries; selected rule:', bool(r.get('selectedRule')))
PY

# v0.11.2 publishes the research repository, not a GitHub Release.
VERSION="$(cat VERSION)"
if [[ "$VERSION" != "0.11.2-experimental" ]]; then
  echo "refusing repository-preview publication from unexpected version $VERSION" >&2
  exit 1
fi

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

git push -u origin main
git push origin --tags

echo "repository published: https://github.com/$REPO"
echo "GitHub Release intentionally deferred to v0.12.0-research-preview"
