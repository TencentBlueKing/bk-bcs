#!/usr/bin/env bash
# F5: cross-doc contradictions (vitest force vs "no unit tests").
# shellcheck shell=bash

verify_consistency() {
  local target="$1"
  local forces_vitest=0
  local forbids_claim=0
  local f

  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    if grep -qE 'npx[[:space:]]+vitest|vitest[[:space:]]+run|强制.*vitest' "$f" \
      && grep -qE '必须全部通过|每次提交前必须' "$f"; then
      forces_vitest=1
    fi
    # also treat bare mandatory vitest lines in standards
    if grep -qE 'npx[[:space:]]+vitest[[:space:]]+run' "$f"; then
      forces_vitest=1
    fi
    if grep -qE '未定义单元测试|不得声称.*测试已通过' "$f"; then
      forbids_claim=1
    fi
  done < <(verify_list_md "$target")

  if [[ $forces_vitest -eq 1 && $forbids_claim -eq 1 ]]; then
    verify_error "cross-doc conflict: vitest/mandatory test commanded while another doc forbids claiming tests passed"
  fi
}
