#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

bash "${SCRIPT_ROOT}/hack/verify-codegen.sh"
bash "${SCRIPT_ROOT}/hack/verify-crdgen.sh"
bash "${SCRIPT_ROOT}/hack/verify-gofmt.sh"

echo "verifying if there is any unused dependency in go module"
make -C "${SCRIPT_ROOT}" tidy
STATUS=$(cd "${SCRIPT_ROOT}" && git status --porcelain go.mod go.sum)
if [ ! -z "$STATUS" ]; then
  echo "Running 'go mod tidy' to fix your 'go.mod' and/or 'go.sum'"
  exit 1
fi
echo "go module is tidy."
