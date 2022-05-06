#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

PROJECT_ROOT=$(cd $(dirname ${BASH_SOURCE})/..; pwd)
GOPATH=${GOPATH:-$(go env GOPATH)}
# corresponding to go mod init <module>
MODULE=github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager
SUBPATH=${PROJECT_ROOT:${#GOPATH}+5}
# api package
APIS_PKG=pkg/apis
# generated output package
OUTPUT_PKG=pkg/client
# group-version such as foo:v1alpha1
GROUP_VERSION=tkex:v1alpha1

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
echo "script root: ${SCRIPT_ROOT}"
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
echo "codegen package: ${CODEGEN_PKG}"

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
#bash "${CODEGEN_PKG}"/generate-groups.sh \
#  "client,lister,informer,deepcopy" \
#  ${SUBPATH}/${OUTPUT_PKG} \
#  ${SUBPATH}/${APIS_PKG} \
#  ${GROUP_VERSION} \
#  --output-base "../../.." \
#  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

bash "${CODEGEN_PKG}"/generate-groups.sh \
  "client,lister,informer,deepcopy" \
  ${MODULE}/${OUTPUT_PKG} \
  ${MODULE}/${APIS_PKG} \
  ${GROUP_VERSION} \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt \
  --output-base generated

cp -r generated/${MODULE}/pkg/client pkg
cp -r generated/${MODULE}/pkg/apis pkg
rm -rf generated

echo "generate clientset,lister,informer,deepcopy done"