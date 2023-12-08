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

clean_container() {
  crictl ps -aq | xargs -r crictl rm -f
}

clean_cni() {
  case ${K8S_CNI} in
    "flannel")
      ip l | awk '/flannel/{eth=$2;gsub(":","",eth);print eth}' | xargs -r -n 1 ip l d
      ;;
    *)
      return 0
      ;;
  esac
  rm -rf /etc/cni/net.d/*
}

clean_vni() {
  ip l | awk '/cni0|kube-ipvs0/{eth=$2;gsub(":","",eth);print eth}' | xargs -r -n 1 ip l d
  ip l | awk '/veth/{eth=$2;split(eth,a,"@");print a[1]}' | xargs -r -n 1 ip l d
}

utils::log "INFO" "reseting kubelet"
kubeadm reset phase cleanup-node \
  --cri-socket "$(crictl config --get runtime-endpoint)" --v=5
systemctl disable --now kubelet

utils::log "INFO" "cleaning remaining containers"
clean_container


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

"${ROOT_DIR}"/system/config_iptables.sh clean \
  && utils::log "OK" "Clean k8s-components iptables rules"

utils::log "INFO" "cleaning remain kubelet mount path"
df -h | awk '#'"${BAK_DIR}"'/kubelet#{print $NF}' | xargs -r umount

utils::log "INFO" "cleaning remain virtual interface"
clean_cni
clean_vni
