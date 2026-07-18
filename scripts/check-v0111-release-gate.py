#!/usr/bin/env python3
"""Select a v0.11.1 connection-guard candidate using replay and 10k benchmark gates.

Usage:
  check-v0111-release-gate.py REPLAY_JSON BENCHMARK_JSON [OUTPUT_JSON]
"""
from __future__ import annotations

import json
import sys
from pathlib import Path

CANDIDATES = ("aegis-connect-32", "aegis-connect-40", "aegis-connect-32x16")
MAX_GLOBAL_REGRESSION = 0.01
MIN_SCHEDULER_TAIL_GAIN_NS = 500_000
MAX_PERSISTENT_TAIL_REGRESSION_NS = 1_000_000
MAX_WORK_REGRESSION = 0.01


def load(path: str) -> dict:
    return json.loads(Path(path).read_text(encoding="utf-8"))


def summary(report: dict, algorithm: str) -> dict:
    for row in report.get("summary", []):
        if row.get("algorithm") == algorithm:
            return row
    raise KeyError(f"missing summary for {algorithm}")


def guard(case: dict, algorithm: str) -> dict:
    for row in case.get("guards", []):
        if row.get("algorithm") == algorithm:
            return row
    raise KeyError(f"missing replay guard {algorithm}")


def ratio(candidate: int, baseline: int) -> float:
    if baseline <= 0:
        raise ValueError("baseline must be positive")
    return candidate / baseline


def evaluate(replay: dict, benchmark: dict, candidate: str) -> dict:
    failures: list[str] = []
    scheduler = [c for c in replay.get("cases", []) if c.get("classification") == "adaptive-scheduler-tail"]
    persistent = [c for c in replay.get("cases", []) if c.get("classification") == "persistent-classical-tail"]
    scheduler_gains: list[int] = []
    persistent_regressions: list[int] = []

    if not scheduler:
        failures.append("no reproduced adaptive-scheduler-tail case")
    for case in scheduler:
        row = guard(case, candidate)
        gain = int(row.get("advantageNs", 0))
        scheduler_gains.append(gain)
        if gain < MIN_SCHEDULER_TAIL_GAIN_NS:
            failures.append(
                f"scheduler-tail {case.get('sourceReport')} query={case.get('queryIndex')} "
                f"gain {gain/1e6:.3f}ms < 0.500ms"
            )
    for case in persistent:
        row = guard(case, candidate)
        regression = int(row.get("regressionNs", 0))
        persistent_regressions.append(regression)
        if regression >= MAX_PERSISTENT_TAIL_REGRESSION_NS:
            failures.append(
                f"persistent-tail {case.get('sourceReport')} query={case.get('queryIndex')} "
                f"regression {regression/1e6:.3f}ms >= 1.000ms"
            )

    base = summary(benchmark, "aegis")
    cand = summary(benchmark, candidate)
    if int(cand.get("correct", 0)) != int(cand.get("runs", -1)):
        failures.append("candidate benchmark is not fully correct")

    timing = {}
    for label, key in (("mean", "meanNs"), ("median", "medianNs"), ("p95", "p95Ns")):
        r = ratio(int(cand.get(key, 0)), int(base.get(key, 0)))
        timing[label] = r
        if r > 1 + MAX_GLOBAL_REGRESSION:
            failures.append(f"global {label} regression {(r-1)*100:.2f}% > 1.00%")

    work = {}
    for label, key in (("relaxed", "medianRelaxed"), ("expanded", "medianExpanded")):
        b = int(base.get(key, 0))
        c = int(cand.get(key, 0))
        r = ratio(c, b) if b > 0 else 1.0
        work[label] = r
        if r > 1 + MAX_WORK_REGRESSION:
            failures.append(f"global {label} regression {(r-1)*100:.2f}% > 1.00%")

    min_gain = min(scheduler_gains) if scheduler_gains else 0
    max_persistent = max(persistent_regressions) if persistent_regressions else 0
    # Higher is better: prioritize scheduler-tail repair, then p95 and work stability.
    score = (min_gain / 1e6) - 10 * max(0.0, timing["p95"] - 1) - 5 * max(0.0, work["expanded"] - 1)
    return {
        "algorithm": candidate,
        "pass": not failures,
        "failures": failures,
        "minSchedulerTailGainNs": min_gain,
        "maxPersistentRegressionNs": max_persistent,
        "timingRatios": timing,
        "workRatios": work,
        "score": score,
    }


def main() -> int:
    if len(sys.argv) not in (3, 4):
        print(__doc__.strip(), file=sys.stderr)
        return 2
    replay = load(sys.argv[1])
    benchmark = load(sys.argv[2])
    all_correct = bool(replay.get("allCorrect")) and bool(benchmark.get("allCorrect"))
    results = []
    for candidate in CANDIDATES:
        try:
            result = evaluate(replay, benchmark, candidate)
        except (KeyError, ValueError) as exc:
            result = {"algorithm": candidate, "pass": False, "failures": [str(exc)], "score": -1e9}
        if not all_correct:
            result["pass"] = False
            result.setdefault("failures", []).append("replay or benchmark contains an incorrect route")
        results.append(result)

    passing = [r for r in results if r["pass"]]
    selected = max(passing, key=lambda r: r["score"])["algorithm"] if passing else None
    report = {
        "version": "v0.11.1-connection-guard-gate-v1",
        "allCorrect": all_correct,
        "selected": selected,
        "pass": selected is not None,
        "candidates": results,
    }

    print("v0.11.1 connection-guard release gate")
    for result in results:
        timing = result.get("timingRatios", {})
        work = result.get("workRatios", {})
        print(
            f"  {result['algorithm']}: {'PASS' if result['pass'] else 'FAIL'} "
            f"gain={result.get('minSchedulerTailGainNs', 0)/1e6:.3f}ms "
            f"persistent-regression={result.get('maxPersistentRegressionNs', 0)/1e6:.3f}ms "
            f"mean={timing.get('mean', 0):.4f}x median={timing.get('median', 0):.4f}x "
            f"p95={timing.get('p95', 0):.4f}x relaxed={work.get('relaxed', 0):.4f}x "
            f"expanded={work.get('expanded', 0):.4f}x"
        )
        for failure in result.get("failures", []):
            print(f"    - {failure}")
    if selected:
        print(f"SELECTED: {selected}")
        print("RELEASE GATE: PASS")
    else:
        print("SELECTED: none")
        print("RELEASE GATE: FAIL")

    if len(sys.argv) == 4:
        Path(sys.argv[3]).write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    return 0 if selected else 1


if __name__ == "__main__":
    raise SystemExit(main())
