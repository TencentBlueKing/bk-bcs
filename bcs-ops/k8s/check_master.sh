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

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."
CHECK_NAMESAPCE=kube-system
RULE='kube-controller-manager|kube-scheduler|kube-apiserver'

readonly SELF_DIR ROOT_DIR

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
        echo "[ERROR]: FAIL to source, missing ${source_file}" >&2
        exit 1
    fi
    return 0
}

safe_source "${ROOT_DIR}/functions/utils.sh"
RESULT_PATH="/tmp/check_master"
# 创建文件夹
[[ -d $RESULT_PATH ]] || mkdir -p $RESULT_PATH

# systemapps=(kube-controller-manager kube-scheduler kube-apiserver)

for pod in $(kubectl get po -n ${CHECK_NAMESAPCE} | grep -E ${RULE} | awk '{print $1}'); do
    kubectl get po "$pod" -n ${CHECK_NAMESAPCE} -o json | jq .spec.containers[0].command >"$RESULT_PATH/$pod"
    host_ip=$(kubectl get po "$pod" -n ${CHECK_NAMESAPCE} -o json | jq .status.hostIP | sed 's/"//g')
    sed -i "s/${host_ip}//" "$RESULT_PATH/$pod"
done

IFS='|' read -ra RULES <<<"$RULE"

for RULE in "${RULES[@]}"; do
    find $RESULT_PATH/"$RULE"* | while read -r file_name; do
        jq -c . "$file_name" >>/tmp/temp.txt."$RULE"
        sort /tmp/temp.txt."$RULE" >/tmp/sorted.txt."$RULE"
        if [ "$(uniq -c /tmp/sorted.txt."$RULE" | wc -l)" -eq 1 ]; then
            utils::log "OK" "All $RULE have the same component configuration"
        else
            pod_name=$(echo "$file_name" | awk -F'/' '{print $4}')
            utils::log "WARN" "have difference to $RULE，Please check $pod_name"
            utils::log "WARN" "You can use commands == kubectl describe po ${pod_name} -n ${CHECK_NAMESAPCE}| grep -A 50 Command == to check"
            cat "$file_name"
        fi
    done
    rm /tmp/temp.txt."$RULE" /tmp/sorted.txt."$RULE"
done
rm "${RESULT_PATH}"/* >/dev/null 2>&1
