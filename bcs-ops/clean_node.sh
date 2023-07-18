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

BAK_DIR=${BAK_DIR:-"/data/backup/$(date +%s)"}
SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR=${SELF_DIR}

readonly BAK_DIR SELF_DIR ROOT_DIR

clean_container() {
  case "${CRI_TYPE,,}" in
    "containerd")
      ctr -n k8s.io t ls | grep -qv PID && ctr -n k8s.io t rm -f "$(ctr -n k8s.io t ls -q)"
      ctr -n k8s.io c ls | grep -qv CONTAINER && ctr -n k8s.io c rm "$(ctr -n k8s.io c ls -q)"
      systemctl disable --now containerd
      ;;
    "docker")
      docker ps | grep -qv NAME && docker rm -f "$(docker ps -aq)"
      systemctl disable --now docker
      ;;
  esac
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

source_files=("${ROOT_DIR}/functions/utils.sh" "${ROOT_DIR}/env/bcs.env")
for file in "${source_files[@]}"; do
  safe_source "$file"
done

systemctl disable --now kubelet
kubeadm reset phase cleanup-node
clean_container

ip l d cni0 || utils::log "WARN" "link cni0 does not exist"
ip l d kube-ipvs0 || utils::log "WARN" "link kube-ipvs0 does not exist"

utils::log "INFO" "Backing Files"
install -dv "${BAK_DIR}" || utils::log "FATAL" "create backup dir $BAK_DIR failed"
[[ -d /etc/kubernetes ]] && mv -v /etc/kubernetes "$BAK_DIR"/
[[ -d /var/lib/kubelet ]] && mv -v /var/lib/kubelet "$BAK_DIR"/
[[ -d ${KUBELET_LIB} ]] && mv -v "${KUBELET_LIB}" "$BAK_DIR"/kubelet
[[ -d "$HOME"/.kube ]] && mv -v "$HOME"/.kube "$BAK_DIR"/
[[ -d ${ETCD_LIB} ]] && mv -v "${ETCD_LIB}" "$BAK_DIR"/
[[ -L /var/lib/etcd ]] && rm -vf /var/lib/etcd
[[ -d /var/lib/etcd ]] && mv -v /var/lib/etcd "$BAK_DIR"/
utils::log "OK" "Back Files >>> Done"

"${ROOT_DIR}"/k8s/operate_completion clean \
  && utils::log "OK" "Uninstall kubeadm kubelet >>> Done"
"${ROOT_DIR}"/k8s/uninstall_k8s_tools \
  && utils::log "OK" "Uninstall kubeadm kubelet >>> Done"
"${ROOT_DIR}"/k8s/uninstall_cri.sh \
  && utils::log "OK" "Uninsatll cri runtime >>> Done"
"${ROOT_DIR}"/system/config_system.sh -d dns sysctl \
  && utils::log "OK" "Clean bcs hosts entry,bcs sysctl >>> Done"
"${ROOT_DIR}"/system/config_envfile.sh -c clean \
  && utils::log "OK" "Clean bcs.env"
