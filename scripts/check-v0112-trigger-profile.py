#!/usr/bin/env python3
"""Summarize the v0.11.2 whole-suite trigger profile without promoting it."""

from __future__ import annotations

import json
import sys
from pathlib import Path


def main() -> int:
    if len(sys.argv) != 2:
        raise SystemExit("usage: check-v0112-trigger-profile.py trigger-profile.json")
    path = Path(sys.argv[1])
    report = json.loads(path.read_text())
    selected = report.get("selectedRule")
    queries = int(report.get("queries", 0))
    scheduler = int(report.get("schedulerTails", 0))
    persistent = int(report.get("persistentTails", 0))
    correct = bool(report.get("allCorrect", False))
    errors = int(report.get("traceErrors", 0))
    unstable = int(report.get("unstableLabels", 0))

    print("v0.11.2 trigger-profile gate")
    print(f"  queries: {queries}")
    print(f"  scheduler tails: {scheduler}")
    print(f"  persistent tails: {persistent}")
    print(f"  trace errors: {errors}")
    print(f"  unstable labels: {unstable}")
    print(f"  all correct: {correct}")
    if selected:
        print(
            "  selected: checkpoint={checkpoint} matches={matches} "
            "positives={positiveMatches} false-positives={falsePositives}".format(**selected)
        )
        for condition in selected.get("conditions", []):
            print(
                f"    {condition['feature']} {condition['operator']} "
                f"{condition['threshold']:.9g}"
            )
    else:
        print("  selected: none")

    profile_valid = correct and errors == 0 and unstable == 0 and queries > 0
    narrow = profile_valid and selected is not None
    print(f"PROFILE VALID: {'PASS' if profile_valid else 'FAIL'}")
    print(f"NARROW TRIGGER: {'FOUND' if narrow else 'NOT FOUND'}")
    print("PROMOTION: diagnostic only; requires independent seed/city validation")
    return 0 if profile_valid else 1


if __name__ == "__main__":
    raise SystemExit(main())
