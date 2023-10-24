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

# Process encapsulation of tool installation
VERSION="1.0.0"
PROGRAM="$(basename "$0")"
SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."

PROJECTS=(yq jq)

readonly SELF_DIR ROOT_DIR PROJECTS

usage_and_exit() {
  cat <<EOF
Usage:
    $PROGRAM
      [ -h --help -?  show usage ]
      [ -v -V --version show script version]
      [ ${PROJECTS[*]} ]
EOF
  exit "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}

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

main() {
  local source_files
  source_files=("${ROOT_DIR}/functions/utils.sh")
  for file in "${source_files[@]}"; do
    safe_source "$file"
  done

  local project

  (($# == 0)) && usage_and_exit 1
  while (($# > 0)); do
    case "$1" in
      --help | -h | '-?')
        usage_and_exit 0
        ;;
      --version | -v | -V)
        version
        exit 0
        ;;
      -*)
        # ToDo: Unified standard error code
        export ERR_CODE=1
        utils::log "ERROR" "unkown para: $1"
        ;;
      *)
        project="${ROOT_DIR}/tools/install_$1"
        if [[ -x "${project}" ]]; then
          "${project}"
        else
          utils::log "ERROR" "can't exec ${project}"
        fi
        ;;
    esac
    shift
  done
  return 0
}

main "$@"
