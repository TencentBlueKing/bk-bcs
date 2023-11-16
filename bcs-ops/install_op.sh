#!/bin/bash

#######################################
# Tencent is pleased to support the open source community by making Blueking Container Service available.
# Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
# Licensed under the MIT License (the "License"); you may not use this file except
# in compliance with the License. You may obtain a copy of the License at
# http://opensource.org/licenses/MIT
# Unless required by applicable law or agreed to in writing, software distributed under
# the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
# either express or implied. See the License for the specific language governing permissions and
# limitations under the License.
#######################################

# config O&M-related configuration
set -euo pipefail

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR=${SELF_DIR}
readonly SELF_DIR ROOT_DIR

safe_source() {
  local source_file=$1
  if [[ -f ${source_file} ]]; then
    #shellcheck source=/dev/null
    source "${source_file}"
  else
    echo "[ERROR]: FAIL to source, missing ${source_file}"
    exit 1
  fi
}

source_files=("${ROOT_DIR}/functions/utils.sh")
for file in "${source_files[@]}"; do
  safe_source "$file"
done

"${ROOT_DIR}"/k8s/operate_completion kubeadm kubectl helm ctr yq crictl

if [[ -n "${BKREPO_URL:-}" ]]; then
  if command -v helm &>/dev/null; then
    utils::log "INFO" "Add repo blueking: ${BKREPO_URL}"
    k8s::safe_add_helmrepo blueking "${BKREPO_URL}"
    utils::log "OK" "blueking community helm chart repo added"
  else
    warning "helm command not found, skipping"
    return 0
  fi
else
  utils::log "WARN" "BKREPO_URL is null, skipping"
fi
