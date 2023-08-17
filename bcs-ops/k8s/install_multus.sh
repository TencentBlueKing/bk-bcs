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

VERSION="1.0.0"
PROGRAM="$(basename "$0")"

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="$SELF_DIR/.."

readonly VERSION PROGRAM SELF_DIR ROOT_DIR

usage_and_exit() {
  local PROGRAM
  cat <<EOF
Usage:
    $PROGRAM [ -h --help -?  show usage ]
             [ -v -V --version show script version]
EOF
  return "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}


main() {
  local source_files
  source_files=("${ROOT_DIR}/functions/utils.sh" "${ROOT_DIR}/env/bcs.env")
  for file in "${source_files[@]}"; do
    safe_source "$file"
  done

  "${ROOT_DIR}"/k8s/operate_multus apply
}