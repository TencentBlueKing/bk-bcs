#!/usr/bin/env python3
"""Local gate checker for code-review (no shared runtime; skill-private)."""

from __future__ import annotations

import argparse
import sys
from pathlib import Path

SKILL_DIR = Path(__file__).resolve().parents[1]
KNOWN_ROUTES = frozenset(
    {
        "cr-auto",
        "cr-staged",
        "cr-commit",
        "cr-ai",
        "cr-confidence",
        "cr-report",
        "cr-example",
    }
)


def check_skill_docs() -> list[str]:
    errors: list[str] = []
    for name in ("routing.md", "gates.md", "checklist.md"):
        path = SKILL_DIR / "references" / name
        if not path.is_file():
            errors.append(f"missing {path.relative_to(SKILL_DIR)}")
    return errors


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="code-review local gate check")
    parser.add_argument("--route", required=True, help="routing id from routing.md")
    args = parser.parse_args(argv)

    errors = check_skill_docs()
    if args.route not in KNOWN_ROUTES:
        errors.append(
            f"unknown route '{args.route}'; known: {', '.join(sorted(KNOWN_ROUTES))}"
        )

    if errors:
        print(f"FAIL: code-review gates (route={args.route})")
        for err in errors:
            print(f"  - {err}")
        return 1

    print(f"PASS: code-review gates route={args.route}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
