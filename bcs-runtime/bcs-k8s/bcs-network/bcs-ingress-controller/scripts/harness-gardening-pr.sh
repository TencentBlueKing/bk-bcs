#!/usr/bin/env bash
# Lightweight harness-gardening PR check (mode=pr).
# Usage: ./scripts/harness-gardening-pr.sh [base_ref]
# Default base_ref: HEAD~1 (last commit). Use origin/main for PR diff.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

BASE_REF="${1:-}"
if [ -z "$BASE_REF" ]; then
  # Prefer working tree (staged + unstaged); fall back to last commit
  CHANGED="$(git diff --name-only HEAD 2>/dev/null; git diff --name-only --cached HEAD 2>/dev/null; git ls-files --others --exclude-standard 2>/dev/null)" 
  CHANGED="$(echo "$CHANGED" | sort -u | grep -v '^$' || true)"
  SCOPE="working tree"
  if [ -z "$CHANGED" ]; then
    BASE_REF="HEAD~1"
    CHANGED="$(git diff --name-only "$BASE_REF" HEAD 2>/dev/null || true)"
    SCOPE="commit $BASE_REF..HEAD"
  fi
else
  CHANGED="$(git diff --name-only "$BASE_REF" HEAD 2>/dev/null || true)"
  SCOPE="$BASE_REF..HEAD"
fi

if [ -z "$CHANGED" ]; then
  echo "[gardening-pr] No changes in $SCOPE, skip."
  exit 0
fi

echo "[gardening-pr] Checking $SCOPE ..."
WARN=0

# Dimension 1: harness doc link validity (sample)
if [ -f docs/harness/README.md ]; then
  while IFS= read -r doc; do
    while IFS= read -r link; do
      target="$ROOT/$(dirname "$doc")/$link"
      target="$(realpath -m "$target" 2>/dev/null || echo "$target")"
      if [ ! -e "$target" ]; then
        echo "[WARN] Broken link in $doc -> $link"
        WARN=$((WARN + 1))
      fi
    done < <(grep -oP '\]\(\K[^)#]+' "$doc" 2>/dev/null | grep -v '^https\?://' || true)
  done < <(find docs -name '*.md' 2>/dev/null)
fi

# Dimension 6/8: Go file changes suggest dev-map update
GO_CHANGED="$(echo "$CHANGED" | grep '\.go$' | grep -v '_test\.go$' || true)"
if [ -n "$GO_CHANGED" ]; then
  echo "[WARN] Go source changed; consider updating docs/dev-map/ (source-index, module-index)"
  echo "$GO_CHANGED" | sed 's/^/  - /'
  WARN=$((WARN + 1))
fi

# Secret / credential guard
if echo "$CHANGED" | grep -qE '\.token\.env|\.env$|credentials'; then
  echo "[WARN] Sensitive file in changeset — do NOT commit secrets (.token.env, .env)"
  WARN=$((WARN + 1))
fi

# Dimension 7: project.json sanity
if [ -f project.json ]; then
  if ! grep -q '"workspace_id"' project.json || ! grep -q '"owner"' project.json; then
    echo "[WARN] project.json missing workspace_id or owner"
    WARN=$((WARN + 1))
  fi
else
  echo "[WARN] project.json not found (TAPD pipeline needs it)"
  WARN=$((WARN + 1))
fi

if [ "$WARN" -gt 0 ]; then
  echo "[gardening-pr] $WARN warning(s). Run full scan: say '文档巡检' to agent."
  exit 0
fi

echo "[gardening-pr] OK"
exit 0
