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
    echo "[ERROR]: FAIL to source, missing ${source_file}"
    exit 1
  fi
  return 0
}

safe_source "${ROOT_DIR}/functions/utils.sh"
RESULT_PATH="/tmp/check_master"
# 创建文件夹
if [ ! -d $RESULT_PATH ]
then
  mkdir -p $RESULT_PATH
fi

rm  "${RESULT_PATH:?}/"*

K8S_API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
K8S_TOKEN=$(kubectl -n kube-system describe "$(kubectl -n kube-system get secret -n kube-system -o name | grep namespace-controller-)" | grep token:|awk -F ':' '{print $2}')
MANAGE_POD_NAME=$(curl -s "$K8S_API_SERVER"/api/v1/namespaces/kube-system/pods/ -k|jq .items[].metadata.name| grep kube-controller-manager| sed 's/"//g')

for pod in $MANAGE_POD_NAME;do
  curl -H "Authorization: Bearer $K8S_TOKEN" "$K8S_API_SERVER/api/v1/namespaces/kube-system/pods/$pod" -k -s|jq .spec.containers[0].command > "$RESULT_PATH/$pod"
done

find $RESULT_PATH/* | while read -r file; do
    jq -c . "$file" >> /tmp/temp.txt
done

sort /tmp/temp.txt > /tmp/sorted.txt

if [ "$(uniq -c /tmp/sorted.txt | wc -l)" -eq 1 ]; then
    utils::log "OK" "所有 kube-controller-manager-master 配置内容相同"
else
    for file in "$RESULT_PATH"/*; do
    utils::log "ERROR" "有部分 kube-controller-manager-master 配置不相同，请检查"
    utils::log "ERROR" "echo $file | awk -F'/' '{print $4}'"
    cat "$file"
done
fi

rm /tmp/temp.txt /tmp/sorted.txt
