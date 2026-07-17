#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
"$ROOT/scripts/prepare-japan-regional-benchmarks.sh"
AEGIS_GRAPH_DIR="${AEGIS_GRAPH_DIR:-$ROOT/.data/regional-graphs}" "$ROOT/scripts/benchmark-matrix.sh"
