#!/usr/bin/env bash
# verify-standards-gate.sh — AGENTS gate, README load steps, standards-* rules裁剪.
# shellcheck shell=bash

verify_standards_gate() {
  local target="$1"
  local readme="$target/docs/standards/README.md"
  local agents="$target/AGENTS.md"

  if [[ ! -f "$readme" ]]; then
    verify_info "standards README absent — skip standards-gate"
    return 0
  fi

  if [[ ! -f "$agents" ]] || ! grep -qE '编码前必读|门闩' "$agents"; then
    verify_error "AGENTS.md missing 编码前必读（门闩） while docs/standards/README.md exists"
  fi

  if ! grep -qE 'Agent 加载步骤（强制）|加载步骤（强制）' "$readme"; then
    verify_error "docs/standards/README.md missing Agent 加载步骤（强制）"
  fi

  if ! grep -qE '加载预算|按节' "$readme"; then
    verify_error "docs/standards/README.md missing 加载预算 or 按节 loading guidance"
  fi

  if grep -qE 'Read 本任务相关端的规范文件|加载这些文件（不要凭记忆|必须 Read：`docs/standards/[^`]+\.md`' "$readme" \
    && ! grep -qE '按节|相关章节|加载预算' "$readme"; then
    verify_error "docs/standards/README.md still implies whole-file load without section budget"
  fi

  local has_fe=0 has_api=0 has_be=0 has_sec=0 in=0 line fname
  while IFS= read -r line || [[ -n "$line" ]]; do
    if [[ "$line" == "## 当前项目选用"* ]]; then
      in=1
      continue
    fi
    if [[ "$in" -eq 1 && "$line" == "## "* ]]; then
      break
    fi
    [[ "$in" -eq 1 ]] || continue
    while IFS= read -r fname; do
      [[ -n "$fname" ]] || continue
      fname="${fname##*/}"
      case "$fname" in
        frontend-*) has_fe=1 ;;
        api-*) has_api=1 ;;
        backend-*) has_be=1 ;;
        security-*) has_sec=1 ;;
      esac
    done < <(printf '%s\n' "$line" | sed -nE 's/.*\]\(([^)]+\.md)\).*/\1/p; s/.*`([^`]+\.md)`.*/\1/p')
  done < "$readme"

  local roots=()
  local p
  for p in "$target/.agents/rules" "$target/.cursor/rules" "$target/.codebuddy/rules" "$target/.claude/rules"; do
    [[ -d "$p" ]] && roots+=("$p")
  done

  local f base lines
  shopt -s nullglob
  for p in "${roots[@]+"${roots[@]}"}"; do
    for f in "$p"/standards-*.mdc "$p"/standards-*.md; do
      base="$(basename "$f")"
      base="${base%.mdc}"
      base="${base%.md}"
      case "$base" in
        standards-frontend)
          [[ "$has_fe" -eq 0 ]] && verify_error "orphan rule $f but frontend not in 当前项目选用"
          ;;
        standards-api)
          [[ "$has_api" -eq 0 ]] && verify_error "orphan rule $f but api not in 当前项目选用"
          ;;
        standards-backend)
          [[ "$has_be" -eq 0 ]] && verify_error "orphan rule $f but backend not in 当前项目选用"
          ;;
        standards-security)
          [[ "$has_sec" -eq 0 ]] && verify_error "orphan rule $f but security not in 当前项目选用"
          ;;
      esac
      if grep -q 'alwaysApply:[[:space:]]*true' "$f" 2>/dev/null; then
        lines="$(wc -l < "$f" | tr -d ' ')"
        if [[ "$lines" -gt 80 ]]; then
          echo "WARN: alwaysApply rule too long ($lines lines): $f" >&2
        fi
      fi
    done
  done
  shopt -u nullglob

  local has_layout=0
  if [[ -d "$target/.agents" || -d "$target/.cursor" || -d "$target/.codebuddy" || -d "$target/.claude" ]]; then
    has_layout=1
  fi
  if [[ "$has_fe" -eq 1 && "$has_layout" -eq 1 ]]; then
    local found=0
    for p in "${roots[@]+"${roots[@]}"}"; do
      if [[ -f "$p/standards-frontend.mdc" || -f "$p/standards-frontend.md" ]]; then
        found=1
        break
      fi
    done
    if [[ "$found" -eq 0 ]]; then
      verify_error "frontend selected but standards-frontend rule missing under .agents/.cursor/.codebuddy/.claude rules"
    fi
  fi
}
