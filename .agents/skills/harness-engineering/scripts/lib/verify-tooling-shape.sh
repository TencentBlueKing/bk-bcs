#!/usr/bin/env bash
# verify-tooling-shape.sh — F6: tooling.md must not store runtime readiness columns/cells.
# shellcheck shell=bash

verify_tooling_shape() {
  local target="$1"
  local tooling="$target/docs/harness/tooling.md"

  if [[ ! -f "$tooling" ]]; then
    verify_info "tooling.md absent — skip shape check"
    return 0
  fi

  if grep -nE '环境状态' "$tooling" >/dev/null 2>&1; then
    verify_error "tooling.md must not contain column/header 环境状态 ($tooling)"
  fi

  # Status cells that belong in harness-doctor stdout, not the repo.
  if grep -nE '✅[[:space:]]*已就绪|❌[[:space:]]*未安装|未接入' "$tooling" >/dev/null 2>&1; then
    verify_error "tooling.md must not store runtime status cells (已就绪/未安装/未接入) ($tooling)"
  fi
}
