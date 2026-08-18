#!/usr/bin/env bash
# verify-desensitize.sh — F7: no personal identifiers / machine absolute paths in harness docs.
# shellcheck shell=bash

verify_desensitize() {
  local target="$1"
  local f

  while IFS= read -r f; do
    [[ -z "$f" ]] && continue
    # Scope: AGENTS.md, docs/harness/**, docs/glossary.md
    case "$f" in
      "$target/AGENTS.md"|"$target/docs/glossary.md") ;;
      "$target/docs/harness/"*) ;;
      *) continue ;;
    esac

    if grep -nE '/(home|Users|data/go|root|tmp)/' "$f" >/dev/null 2>&1; then
      # Allow mention inside code fences that document detection patterns? Still ERROR per plan.
      verify_error "absolute machine path in harness doc: $f"
    fi
    if grep -nE '[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}' "$f" >/dev/null 2>&1; then
      verify_error "email address in harness doc: $f"
    fi
  done < <(verify_list_md "$target")
}
