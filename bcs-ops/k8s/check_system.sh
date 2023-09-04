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
# 定义检查的应用
systemapps=(etcd kube-controller-manager kube-scheduler kube-apiserver kube-proxy kube-dns)

# 获取当前集群上下文
context=$(kubectl config current-context)

# 输出检查信息
utils::log "INFO" "检查K8s集群 $context 的系统应用状态:"

# 循环检查每个系统应用
for app in "${systemapps[@]}"; do
  # 获取应用的Pod名称

  case "${app}" in
    etcd|kube-controller-manager|kube-scheduler|kube-apiserver)
        pod_name=$(kubectl get pods --all-namespaces -o jsonpath='{.items[?(@.metadata.labels.component=="'"$app"'")].metadata.name}')
        # 获取应用的命名空间
        namespace=$(kubectl get pods --all-namespaces -o jsonpath='{.items[?(@.metadata.labels.component=="'"$app"'")].metadata.namespace}' | awk '{print $1}')
    ;;
    *)
        pod_name=$(kubectl get pods --all-namespaces -o jsonpath='{.items[?(@.metadata.labels.k8s-app=="'"$app"'")].metadata.name}')
        # 获取应用的命名空间
        namespace=$(kubectl get pods --all-namespaces -o jsonpath='{.items[?(@.metadata.labels.k8s-app=="'"$app"'")].metadata.namespace}' | awk '{print $1}')
    ;;
  esac

  # 检查应用是否存在
  if [ -z "$pod_name" ]; then
    utils::log "ERROR" "系统应用 $app 不存在"
    continue
  fi

  for pod in $pod_name;do
# 获取应用的状态
    status=$(kubectl get pods -n "$namespace" "$pod" -o jsonpath='{.status.phase}')
    hostIP=$(kubectl get pods -n "$namespace" "$pod" -o jsonpath='{.status.hostIP}')

    # 输出应用的状态
    case $status in
        Running)
        utils::log "INFO" "系统应用 $app 正在运行,主机IP: $hostIP"
        ;;
        Pending)
        utils::log "INFO" "系统应用 $app 处于等待状态,主机IP: $hostIP"
        ;;
        Succeeded)
        utils::log "OK" "系统应用 $app 已成功完成,主机IP: $hostIP"
        ;;
        Failed)
        utils::log "FATAL" "系统应用 $app 失败,主机IP: $hostIP"
        ;;
        Unknown)
        utils::log "ERROR" "系统应用 $app 状态未知,主机IP: $hostIP"
        ;;
        *)
        utils::log "ERROR" "系统应用 $app 状态未知,主机IP: $hostIP"
        ;;
    esac
  done
done
