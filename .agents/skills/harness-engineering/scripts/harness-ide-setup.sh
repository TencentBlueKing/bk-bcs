#!/usr/bin/env bash
# harness-ide-setup.sh
# 为接入仓写入 IDE graphify 集成：Cursor/CodeBuddy 双格式 Rules + CodeBuddy hook-guard。
# Rules 落盘路径对齐 install-to-target / standards-compliance（与 sync-standards-rules.sh 一致）。
# 幂等可重复执行；.codebuddy/settings.json 合并时需要 jq。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ASSET_DIR="$(cd "$SCRIPT_DIR/../assets/ide-rules" && pwd)"
WORKSPACE="${1:-.}"
WORKSPACE="$(cd "$WORKSPACE" && pwd)"
cd "$WORKSPACE"

# ── 1. 确定 graphify 路径 ───────────────────────────────────────────
GRAPHIFY_BIN=""
if command -v graphify &>/dev/null; then
    GRAPHIFY_BIN=$(command -v graphify)
else
    GRAPHIFY_BIN="graphify"
    echo "[WARN] graphify not found in PATH; writing placeholder 'graphify'"
fi

# ── 2. Rules 落盘目标（与 sync-standards-rules.sh 同一算法）──────────
resolve_rule_targets() {
  TARGETS=()
  local ide p t
  local fallback=0
  for ide in cursor codebuddy claude codex; do
    p="$WORKSPACE/.$ide"
    if [[ -L "$p" ]]; then
      t="$(readlink "$p" || true)"
      if [[ "$t" == ".agents" || "$t" == */.agents || "$t" == ".agents/"* ]]; then
        fallback=1
      fi
    fi
  done

  if [[ "$fallback" -eq 1 ]]; then
    TARGETS+=("mdc|$WORKSPACE/.agents/rules")
    TARGETS+=("md|$WORKSPACE/.agents/rules")
    return
  fi

  if [[ -d "$WORKSPACE/.cursor" && ! -L "$WORKSPACE/.cursor" ]]; then
    TARGETS+=("mdc|$WORKSPACE/.cursor/rules")
  fi
  if [[ -d "$WORKSPACE/.codebuddy" && ! -L "$WORKSPACE/.codebuddy" ]]; then
    TARGETS+=("md|$WORKSPACE/.codebuddy/rules")
  fi
  if [[ -d "$WORKSPACE/.claude" && ! -L "$WORKSPACE/.claude" ]]; then
    TARGETS+=("claude|$WORKSPACE/.claude/rules")
  fi
  # Codex：无模块化 agent instruction rules；graphify 门闩依赖 AGENTS.md / CLAUDE.md 侧
  if [[ ${#TARGETS[@]} -eq 0 && -d "$WORKSPACE/.agents" ]]; then
    TARGETS+=("mdc|$WORKSPACE/.agents/rules")
    TARGETS+=("md|$WORKSPACE/.agents/rules")
  fi
}

write_graphify_rule() {
  local format="$1" dest_dir="$2"
  local tpl out expected
  case "$format" in
    mdc)
      tpl="$ASSET_DIR/cursor/graphify.mdc.template"
      out="$dest_dir/graphify.mdc"
      ;;
    claude)
      tpl="$ASSET_DIR/claude/graphify.md.template"
      out="$dest_dir/graphify.md"
      ;;
    *)
      tpl="$ASSET_DIR/codebuddy/graphify.md.template"
      out="$dest_dir/graphify.md"
      ;;
  esac
  if [[ ! -f "$tpl" ]]; then
    echo "[WARN] missing template $tpl"
    return 0
  fi
  expected="$(cat "$tpl")"
  mkdir -p "$dest_dir"
  if [[ ! -f "$out" ]]; then
    printf '%s\n' "$expected" > "$out"
    echo "[OK] Created $out"
  elif [[ "$(cat "$out")" != "$expected" ]]; then
    printf '%s\n' "$expected" > "$out"
    echo "[UPDATED] Replaced outdated $out"
  else
    echo "[SKIP] $out already matches expected content"
  fi
}

resolve_rule_targets
if [[ ${#TARGETS[@]} -eq 0 ]]; then
  echo "[SKIP] graphify Rules — no .cursor/.codebuddy/.claude/.agents layout"
else
  for entry in "${TARGETS[@]}"; do
    write_graphify_rule "${entry%%|*}" "${entry#*|}"
  done
fi

# ── 3. CodeBuddy hook-guard（settings.json）──────────────────────────
# 与旧逻辑一致：.codebuddy 存在，或 .cursor 不存在时写入（软链 fallback 时
# .codebuddy → .agents，settings 落在 .agents/settings.json）。
if [[ -d "$WORKSPACE/.codebuddy" ]] || [[ ! -d "$WORKSPACE/.cursor" ]]; then
    mkdir -p "$WORKSPACE/.codebuddy"
    SETTINGS="$WORKSPACE/.codebuddy/settings.json"

    HOOK_SEARCH='{"matcher":"Bash|Grep","hooks":[{"type":"command","command":"GRAPHIFY_OUT=\"$(git rev-parse --show-toplevel)/docs/dev-map\" '"$GRAPHIFY_BIN"' hook-guard search"}]}'
    HOOK_READ='{"matcher":"Read|Glob","hooks":[{"type":"command","command":"GRAPHIFY_OUT=\"$(git rev-parse --show-toplevel)/docs/dev-map\" '"$GRAPHIFY_BIN"' hook-guard read"}]}'

    if [[ ! -f "$SETTINGS" ]]; then
        printf '{\n  "hooks": {\n    "PreToolUse": [\n      %s,\n      %s\n    ]\n  }\n}\n' \
            "$HOOK_SEARCH" "$HOOK_READ" > "$SETTINGS"
        echo "[OK] Created $SETTINGS"
    else
        CHANGED=false
        if ! jq -e '.hooks.PreToolUse[]? | select(.hooks[]?.command | test("hook-guard search"))' "$SETTINGS" &>/dev/null; then
            jq --argjson e "$HOOK_SEARCH" '.hooks.PreToolUse += [$e]' "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"
            CHANGED=true
        fi
        if ! jq -e '.hooks.PreToolUse[]? | select(.hooks[]?.command | test("hook-guard read"))' "$SETTINGS" &>/dev/null; then
            jq --argjson e "$HOOK_READ" '.hooks.PreToolUse += [$e]' "$SETTINGS" > "$SETTINGS.tmp" && mv "$SETTINGS.tmp" "$SETTINGS"
            CHANGED=true
        fi
        if $CHANGED; then
            echo "[MERGED] $SETTINGS — added missing hook-guard entries"
        else
            echo "[SKIP] $SETTINGS already has hook-guard entries"
        fi
    fi
fi
