#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
export AEGIS_QUERIES="${AEGIS_QUERIES:-1000}"
export AEGIS_REPEATS="${AEGIS_REPEATS:-3}"
export AEGIS_BATCH="${AEGIS_BATCH:-1}"
export AEGIS_ORDER="${AEGIS_ORDER:-interleaved}"
export AEGIS_MEASURE_MEMORY="${AEGIS_MEASURE_MEMORY:-1}"
export AEGIS_SEEDS="${AEGIS_SEEDS:-1010 20260717 424242 8675309 123456789 314159265 271828182 161803398 141421356 173205080}"
export AEGIS_REPORT_DIR="${AEGIS_REPORT_DIR:-$ROOT/artifacts/research-validation}"
export GOMAXPROCS="${GOMAXPROCS:-1}"

"$ROOT/scripts/run-japan-matrix.sh"
echo "validation matrix: $AEGIS_REPORT_DIR/benchmark-matrix.html"
