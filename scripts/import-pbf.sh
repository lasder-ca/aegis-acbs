#!/usr/bin/env bash
set -euo pipefail
# PBF support deliberately remains outside the core binary to keep the release
# dependency-free. osmium-tool performs the standards-compliant PBF decoding.
if [[ $# -lt 2 ]]; then
  echo "usage: $0 INPUT.osm.pbf OUTPUT.aegis [car|bike|walk] [distance|time]" >&2
  exit 2
fi
command -v osmium >/dev/null || { echo "osmium-tool is required: sudo apt install osmium-tool" >&2; exit 1; }
command -v aegis >/dev/null || { echo "aegis must be on PATH" >&2; exit 1; }
input="$1" output="$2" profile="${3:-car}" metric="${4:-distance}"
tmp="$(mktemp --suffix=.osm)"
trap 'rm -f "$tmp"' EXIT
osmium tags-filter "$input" w/highway -f osm -o "$tmp"
aegis import-osm --input "$tmp" --output "$output" --profile "$profile" --metric "$metric"
