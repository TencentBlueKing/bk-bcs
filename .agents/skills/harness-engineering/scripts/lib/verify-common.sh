#!/usr/bin/env bash
# Shared helpers for harness-verify subchecks.
# shellcheck shell=bash

VERIFY_ERROR_COUNT=0

verify_error() {
  local msg="$1"
  echo "ERROR: $msg" >&2
  VERIFY_ERROR_COUNT=$((VERIFY_ERROR_COUNT + 1))
}

verify_info() {
  echo "INFO: $1"
}

# List harness deliverable markdown (not the whole monorepo / skill fixtures).
verify_list_md() {
  local target="$1"
  local p
  for p in \
    "$target/AGENTS.md" \
    "$target/docs/glossary.md"
  do
    [[ -f "$p" ]] && printf '%s\n' "$p"
  done
  for p in harness standards business-standards dev-map; do
    if [[ -d "$target/docs/$p" ]]; then
      find "$target/docs/$p" -type f -name '*.md' -print 2>/dev/null
    fi
  done
}
