#!/usr/bin/env bash
# sync-standards-rules.sh — render IDE standards-* rules from docs/standards/README.md 选用表.
# Shared by harness-generating and harness-gardening. See references/standards-compliance.md.
# Usage: sync-standards-rules.sh [--dry-run] <workspace_root>
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ASSET_DIR="$(cd "$SCRIPT_DIR/../assets/ide-rules" && pwd)"
DRY=0
ROOT=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY=1; shift ;;
    -h|--help)
      echo "Usage: sync-standards-rules.sh [--dry-run] <workspace_root>"
      exit 0
      ;;
    *)
      ROOT="$1"; shift ;;
  esac
done

ROOT="${ROOT:-.}"
ROOT="$(cd "$ROOT" && pwd)"
README="$ROOT/docs/standards/README.md"

if [[ ! -f "$README" ]]; then
  echo "sync-standards-rules: skip (no $README)"
  exit 0
fi

# Parse selected standard basenames from 「当前项目选用」 section links.
parse_selected() {
  local in=0 line fname
  SELECTED=()
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" == "## 当前项目选用"* ]]; then
      in=1
      continue
    fi
    if [[ "$in" -eq 1 && "$line" == "## "* ]]; then
      break
    fi
    [[ "$in" -eq 1 ]] || continue
    # markdown link: [text](path.md) or `path.md`
    while IFS= read -r fname; do
      [[ -n "$fname" ]] || continue
      fname="${fname##*/}"
      SELECTED+=("$fname")
    done < <(printf '%s\n' "$line" | sed -nE 's/.*\]\(([^)]+\.md)\).*/\1/p; s/.*`([^`]+\.md)`.*/\1/p')
  done < "$README"
}

category_of() {
  case "$1" in
    frontend-*) echo frontend ;;
    api-*) echo api ;;
    backend-*) echo backend ;;
    security-*) echo security ;;
    *) echo "" ;;
  esac
}

resolve_targets() {
  TARGETS=()
  local ide p t
  local fallback=0
  for ide in cursor codebuddy claude codex; do
    p="$ROOT/.$ide"
    if [[ -L "$p" ]]; then
      t="$(readlink "$p" || true)"
      if [[ "$t" == ".agents" || "$t" == */.agents || "$t" == ".agents/"* ]]; then
        fallback=1
      fi
    fi
  done

  if [[ "$fallback" -eq 1 ]]; then
    # .md 带 CodeBuddy frontmatter + Claude paths，供 .codebuddy/.claude → .agents 共用
    TARGETS+=("mdc|$ROOT/.agents/rules")
    TARGETS+=("md|$ROOT/.agents/rules")
    return
  fi

  if [[ -d "$ROOT/.cursor" && ! -L "$ROOT/.cursor" ]]; then
    TARGETS+=("mdc|$ROOT/.cursor/rules")
  fi
  if [[ -d "$ROOT/.codebuddy" && ! -L "$ROOT/.codebuddy" ]]; then
    TARGETS+=("md|$ROOT/.codebuddy/rules")
  fi
  if [[ -d "$ROOT/.claude" && ! -L "$ROOT/.claude" ]]; then
    TARGETS+=("claude|$ROOT/.claude/rules")
  fi
  # Codex：无模块化 agent instruction rules（.codex/rules 为 execpolicy Starlark）；门闩走 AGENTS.md
  if [[ ${#TARGETS[@]} -eq 0 && -d "$ROOT/.agents" ]]; then
    TARGETS+=("mdc|$ROOT/.agents/rules")
    TARGETS+=("md|$ROOT/.agents/rules")
  fi
}

render_one() {
  local format="$1" dest_dir="$2" base="$3" std_path="$4"
  local tpl out body
  case "$format" in
    mdc)
      tpl="$ASSET_DIR/cursor/${base}.mdc.template"
      out="$dest_dir/${base}.mdc"
      ;;
    claude)
      tpl="$ASSET_DIR/claude/${base}.md.template"
      out="$dest_dir/${base}.md"
      ;;
    *)
      tpl="$ASSET_DIR/codebuddy/${base}.md.template"
      out="$dest_dir/${base}.md"
      ;;
  esac
  [[ -f "$tpl" ]] || return 0
  body="$(sed "s|\${STANDARD_PATH}|${std_path}|g" "$tpl")"
  if [[ "$DRY" -eq 1 ]]; then
    echo "DRY write $out"
    return
  fi
  mkdir -p "$dest_dir"
  printf '%s\n' "$body" > "$out"
  echo "wrote $out"
}

remove_stale() {
  local dest_dir="$1" keep_list="$2"
  local f base
  [[ -d "$dest_dir" ]] || return 0
  shopt -s nullglob
  for f in "$dest_dir"/standards-*.mdc "$dest_dir"/standards-*.md; do
    base="$(basename "$f")"
    base="${base%.mdc}"
    base="${base%.md}"
    if [[ "$keep_list" != *"|$base|"* ]]; then
      if [[ "$DRY" -eq 1 ]]; then
        echo "DRY rm $f"
      else
        rm -f "$f"
        echo "removed $f"
      fi
    fi
  done
  shopt -u nullglob
}

parse_selected
resolve_targets

if [[ ${#TARGETS[@]} -eq 0 ]]; then
  echo "sync-standards-rules: skip rules write (no .cursor/.codebuddy/.claude/.agents layout)"
  exit 0
fi

FE_PATH="" API_PATH="" BE_PATH="" SEC_PATH=""
for fname in "${SELECTED[@]+"${SELECTED[@]}"}"; do
  case "$(category_of "$fname")" in
    frontend) [[ -z "$FE_PATH" ]] && FE_PATH="docs/standards/$fname" ;;
    api) [[ -z "$API_PATH" ]] && API_PATH="docs/standards/$fname" ;;
    backend) [[ -z "$BE_PATH" ]] && BE_PATH="docs/standards/$fname" ;;
    security) [[ -z "$SEC_PATH" ]] && SEC_PATH="docs/standards/$fname" ;;
  esac
done

NEED_GATE=0
KEEP="|"
if [[ -n "$FE_PATH$API_PATH$BE_PATH$SEC_PATH" ]]; then
  NEED_GATE=1
  KEEP="|standards-gate|"
fi
[[ -n "$FE_PATH" ]] && KEEP+="standards-frontend|"
[[ -n "$API_PATH" ]] && KEEP+="standards-api|"
[[ -n "$BE_PATH" ]] && KEEP+="standards-backend|"
[[ -n "$SEC_PATH" ]] && KEEP+="standards-security|"

for entry in "${TARGETS[@]}"; do
  format="${entry%%|*}"
  dest="${entry#*|}"
  if [[ "$NEED_GATE" -eq 1 ]]; then
    render_one "$format" "$dest" "standards-gate" "docs/standards/README.md"
  fi
  [[ -n "$FE_PATH" ]] && render_one "$format" "$dest" "standards-frontend" "$FE_PATH"
  [[ -n "$API_PATH" ]] && render_one "$format" "$dest" "standards-api" "$API_PATH"
  [[ -n "$BE_PATH" ]] && render_one "$format" "$dest" "standards-backend" "$BE_PATH"
  [[ -n "$SEC_PATH" ]] && render_one "$format" "$dest" "standards-security" "$SEC_PATH"
  remove_stale "$dest" "$KEEP"
done

echo "sync-standards-rules: done keep=$KEEP"
exit 0
