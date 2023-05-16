#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(
  cd "${SCRIPT_ROOT}"
  go mod vendor
  ls -d -1 ./vendor/k8s.io/code-generator
)}

bash "${CODEGEN_PKG}/generate-groups.sh" all \
  github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/generated \
  github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis \
  "mcs:v1alpha1" \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"


