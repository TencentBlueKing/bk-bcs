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

set -euo pipefail
trap "utils::on_ERR;" ERR
# install k8s node

VERSION=0.1.0
PROGRAM="$(basename "$0")"

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="$SELF_DIR"

PROJECT=(init join)
readonly VERSION PROGRAM SELF_DIR ROOT_DIR

usage_and_exit() {
  cat <<EOF
Usage:
    $PROGRAM
      [ -h --help -?  show usage ]
      [ -v -V --version show script version]
      [ -i --install, supprt: ${PROJECT[*]}]
EOF
  return "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}

#######################################
# check file and source
# Arguments:
# $1: source_file
# Return:
# if file exists, source return 0; else exit 1
#######################################
safe_source() {
  local source_file=$1
  if [[ -f ${source_file} ]]; then
    #shellcheck source=/dev/null
    source "${source_file}"
  else
    echo "[ERROR]: FAIL to source, missing ${source_file}"
    exit 1
  fi
  return 0
}

source_files=("${ROOT_DIR}/functions/utils.sh")
for file in "${source_files[@]}"; do
  safe_source "$file"
done
"${ROOT_DIR}"/system/config_envfile.sh -c init
"${ROOT_DIR}"/system/config_system.sh -c dns sysctl
"${ROOT_DIR}"/k8s/install_cri.sh
"${ROOT_DIR}"/k8s/install_k8s_tools
"${ROOT_DIR}"/k8s/render_kubeadm

# pull image
if [[ -n ${BCS_OFFLINE:-} ]]; then
  # import local image
  true
fi
kubeadm --config="${ROOT_DIR}/kubeadm-config" config images pull \
  || utils::log "FATAL" "fail to pull k8s image"

kubeadm join --config="${ROOT_DIR}/kubeadm-config" -v 11
