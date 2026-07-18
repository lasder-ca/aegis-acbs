#!/usr/bin/env python3
"""Validate the Tokyo evidence required by v0.12.0-research-preview."""

from __future__ import annotations

import json
import math
import sys
from pathlib import Path
from typing import Any


def load(path: Path) -> dict[str, Any]:
    try:
        return json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError as exc:
        raise SystemExit(f"missing evidence: {path}") from exc


def require(condition: bool, message: str) -> None:
    if not condition:
        raise SystemExit(f"evidence check failed: {message}")


def close(value: float, expected: float, tolerance: float = 1e-9) -> bool:
    return math.isclose(float(value), expected, rel_tol=tolerance, abs_tol=tolerance)


def main() -> int:
    if len(sys.argv) != 2:
        raise SystemExit("usage: check-v012-release-evidence.py EVIDENCE_DIR")

    root = Path(sys.argv[1])
    validation = load(root / "regret-validation.json")
    replay = load(root / "regret-replay.json")
    guard = load(root / "rejected-guards-gate.json")
    profile = load(root / "trigger-profile.json")

    require(int(validation.get("totalQueries", 0)) == 10000, "tail validation must contain 10,000 queries")
    require(int(validation.get("totalMeaningful", -1)) == 11, "expected 11 initially meaningful slowdowns")
    require(bool(validation.get("allCorrect")), "tail validation contains an incorrect route")

    require(int(replay.get("requestedCases", 0)) == 11, "replay must request 11 cases")
    require(int(replay.get("replayedCases", 0)) == 11, "replay must complete 11 cases")
    require(int(replay.get("reproducedMeaningful", -1)) == 2, "expected two reproduced meaningful cases")
    require(int(replay.get("adaptiveSchedulerTail", -1)) == 1, "expected one scheduler tail")
    require(int(replay.get("persistentClassical", -1)) == 1, "expected one persistent classical tail")
    require(int(replay.get("notReproduced", -1)) == 9, "expected nine non-reproduced cases")
    require(bool(replay.get("allCorrect")), "replay contains an incorrect route")

    require(not bool(guard.get("pass")), "rejected guard gate unexpectedly passed")
    require(not guard.get("selected"), "rejected guard gate unexpectedly selected a candidate")

    require(int(profile.get("queries", 0)) == 10000, "trigger profile must contain 10,000 queries")
    require(int(profile.get("schedulerTails", -1)) == 1, "expected one scheduler-tail label")
    require(int(profile.get("persistentTails", -1)) == 1, "expected one persistent-tail label")
    require(int(profile.get("traceErrors", -1)) == 0, "trigger profile contains trace errors")
    require(int(profile.get("unstableLabels", -1)) == 0, "trigger profile contains unstable labels")
    require(bool(profile.get("allCorrect")), "trigger profile contains an incorrect route")

    selected = profile.get("selectedRule")
    require(isinstance(selected, dict), "narrow diagnostic trigger was not selected")
    require(int(selected.get("checkpoint", 0)) == 48, "selected checkpoint must be 48")
    require(int(selected.get("matches", 0)) == 1, "selected rule must match one query")
    require(int(selected.get("positiveMatches", 0)) == 1, "selected rule must match the scheduler tail")
    require(int(selected.get("falsePositives", -1)) == 0, "selected rule must have zero in-suite false positives")

    conditions = selected.get("conditions") or []
    require(len(conditions) == 1, "selected rule must contain one condition")
    condition = conditions[0]
    require(condition.get("feature") == "switchRate", "selected feature must be switchRate")
    require(condition.get("operator") == "<=", "selected operator must be <=")
    require(close(condition.get("threshold", 0), 0.4583333333333333, 1e-7), "unexpected switchRate threshold")

    print("v0.12.0 research-preview evidence: PASS")
    print("  exact routes: 10,000/10,000")
    print("  initial meaningful tails: 11")
    print("  isolated replay: scheduler=1 persistent=1 not-reproduced=9")
    print("  rejected guard selected: none")
    print("  diagnostic rule: checkpoint=48 switchRate<=0.458333333 matches=1 false-positive=0")
    print("  promotion: diagnostic only")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
