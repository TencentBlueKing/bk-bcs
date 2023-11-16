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

# generic k8s function
# depend on utils.sh
# independent of business

#######################################
# add helmrepo safely
# Arguments:
# $1: repo_name
# $2: repo_url
# Return:
# can't find helm - return 0
# helm update success - return 0
# helm update fail - return 1
#######################################
k8s::safe_add_helmrepo() {
  if ! command -v helm &>/dev/null; then
    utils::log "WARN" "Did helm installed?"
    return 1
  fi

  local repo_name repo_url
  repo_name=$1
  repo_url=$2
  if helm repo list | grep -q "$repo_name"; then
    utils::log "INFO" "remove old helm repo: $repo_name"
    helm repo remove "$repo_name"
  fi
  if ! helm repo add "$repo_name" "$repo_url"; then
    utils::log "ERROR" "can't add helm repo $repo_name $repo_url"
    return 1
  fi
  helm repo list
  if ! helm repo update; then
    utils::log "ERROR" "can't update helm repo"
    return 1
  fi
  return 0
}

#######################################
# add vip to K8S apiserver certs
# Arguments:
# $1: vip
# Return:
# add vip success - return 0
# add vip fail - return 1
#######################################
k8s::add_vip_to_cert() {
  vip=$1
  local kubeadm_config_file
  kubeadm_config_file="/tmp/kubeadm-$(date +%Y-%m-%d).yaml"
  kubectl -n kube-system get configmap kubeadm-config -o jsonpath='{.data.ClusterConfiguration}' >"${kubeadm_config_file}"
  if grep -q certSANs "${kubeadm_config_file}"; then
    sed -i "/certSANs/a \  \- \"${vip}\"" "${kubeadm_config_file}"
  else
    sed -i "/apiServer:/a \  certSANs:\n  - \"${vip}\"" "${kubeadm_config_file}"
  fi
  install -dv "/etc/kubernetes/pki/backup-$(date +%Y-%m-%d)"
  mv -f /etc/kubernetes/pki/apiserver.{crt,key} "/etc/kubernetes/pki/backup-$(date +%Y-%m-%d)"
  kubeadm init phase certs apiserver --config "${kubeadm_config_file}" \
    || utils::log "ERROR" "failed to add ${vip} to apiserver cert"
  rm -f "${kubeadm_config_file}"
  utils::log "OK" "added ${vip} to apiserver cert"
}

#######################################
# restart kubelet service
# Return:
# restart kubelet service success - return 0
# restart kubelet service fail - return 1
#######################################
k8s::restart_kubelet() {
  if systemctl restart kubelet.service &>/dev/null; then
    utils::log "INFO" "kubelet service restarted"
    utils::log "INFO" "checking kubelet service status..."
    sleep 10
    if systemctl is-active kubelet.service -q; then
      utils::log "OK" "kubelet service is active now"
      return 0
    fi
    utils::log "ERROR" "kubelet service is inactive"
    return 1
  fi
  utils::log "ERROR" "restart kubelet service failed"
  return 1
}

k8s::check_master() {
  local timeout=5
  while ((timeout > 0)); do
    if ! kubectl cluster-info; then
      utils::log "WARN" "timeout=$timeout, \
controlplane has not been started yet, waiting"
      crictl ps
    else
      return 0
    fi
	sleep 30
  done
  return 1
}
