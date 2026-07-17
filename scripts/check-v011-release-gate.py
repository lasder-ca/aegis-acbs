#!/usr/bin/env python3
"""Validate the v0.11 Tokyo late-guard release gates.

Usage:
  scripts/check-v011-release-gate.py REGRET_REPLAY_JSON BENCHMARK_JSON

The benchmark JSON must contain both `aegis` and `aegis-late-guard` summaries.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

MAX_GLOBAL_REGRESSION = 0.01
MIN_SCHEDULER_TAIL_GAIN_NS = 500_000
MAX_PERSISTENT_TAIL_REGRESSION_NS = 1_000_000


def load(path: str) -> dict:
    return json.loads(Path(path).read_text(encoding="utf-8"))


def summary(report: dict, algorithm: str) -> dict:
    for row in report.get("summary", []):
        if row.get("algorithm") == algorithm:
            return row
    raise SystemExit(f"missing summary for {algorithm}")


def ratio(candidate: int, baseline: int) -> float:
    if baseline <= 0:
        raise SystemExit("baseline timing must be positive")
    return candidate / baseline


def main() -> int:
    if len(sys.argv) != 3:
        print(__doc__.strip(), file=sys.stderr)
        return 2
    replay = load(sys.argv[1])
    benchmark = load(sys.argv[2])
    failures: list[str] = []

    if not replay.get("allCorrect", False):
        failures.append("replay contains an incorrect route")
    if not benchmark.get("allCorrect", False):
        failures.append("benchmark contains an incorrect route")

    scheduler_tails = [c for c in replay.get("cases", []) if c.get("classification") == "adaptive-scheduler-tail"]
    persistent_tails = [c for c in replay.get("cases", []) if c.get("classification") == "persistent-classical-tail"]
    if not scheduler_tails:
        failures.append("no reproduced adaptive-scheduler-tail case was present")
    for case in scheduler_tails:
        gain = int(case.get("lateGuardAdvantageNs", 0))
        if gain < MIN_SCHEDULER_TAIL_GAIN_NS:
            failures.append(
                f"scheduler-tail {case.get('sourceReport')} query={case.get('queryIndex')} "
                f"late guard gain {gain/1e6:.3f}ms is below 0.500ms"
            )
    for case in persistent_tails:
        regression = int(case.get("lateGuardRegressionNs", 0))
        if regression >= MAX_PERSISTENT_TAIL_REGRESSION_NS:
            failures.append(
                f"persistent-tail {case.get('sourceReport')} query={case.get('queryIndex')} "
                f"late guard regression {regression/1e6:.3f}ms is >= 1.000ms"
            )

    base = summary(benchmark, "aegis")
    guard = summary(benchmark, "aegis-late-guard")
    comparisons = {
        "mean": (int(base.get("meanNs", 0)), int(guard.get("meanNs", 0))),
        "median": (int(base.get("medianNs", 0)), int(guard.get("medianNs", 0))),
        "p95": (int(base.get("p95Ns", 0)), int(guard.get("p95Ns", 0))),
    }
    print("v0.11 release gate")
    for name, (old, new) in comparisons.items():
        r = ratio(new, old)
        print(f"  {name:6s}: aegis={old/1e6:.3f}ms guard={new/1e6:.3f}ms ratio={r:.4f}x")
        if r > 1 + MAX_GLOBAL_REGRESSION:
            failures.append(f"global {name} regression {(r-1)*100:.2f}% exceeds 1.00%")

    print(f"  scheduler tails: {len(scheduler_tails)}")
    print(f"  persistent tails: {len(persistent_tails)}")
    if failures:
        print("RELEASE GATE: FAIL")
        for failure in failures:
            print(f"  - {failure}")
        return 1
    print("RELEASE GATE: PASS")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
