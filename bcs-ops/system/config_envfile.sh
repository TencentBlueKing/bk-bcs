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

# config env file, default value can be overridden with ${CLUSTER_ENV}
# ToDo: atomization of configuration modification

VERSION=0.1.0
PROGRAM="$(basename "$0")"

SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."
BCS_ENV_FILE="${ROOT_DIR}/env/bcs.env"

readonly VERSION PROGRAM SELF_DIR ROOT_DIR BCS_ENV_FILE

usage_and_exit() {
  cat <<EOF
Usage:
    $PROGRAM
      [ -h --help -?,  show usage ]
      [ -v -V --version, show script version]
      [ -c --config
        init: READ CLUSTER_ENV and init bcs.env file
        dump: print bcs.env file.
        clean: clean bcs.env file ]
EOF
  exit "$1"
}

version() {
  echo "$PROGRAM version $VERSION"
}

# default env value in function `init_env`
init_env() {
  trap "utils::on_ERR;" ERR
  # host
  BK_HOME=${BK_HOME:-"/data/bcs"}
  K8S_IPv6_STATUS=${K8S_IPv6_STATUS:-"Disable"}
  LAN_IP=${LAN_IP:-}
  LAN_IPv6=${LAN_IPv6:-}
  BCS_SYSCTL=${BCS_SYSCTL:=1}
  if [[ -z ${LAN_IP} ]] && [[ ${K8S_IPv6_STATUS,,} != "singlestack" ]]; then
    LAN_IP="$("${ROOT_DIR}"/system/get_lan_ip -4)"
  fi
  if [[ -z $LAN_IPv6 ]] && [[ ${K8S_IPv6_STATUS,,} != "disable" ]]; then
    LAN_IPv6="$("${ROOT_DIR}"/system/get_lan_ip -6)"
  fi
  BCS_OFFLINE=${BCS_OFFLINE:-}
  INSTALL_METHOD=${INSTALL_METHOD:-"yum"}

  # cri
  CRI_TYPE=${CRI_TYPE:-"docker"}
  ## insregistry
  INSECURE_REGISTRY=${INSECURE_REGISTRY:-""}
  ## DOCKER
  DOCKER_VER=${DOCKER_VER:-"19.03.9"}
  DOCKER_LIB=${DOCKER_LIB:-"${BK_HOME}/lib/docker"}
  DOCKER_LIVE_RESTORE=${DOCKER_LIVE_RESTORE:-false}
  DOCKER_BRIDGE=${DOCKER_BRIDGE:-}
  ## CONTAINERD
  CONTAINERD_VER=${CONTAINERD_VER:-"1.6.21"}
  CONTAINERD_LIB=${CONTAINERD_LIB:-"${BK_HOME}/lib/containerd"}

  # k8s
  ## LIB
  ETCD_LIB=${ETCD_LIB:-"${BK_HOME}/lib/etcd"}
  KUBELET_LIB=${KUBELET_LIB:-"${BK_HOME}/lib/kubelet"}
  ## K8S_VER
  K8S_VER=${K8S_VER:-"1.20.15"}
  ## K8S_CIDR
  K8S_CTRL_IP=${K8S_CTRL_IP:-"$LAN_IP"}
  K8S_SVC_CIDR=${K8S_SVC_CIDR:-"10.96.0.0/12"}
  K8S_POD_CIDR=${K8S_POD_CIDR:-"10.244.0.0/16"}
  K8S_SVC_CIDRv6=${K8S_SVC_CIDRv6:-"fd00::1234:5678:1:0/112"}
  K8S_POD_CIDRv6=${K8S_POD_CIDRv6:-"fd00::1234:5678:100:0/104"}
  K8S_MASK=${K8S_MASK:-"24"}
  K8S_IPv6_MASK=${K8S_IPv6_MASK:-120}
  ## K8S_CNI
  K8S_CNI=${K8S_CNI:-"flannel"}
  ## K8S_extra
  local kubectl_extra_args
  kubectl_extra_args="allowed-unsafe-sysctls: 'net.ipv4.tcp_tw_reuse'"
  K8S_EXTRA_ARGS=${K8S_EXTRA_ARGS:-${kubectl_extra_args}}
  ## if BCS_CP_WORKER==1, means single master cluster, then untaint master
  BCS_CP_WORKER=${BCS_CP_WORKER:-0}

  # csi
  K8S_CSI=${K8S_CSI:-""}
  ## localpv
  LOCALPV_DIR=${LOCALPV_DIR:-${BK_HOME}/localpv}
  LOCALPV_DST_DIR=${LOCALPV_DST_DIR:-"/mnt/blueking"}
  LOCALPV_COUNT=${LOCALPV_COUNT:-20}
  LOCALPV_reclaimPolicy=${LOCALPV_reclaimPolicy:-"Delete"}

  # mirror
  ## yum_mirror
  MIRROR_URL=${MIRROR_URL:-"https://mirrors.tencent.com"}
  ## repo_url
  REPO_URL=${REPO_URL:-"https://bkopen-1252002024.file.myqcloud.com/ce7/tools"}
  ##
  MIRROR_IP=${MIRROR_IP:-}
  ## image_registry
  ### docker.io
  REPO_MIRRORS=${REPO_MIRRORS:-"https://mirror.ccs.tencentyun.com"}
  ### other image
  BK_PUBLIC_REPO=${BK_PUBLIC_REPO:-"hub.bktencent.com"}

  # helm
  BKREPO_URL=${BKREPO_URL:-"https://hub.bktencent.com/chartrepo/blueking"}

  # apiserver HA
  ENABLE_APISERVER_HA=${ENABLE_APISERVER_HA:-"false"}
  APISERVER_HA_MODE=${APISERVER_HA_MODE:-"bcs-apiserver-proxy"}
  VIP=${VIP:-}
  APISERVER_HOST=${APISERVER_HOST:-}
  ## bcs apiserver proxy
  APISERVER_PROXY_VERSION=${APISERVER_PROXY_VERSION:-"v1.29.0-alpha.130-tencent"}
  PROXY_TOOL_PATH=${PROXY_TOOL_PATH:-"/usr/bin"}
  VS_PORT=${VS_PORT:-"6443"}
  LVS_SCHEDULER=${LVS_SCHEDULER:-"rr"}
  PERSIST_DIR=${PERSIST_DIR:-"/root/.bcs"}
  MANAGER_INTERVAL=${MANAGER_INTERVAL:-"10"}
  LOG_LEVEL=${LOG_LEVEL:-"3"}
  DEBUG_MODE=${DEBUG_MODE:-"true"}
  ## kube-vip
  KUBE_VIP_VERSION=${KUBE_VIP_VERSION:-"v0.5.12"}
  BIND_INTERFACE=${BIND_INTERFACE:-}
  VIP_CIDR=${VIP_CIDR:-"32"}
  ## multus
  ENABLE_MULTUS_HA=${ENABLE_MULTUS_HA:-"true"}
}

source_cluster_env() {
  if [[ -n ${CLUSTER_ENV:-} ]]; then
    local cluster_env
    cluster_env=$(base64 -d <<<"${CLUSTER_ENV}")
    utils::log "DEBUG" "cluster_env: ${cluster_env}"
    # shellcheck source=/dev/null
    source <(echo "${cluster_env}")
  fi
}

_setIPUsage_and_exit() {
  cat <<EOF
you can set LAN_IP manually by following:
set -x
LAN_IP=<YOUR LAN IP>
LAN_IPv6<YOUR LAN ipv6> #if enable K8S_IPv6_STATUS=dualstack
set -x
EOF
  exit 1
}

check_env() {
  trap "utils::on_ERR;" ERR
  # match k8s_ver
  if ! [[ $K8S_VER =~ ^1\.2[0-4] ]]; then
    utils::log "ERROR" \
      "Only support K8S_VER 1.2[0-4].x, here is :${K8S_VER}"
  fi

  # match cri and k8s_ver
  if [[ $CRI_TYPE == "docker" && ${K8S_VER} =~ ^1\.24 ]]; then
    utils::log "ERROR" "K8S_VER:${K8S_VER} unsupport CRI:${CRI_TYPE}"
  fi

  # match CRI_TYPE
  case $CRI_TYPE in
    "docker" | "containerd") ;;
    *)
      utils::log "ERROR" "unsupport CRI_TYPE:${CRI_TYPE}"
      ;;
  esac

  # match ipv6_status
  # ToDo: ip format check
  case ${K8S_IPv6_STATUS,,} in
    "disable")
      if [[ -z $LAN_IP ]]; then
        utils::log "WARN" "missing LAN_IP"
        _setIPUsage_and_exit
      fi
      ;;
    "singlestack")
      if [[ -z $LAN_IPv6 ]]; then
        utils::log "WARN" "missing LAN_IPv6"
        _setIPUsage_and_exit
      fi
      if [[ $K8S_VER =~ ^1\.2[0-2] ]]; then
        utils::log "ERROR" \
          "ipv6 DualStack only support 1.2[3-4].x, here is ${K8S_VER}"
      fi
      LAN_IP=$LAN_IPv6
      K8S_SVC_CIDR=${K8S_SVC_CIDRv6}
      K8S_POD_CIDR=${K8S_POD_CIDRv6}
      ;;
    "dualstack")
      if [[ -z $LAN_IP ]] || [[ -z $LAN_IPv6 ]]; then
        utils::log "WARN" "missing LAN_IP or LAN_IPv6"
        _setIPUsage_and_exit
      fi
      if [[ $K8S_VER =~ ^1\.2[0-2] ]]; then
        utils::log "ERROR" \
          "ipv6 DualStack only support 1.2[3-4].x, here is ${K8S_VER}"
      fi
      ;;
    *)
      utils::log "ERROR" \
        "K8S_IPv6_STATUS only accept: Disable|SingleStack|DualStack, \
now is ${K8S_IPv6_STATUS}"
      ;;
  esac

  [[ -n $K8S_CTRL_IP ]] || K8S_CTRL_IP=$LAN_IP

  if [[ ${ENABLE_APISERVER_HA} == "external" ]];then
    if [[ -z "${VIP}" ]] && [[ -z "${APISERVER_HOST}" ]];then
      utils::log "ERROR" \
        "if ENABLE_APISERVER_HA is ${ENABLE_APISERVER_HA}, VIP or APISERVER_HOST must be set"
    fi
  fi
}

render_env() {
  utils::log "INFO" "RENDERING bcs env file"
  [[ -d "${ROOT_DIR}/env" ]] || install -dv "${ROOT_DIR}/env"
  cat >"${BCS_ENV_FILE}" <<EOF
# bcs config begin
## HOST
BK_HOME="${BK_HOME}"
LAN_IP="${LAN_IP}"
$(
    [[ ${K8S_IPv6_STATUS,,} == "dualstack" ]] \
      && echo LAN_IPv6=\""${LAN_IPv6}"\"
  )
BCS_SYSCTL=${BCS_SYSCTL:=1}
K8S_IPv6_STATUS="${K8S_IPv6_STATUS}"
BCS_OFFLINE="${BCS_OFFLINE}"
INSTALL_METHOD="${INSTALL_METHOD}"

## CRI
CRI_TYPE="${CRI_TYPE}"
INSECURE_REGISTRY=${INSECURE_REGISTRY}
$(
    case "${CRI_TYPE,,}" in
      "containerd")
        cat <<CRI_EOF
CONTAINERD_LIB="${CONTAINERD_LIB}"
CONTAINERD_VER="${CONTAINERD_VER}"
CRI_EOF
        ;;
      "docker")
        cat <<CRI_EOF
DOCKER_LIB="${DOCKER_LIB}"
DOCKER_VER="${DOCKER_VER}"
DOCKER_LIVE_RESTORE="${DOCKER_LIVE_RESTORE}"
DOCKER_BRIDGE="${DOCKER_BRIDGE}"
CRI_EOF
        ;;
    esac
  )

## K8S
ETCD_LIB="${ETCD_LIB}"
KUBELET_LIB="${KUBELET_LIB}"
K8S_VER="${K8S_VER}"
K8S_CTRL_IP="${K8S_CTRL_IP}"
K8S_SVC_CIDR="${K8S_SVC_CIDR}"
K8S_POD_CIDR="${K8S_POD_CIDR}"
$(
    [[ ${K8S_IPv6_STATUS,,} == "dualstack" ]] \
      && cat <<IPv6_EOF
K8S_SVC_CIDRv6="${K8S_SVC_CIDRv6}"
K8S_POD_CIDRv6="${K8S_POD_CIDRv6}"
IPv6_EOF
  )
$(
    [[ ${K8S_IPv6_STATUS,,} != "disable" ]] \
      && echo K8S_IPv6_MASK="${K8S_IPv6_MASK}"
  )
K8S_CNI="${K8S_CNI}"
K8S_EXTRA_ARGS="${K8S_EXTRA_ARGS}"
## if BCS_CP_WORKER==1, means single master cluster, then untaint master
BCS_CP_WORKER="${BCS_CP_WORKER}"

# csi
K8S_CSI="${K8S_CSI}"
$(
    case "${K8S_CSI,,}" in
      "localpv")
        cat <<CSI_EOF
LOCALPV_DIR="${LOCALPV_DIR}"
LOCALPV_DST_DIR="${LOCALPV_DST_DIR}"
LOCALPV_COUNT="${LOCALPV_COUNT}"
LOCALPV_reclaimPolicy="${LOCALPV_reclaimPolicy}"
CSI_EOF
        ;;
    esac
  )

## yum_mirror
MIRROR_URL="${MIRROR_URL}"
REPO_URL="${REPO_URL}"
MIRROR_IP="${MIRROR_IP}"
## image_registry
### docker.io
REPO_MIRRORS="${REPO_MIRRORS}"
### registry.k8s.io
BK_PUBLIC_REPO="${BK_PUBLIC_REPO}"

## helm
BKREPO_URL="${BKREPO_URL}"


# apiserver HA
ENABLE_APISERVER_HA="${ENABLE_APISERVER_HA}"
APISERVER_HA_MODE="${APISERVER_HA_MODE}"
VIP="${VIP}"
APISERVER_HOST="${APISERVER_HOST}"
## bcs apiserver proxy
APISERVER_PROXY_VERSION="${APISERVER_PROXY_VERSION}"
PROXY_TOOL_PATH="${PROXY_TOOL_PATH}"
VS_PORT="${VS_PORT}"
LVS_SCHEDULER="${LVS_SCHEDULER}"
PERSIST_DIR="${PERSIST_DIR}"
MANAGER_INTERVAL="${MANAGER_INTERVAL}"
LOG_LEVEL="${LOG_LEVEL}"
DEBUG_MODE="${DEBUG_MODE}"
## kube-vip
KUBE_VIP_VERSION="${KUBE_VIP_VERSION}"
BIND_INTERFACE="${BIND_INTERFACE}"
VIP_CIDR="${VIP_CIDR}"
## multus
ENABLE_MULTUS_HA="${ENABLE_MULTUS_HA}"
# bcs config end
EOF
}

config_init() {
  trap "utils::on_ERR;" ERR
  if [[ -f ${BCS_ENV_FILE} ]]; then
    #shellcheck source=/dev/null
    source "${BCS_ENV_FILE}"
  fi
  source_cluster_env
  init_env
  check_env
  render_env
  config_dump
}

config_dump() {
  if [[ -f ${BCS_ENV_FILE} ]]; then
    cat "${BCS_ENV_FILE}"
    return 0
  else
    utils::log "ERROR" "${BCS_ENV_FILE} not find!"
  fi
}

config_clean() {
  if [[ -f "${BCS_ENV_FILE}" ]]; then
    if grep -q "bcs config begin" "${BCS_ENV_FILE}"; then
      sed -ri.bcs-"$(date +%s)".bak "/bcs config begin/,/bcs config end/d" "${BCS_ENV_FILE}"
    fi
  fi
  utils::log "OK" "Clean ${BCS_ENV_FILE}"
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
}

main() {
  local source_files
  # root_dir="${SELF_DIR}/" # relative path to script
  source_files=("${ROOT_DIR}/functions/utils.sh")
  for file in "${source_files[@]}"; do
    safe_source "$file"
  done

  local op_type project

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
      --config | -c)
        op_type="config"
        shift
        case "$1" in
          "init")
            project="init"
            ;;
          "dump")
            project="dump"
            ;;
          "clean")
            project="clean"
            ;;
          *)
            utils::log "ERROR" "unkown $1"
            ;;
        esac
        source_cluster_env
        init_env
        ;;
      *)
        utils::log "ERROR" "unkown para: $1"
        ;;
    esac
    shift
  done
  utils::check_op "${op_type}" "${project}"
  "${op_type}_${project}"
}

main "$@"
