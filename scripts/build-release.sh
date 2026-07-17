#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
VERSION="$(cat VERSION)"
DIST="$ROOT/dist"
rm -rf "$DIST"
mkdir -p "$DIST" "$ROOT/artifacts"

go test ./...
go vet ./...
go test -race ./internal/search ./internal/graph ./internal/bench ./internal/server
scripts/reproduce.sh
OLD_TAG=v0.5.0-experimental QUERIES=100 REPEATS=5 BATCH=8 scripts/compare-tags.sh artifacts/hatfield-uk.aegis artifacts/tag-comparison

build_one() {
  local os="$1" arch="$2" ext=""
  [[ "$os" == windows ]] && ext=".exe"
  local base="aegis-acbs_${VERSION}_${os}_${arch}"
  local stage="$(mktemp -d)"
  mkdir -p "$stage/$base"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X github.com/lasder-ca/aegis-acbs/internal/version.Version=$VERSION" -o "$stage/$base/aegis$ext" ./cmd/aegis
  cp README.md LICENSE RELEASE_NOTES.md "$stage/$base/"
  cp -r docs benchmarks scripts benchdata "$stage/$base/"
  rm -f "$stage/$base/benchdata/hatfield-uk.aegis"
  if [[ "$os" == windows ]]; then
    (cd "$stage" && zip -qr "$DIST/$base.zip" "$base")
  else
    tar -C "$stage" -czf "$DIST/$base.tar.gz" "$base"
  fi
  rm -rf "$stage"
}

for target in linux/amd64 linux/arm64 windows/amd64 darwin/amd64; do
  build_one "${target%/*}" "${target#*/}"
done

git archive --format=tar.gz --prefix="aegis-acbs-$VERSION/" -o "$DIST/aegis-acbs-$VERSION-source.tar.gz" HEAD
git archive --format=zip --prefix="aegis-acbs-$VERSION/" -o "$DIST/aegis-acbs-$VERSION-source.zip" HEAD
cp artifacts/hatfield-uk-benchmark.json artifacts/hatfield-uk-benchmark.html "$DIST/"
cp artifacts/matrix/benchmark-matrix.json artifacts/matrix/benchmark-matrix.csv artifacts/matrix/benchmark-matrix.html "$DIST/"
cp artifacts/tag-comparison/summary.md "$DIST/tag-comparison.md"
cp artifacts/tag-comparison/old.html "$DIST/tag-comparison-v0.5.0.html"
cp artifacts/tag-comparison/current.html "$DIST/tag-comparison-v0.6.0.html"

OLD_TAG=v0.5.0-experimental BENCHTIME=20x COUNT=3 scripts/compare-allocations.sh artifacts/allocation-comparison
cp artifacts/allocation-comparison/summary.json artifacts/allocation-comparison/summary.md "$DIST/"
cp artifacts/allocation-comparison/old.txt "$DIST/allocation-v0.5.0.txt"
cp artifacts/allocation-comparison/current.txt "$DIST/allocation-v0.6.0.txt"

python3 - <<PY
import json, os, platform, subprocess, datetime
root=${ROOT@Q}; dist=${DIST@Q}; version=${VERSION@Q}
commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip()
info={
  'name':'Aegis ACBS','version':version,'commit':commit,
  'builtAt':datetime.datetime.now(datetime.timezone.utc).isoformat(),
  'builder':{'platform':platform.platform(),'python':platform.python_version()},
  'tests':'go test ./...','vet':'go vet ./...','benchmark':'Hatfield real OSM-derived fixture with deterministic interleaved order, ACBS ablations, allocation telemetry, a three-seed distance/time matrix, a v0.5/v0.6 tag comparison, and an isolated v0.5/v0.6 allocation regression comparison'
}
open(os.path.join(dist,'BUILD-INFO.json'),'w').write(json.dumps(info,indent=2)+'\n')
sbom={'bomFormat':'CycloneDX','specVersion':'1.5','serialNumber':'urn:uuid:aegis-acbs-'+version,'version':1,'metadata':{'component':{'type':'application','name':'aegis-acbs','version':version}},'components':[]}
open(os.path.join(dist,'SBOM.cdx.json'),'w').write(json.dumps(sbom,indent=2)+'\n')
PY

if git rev-parse --git-dir >/dev/null 2>&1; then
  git bundle create "$DIST/aegis-acbs-$VERSION.bundle" --all
fi

(cd "$DIST" && sha256sum $(find . -maxdepth 1 -type f ! -name SHA256SUMS ! -name '*complete-release.zip' -printf '%f\n' | sort) > SHA256SUMS)
(
  cd "$DIST"
  zip -q "aegis-acbs-v$VERSION-complete-release.zip" \
    $(find . -maxdepth 1 -type f ! -name '*complete-release.zip' -printf '%f\n' | sort)
  sha256sum "aegis-acbs-v$VERSION-complete-release.zip" >> SHA256SUMS
)
