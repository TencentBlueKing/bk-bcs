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
# install k8s master flow script
# two roles: init master and join master

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="$SELF_DIR"

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
safe_source "${ROOT_DIR}/functions/k8s.sh"

"${ROOT_DIR}"/system/config_envfile.sh -c init
"${ROOT_DIR}"/system/config_system.sh -c dns sysctl
"${ROOT_DIR}"/k8s/install_cri.sh
"${ROOT_DIR}"/k8s/install_k8s_tools
"${ROOT_DIR}"/k8s/render_kubeadm

safe_source "${ROOT_DIR}/env/bcs.env"

# ToDo: import image: cni\metric
if [[ -n ${BCS_OFFLINE:-} ]]; then
  true
fi
# pull image
kubeadm --config="${ROOT_DIR}/kubeadm-config" config images pull \
  || utils::log "FATAL" "fail to pull k8s image"

if [[ -z ${MASTER_JOIN_CMD:-} ]]; then
  kubeadm init --config="${ROOT_DIR}/kubeadm-config" -v 11
  install -dv "$HOME/.kube"
  install -v -m 600 -o "$(id -u)" -g "$(id -g)" \
    /etc/kubernetes/admin.conf "$HOME/.kube/config"
  "${ROOT_DIR}"/k8s/install_cni.sh
  "${ROOT_DIR}"/k8s/operate_metrics_server apply
  "${ROOT_DIR}"/k8s/install_helm
  "${ROOT_DIR}"/k8s/render_k8s_joincmd
  if [[ ${ENABLE_APISERVER_HA} == "true" ]]; then
    [[ -z ${VIP} ]] && utils::log "ERROR" "apiserver HA is enabled but VIP is not set"
    if [[ ${APISERVER_HA_MODE} == "kube-vip" ]]; then
      "${ROOT_DIR}"/k8s/operate_kube_vip apply
    elif [[ ${APISERVER_HA_MODE} == "bcs-apiserver-proxy" ]]; then
      "${ROOT_DIR}"/k8s/operate_bap apply
    fi
  fi
else
  kubeadm join --config="${ROOT_DIR}/kubeadm-config" -v 11 \
    || utils::log "FATAL" "${LAN_IP} failed to join master: ${K8S_CTRL_IP}"
  if [[ ${ENABLE_APISERVER_HA} == "true" ]]; then
    [[ -z ${VIP} ]] && utils::log "ERROR" "apiserver HA is enabled but VIP is not set"
    "${ROOT_DIR}"/system/config_bcs_dns -u "${LAN_IP}" k8s-api.bcs.local
    k8s::restart_kubelet
    if [[ ${APISERVER_HA_MODE} == "kube-vip" ]]; then
      "${ROOT_DIR}"/k8s/operate_kube_vip apply
    fi
  fi
fi
