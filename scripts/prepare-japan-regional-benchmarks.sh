#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${AEGIS_BIN:-$ROOT/bin/aegis}"
PBF_DIR="${AEGIS_PBF_DIR:-$ROOT/.data/pbf}"
OSM_DIR="${AEGIS_OSM_DIR:-$ROOT/.data/regional-osm}"
GRAPH_DIR="${AEGIS_GRAPH_DIR:-$ROOT/.data/regional-graphs}"
CITIES="${AEGIS_CITIES:-tokyo-wide yokohama-wide osaka-wide nagoya-wide}"

command -v curl >/dev/null || { echo "curl is required" >&2; exit 1; }
command -v osmium >/dev/null || { echo "osmium-tool is required" >&2; exit 1; }
mkdir -p "$PBF_DIR" "$OSM_DIR" "$GRAPH_DIR" "$ROOT/bin"
[[ -x "$BIN" ]] || (cd "$ROOT" && go build -trimpath -o "$BIN" ./cmd/aegis)

region_url() {
  case "$1" in
    kanto)  echo "https://download.geofabrik.de/asia/japan/kanto-latest.osm.pbf" ;;
    kansai) echo "https://download.geofabrik.de/asia/japan/kansai-latest.osm.pbf" ;;
    chubu)  echo "https://download.geofabrik.de/asia/japan/chubu-latest.osm.pbf" ;;
    *) echo "unknown Geofabrik region: $1" >&2; return 2 ;;
  esac
}

city_spec() {
  case "$1" in
    tokyo-wide)    echo "kanto|139.62,35.56,139.90,35.82" ;;
    yokohama-wide) echo "kanto|139.45,35.30,139.78,35.62" ;;
    osaka-wide)    echo "kansai|135.35,34.55,135.75,34.85" ;;
    nagoya-wide)   echo "chubu|136.75,35.02,137.08,35.30" ;;
    *) echo "unknown city: $1" >&2; return 2 ;;
  esac
}

download_region() {
  local region="$1" url dest
  url="$(region_url "$region")"
  dest="$PBF_DIR/$region-latest.osm.pbf"
  if osmium fileinfo -e "$dest" >/dev/null 2>&1; then
    echo "[$region] using cached $dest"
    return
  fi
  rm -f "$dest.tmp"
  curl --fail --location --retry 5 --retry-all-errors --connect-timeout 30 --progress-bar "$url" -o "$dest.tmp"
  mv "$dest.tmp" "$dest"
  sha256sum "$dest" > "$dest.sha256"
}

prepare_city() {
  local city="$1" spec region bbox source_pbf city_pbf roads_pbf roads_osm
  spec="$(city_spec "$city")"
  region="${spec%%|*}"
  bbox="${spec#*|}"
  download_region "$region"
  source_pbf="$PBF_DIR/$region-latest.osm.pbf"
  city_pbf="$OSM_DIR/$city.osm.pbf"
  roads_pbf="$OSM_DIR/$city-roads.osm.pbf"
  roads_osm="$OSM_DIR/$city-roads.osm"
  echo "[$city] extracting bbox $bbox"
  osmium extract --bbox="$bbox" --strategy=complete_ways --overwrite "$source_pbf" -o "$city_pbf"
  osmium tags-filter "$city_pbf" w/highway --overwrite -o "$roads_pbf"
  osmium cat "$roads_pbf" --output-format=osm --overwrite -o "$roads_osm"
  for metric in distance time; do
    "$BIN" import-osm --input "$roads_osm" --output "$GRAPH_DIR/$city-car-$metric.aegis" --name "$city" --profile car --metric "$metric"
  done
}

for city in $CITIES; do
  prepare_city "$city"
done

cat > "$GRAPH_DIR/DATA-MANIFEST.txt" <<MANIFEST
Generated: $(date -u +%Y-%m-%dT%H:%M:%SZ)
Source: Geofabrik regional PBF extracts derived from OpenStreetMap
Attribution: © OpenStreetMap contributors
Licence: ODbL 1.0
Configuration: $ROOT/benchmarks/japan-regions.json
MANIFEST

ls -lh "$GRAPH_DIR"/*.aegis
