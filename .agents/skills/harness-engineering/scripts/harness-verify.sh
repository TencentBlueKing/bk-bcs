#!/usr/bin/env bash
# harness-verify.sh — machine gate for harness-engineering outputs.
# Usage: harness-verify.sh <target-root>
# Exit 0 on success; non-zero on ERROR (when HARNESS_VERIFY_STRICT=1, default).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/verify-common.sh
source "$SCRIPT_DIR/lib/verify-common.sh"
# shellcheck source=lib/verify-facts.sh
source "$SCRIPT_DIR/lib/verify-facts.sh"
# shellcheck source=lib/verify-skeleton.sh
source "$SCRIPT_DIR/lib/verify-skeleton.sh"
# shellcheck source=lib/verify-consistency.sh
source "$SCRIPT_DIR/lib/verify-consistency.sh"
# shellcheck source=lib/verify-no-workflow.sh
source "$SCRIPT_DIR/lib/verify-no-workflow.sh"
# shellcheck source=lib/verify-tooling-shape.sh
source "$SCRIPT_DIR/lib/verify-tooling-shape.sh"
# shellcheck source=lib/verify-desensitize.sh
source "$SCRIPT_DIR/lib/verify-desensitize.sh"
# shellcheck source=lib/verify-standards-gate.sh
source "$SCRIPT_DIR/lib/verify-standards-gate.sh"

TARGET="${1:?usage: harness-verify.sh <target-root>}"
TARGET="$(cd "$TARGET" && pwd)"
HARNESS_VERIFY_STRICT="${HARNESS_VERIFY_STRICT:-1}"

verify_info "harness-verify target=$TARGET"

verify_facts "$TARGET"
verify_skeleton "$TARGET"
verify_consistency "$TARGET"
verify_no_workflow "$TARGET"
verify_tooling_shape "$TARGET"
verify_desensitize "$TARGET"
verify_standards_gate "$TARGET"

if [[ "$VERIFY_ERROR_COUNT" -gt 0 ]]; then
  echo "harness-verify: $VERIFY_ERROR_COUNT error(s)" >&2
  if [[ "$HARNESS_VERIFY_STRICT" == "1" ]]; then
    exit 1
  fi
  exit 0
fi

echo "harness-verify: OK"
exit 0
