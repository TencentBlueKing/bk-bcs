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
  exit "$1"
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

init_bap_rule() {
  if [[ -z ${BK_PUBLIC_REPO} ]]; then
    utils::log "ERROR" "init bcs-apiserver-proxy failed, empty BK_PUBLIC_REPO"
  else
    bap_image="${BK_PUBLIC_REPO}/blueking/bcs-apiserver-proxy:${APISERVER_PROXY_VERSION}"
  fi

  case "${CRI_TYPE,,}" in
    "docker")
      if ! command -v docker &>/dev/null; then
        utils::log "ERROR" "docker client: docker is not found"
      fi
      docker run -v "${PROXY_TOOL_PATH}":/tmp --rm --entrypoint /bin/cp "${bap_image}" \
        -f /data/bcs/bcs-apiserver-proxy/bcs-apiserver-proxy-tools /tmp/ || utils::log "ERROR" "pull ${bap_image} image failed"
      ;;
    "containerd")
      if ! command -v ctr &>/dev/null; then
        utils::log "ERROR" "containerd client: ctr is not found"
      fi
      if ctr -n k8s.io i pull --hosts-dir "/etc/containerd/certs.d" "${bap_image}"; then
        if ! ctr -n k8s.io run --rm --mount type=bind,src="${PROXY_TOOL_PATH}",dst=/tmp,options=rbind:rw "${bap_image}" \
          bap-copy."$(date +%s)" /bin/cp -f /data/bcs/bcs-apiserver-proxy/bcs-apiserver-proxy-tools /tmp/; then
          utils::log "ERROR" "containerd fail to run ${bap_image}"
        fi
      else
        utils::log "ERROR" "pull ${bap_image} image failed"
      fi
      ;;
    *)
      # ToDo: Unified standard error code
      export ERR_CODE=1
      utils::log "FATAL" "unkown CRI_TYPE: $CRI_TYPE"
      ;;
  esac

  [[ -z "${VIP}" ]] && utils::log "ERROR" "apiserver HA is enabled but VIP is not set"
  "${PROXY_TOOL_PATH}"/bcs-apiserver-proxy-tools -cmd init -vs "${VIP}":"${VS_PORT}" -rs "${K8S_CTRL_IP}":6443 \
    -scheduler "${LVS_SCHEDULER}" -toolPath "${PROXY_TOOL_PATH}"/bcs-apiserver-proxy-tools
  "${ROOT_DIR}"/system/config_bcs_dns -u "${VIP}" k8s-api.bcs.local
  k8s::restart_kubelet
}

safe_source "${ROOT_DIR}/functions/utils.sh"
safe_source "${ROOT_DIR}/functions/k8s.sh"
"${ROOT_DIR}"/system/check_host.sh -c all

"${ROOT_DIR}"/system/config_envfile.sh -c init
"${ROOT_DIR}"/system/config_system.sh -c dns sysctl
"${ROOT_DIR}"/tools/install_tools.sh jq yq
"${ROOT_DIR}"/k8s/install_cri.sh
"${ROOT_DIR}"/k8s/install_k8s_tools
"${ROOT_DIR}"/k8s/render_kubeadm

safe_source "${ROOT_DIR}/env/bcs.env"

case "${K8S_CSI,,}" in
  "localpv")
    "${ROOT_DIR}"/system/mount_localpv
    ;;
  *)
    utils::log "WARN" "unkown csi plugin: $K8S_CSI"
    ;;
esac


# wait kubelet to start
sleep 30
if systemctl is-active kubelet.service -q; then
  utils::log "WARN" "kubelet service is active now, skip kubeadm join"
else
  kubeadm join --config="${ROOT_DIR}/kubeadm-config" -v 11 \
    || utils::log "FATAL" "${LAN_IP} failed to join cluster: ${K8S_CTRL_IP}"
  systemctl enable --now kubelet
fi


if [[ "${ENABLE_APISERVER_HA}" == "true" ]]; then
  if [[ "${APISERVER_HA_MODE}" == "bcs-apiserver-proxy" ]]; then
    init_bap_rule
  else
    "${ROOT_DIR}"/system/config_bcs_dns -u "${VIP}" k8s-api.bcs.local
    k8s::restart_kubelet
  fi
fi
