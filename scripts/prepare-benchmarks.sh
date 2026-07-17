#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DATA="${AEGIS_DATA_DIR:-$ROOT/.data/osm}"
GRAPHS="${AEGIS_GRAPH_DIR:-$ROOT/.data/graphs}"
mkdir -p "$GRAPHS" "$ROOT/bin"
cd "$ROOT"
go build -trimpath -ldflags "-s -w" -o bin/aegis ./cmd/aegis

if [[ ! -f "$DATA/tokyo.osm" ]]; then
  scripts/download-city-osm.sh all
fi
for city in tokyo london shanghai seoul paris; do
  for metric in distance time; do
    bin/aegis import-osm --input "$DATA/$city.osm" --output "$GRAPHS/$city-car-$metric.aegis" --name "$city" --profile car --metric "$metric"
  done
done
