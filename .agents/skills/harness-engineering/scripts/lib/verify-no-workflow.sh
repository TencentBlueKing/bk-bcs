#!/usr/bin/env bash
# F3: harness must not leave *forced* workflow-agent integration in AGENTS or harness docs.
# Deprecation / historical mentions (same line contains 废弃|DEPRECATED) are allowed.
# shellcheck shell=bash

_verify_workflow_force_line() {
  local f="$1"
  local line="$2"
  if [[ "$line" == *不允许跳过工作流* ]]; then
    verify_error "$f contains 不允许跳过工作流 (harness no longer integrates workflow)"
    return
  fi
  if [[ "$line" == *workflow-agent* ]] \
    && [[ "$line" != *已废弃* ]] \
    && [[ "$line" != *废弃* ]] \
    && [[ "$line" != *DEPRECATED* ]]; then
    verify_error "$f contains live workflow-agent reference (mark 已废弃 or remove; harness no longer integrates workflow)"
  fi
}

verify_no_workflow() {
  local target="$1"
  local f line

  for f in "$target/AGENTS.md"; do
    [[ -f "$f" ]] || continue
    while IFS= read -r line || [[ -n "$line" ]]; do
      _verify_workflow_force_line "$f" "$line"
    done <"$f"
    if grep -qE '^##[[:space:]]*开发工作流' "$f"; then
      verify_error "$f contains ## 开发工作流 section (harness no longer integrates workflow)"
    fi
  done

  if [[ -d "$target/docs/harness" ]]; then
    while IFS= read -r f; do
      [[ -z "$f" ]] && continue
      # Skip gardening reports that may quote historical errors
      case "$f" in
        */gardening-report.md|*/gardening-proposals.md|*/generating-report.md) continue ;;
      esac
      while IFS= read -r line || [[ -n "$line" ]]; do
        _verify_workflow_force_line "$f" "$line"
      done <"$f"
    done < <(find "$target/docs/harness" -type f -name '*.md' 2>/dev/null)
  fi
}
