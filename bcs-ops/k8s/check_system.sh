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
CRI_TYPE=$(kubectl get no -o wide | awk 'NR==2{split($NF, a, ":"); print a[1]}')
CRI_TYPE=${CRI_TYPE:-docker}

cacert=${1:-"/etc/kubernetes/pki/etcd/ca.crt"}
cert=${2:-"/etc/kubernetes/pki/etcd/ca.crt"}
key=${3:-"/etc/kubernetes/pki/etcd/ca.key"}


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

check_etcd_status() {
    if [[ $# -ne 4 ]]; then
        utils::log "FATAL" "wrong variable num"
    fi

    local endpoint cacert cert key
    endpoint=${1}
    cacert=${2}
    cert=${3}
    key=${4}

    if [[ ! -f ${cacert} ]] || [[ ! -f ${cert} ]] || [[ ! -f ${key} ]]; then
        utils::log "FATAL" "specified pem file not exist"
    fi

    if ! command -v etcdctl &>/dev/null; then
        # try get etcdctl from etcd container
        case "${CRI_TYPE,,}" in
        "docker")
      if ! command -v docker &>/dev/null; then
        utils::log "ERROR" "docker client: docker is not found"
      fi
            container_id=$(docker ps | awk '/etcd/&&!/pause/{print $1}' | awk 'NR==1{print $1}')
            if ! docker cp "${container_id}":/usr/local/bin/etcdctl /usr/local/bin/; then
                utils::log "FATAL" "can not cp etcdctl command"
            fi
            ;;
        "containerd")
      if ! command -v ctr &>/dev/null; then
        utils::log "ERROR" "containerd client: ctr is not found"
      fi
            mkdir /tmp/container_mount
            container_id=$(ctr -n k8s.io containers ls | awk '/etcd/&&!/pause/{print $1}' | awk 'NR==1{print $1}')
            if [[ -n "${container_id}" ]]; then
                if ! ctr -n k8s.io snapshot mounts /tmp/container_mount "${container_id}" | bash -s; then
                    utils::log "FATAL" "can not mount etcd container"
                fi
                if ! cp /tmp/container_mount/usr/local/bin/etcdctl /usr/local/bin/; then
                    utils::log "FATAL" "can not cp etcdctl command"
                fi
            fi
            umount /tmp/container_mount || job_fail "umount /tmp/container_mount failed"
            ;;
        *)
            utils::log "FATAL" "can not cp etcdctl command"
            ;;
        esac

        chmod 111 /usr/local/bin/etcdctl
    fi

    if ! command -v etcdctl &>/dev/null; then
        utils::log "FATAL" "can not find etcdctl command"
    fi
    export ETCDCTL_API=3

    if ! etcdctl --endpoints https://"${endpoint}":2379 --cacert "${cacert}" --cert "${cert}" --key "${key}" endpoint health &>/dev/null; then
        utils::log "FATAL" "etcd endpoint is not healthy"
    fi

    utils::log "OK" "etcd endpoint is healthy"
    return 0
}

safe_source "${ROOT_DIR}/functions/utils.sh"
# 定义检查的应用
systemapps=(etcd kube-controller-manager kube-scheduler kube-apiserver kube-proxy kube-dns)

# 获取当前集群上下文
context=$(kubectl config current-context)

# 输出检查信息
utils::log "INFO" "Check the system application status of K8s cluster: $context"

# 循环检查每个系统应用
for app in "${systemapps[@]}"; do
    # 获取应用的Pod名称

    case "${app}" in
    etcd | kube-controller-manager | kube-scheduler | kube-apiserver)
        pod_name=$(kubectl get pods -n $CHECK_NAMESAPCE -o jsonpath='{.items[?(@.metadata.labels.component=="'"$app"'")].metadata.name}')
        ;;
    *)
        pod_name=$(kubectl get pods -n $CHECK_NAMESAPCE -o jsonpath='{.items[?(@.metadata.labels.k8s-app=="'"$app"'")].metadata.name}')
        ;;
    esac

    # 检查应用是否存在
    if [[ -z "$pod_name" ]]; then
        utils::log "ERROR" "System application  $app no exist"
        continue
    fi

    for pod in $pod_name; do
        # 获取应用的状态
        status=$(kubectl get pods -n $CHECK_NAMESAPCE "$pod" -o jsonpath='{.status.phase}')
        hostIP=$(kubectl get pods -n $CHECK_NAMESAPCE "$pod" -o jsonpath='{.status.hostIP}')

        if [ "$app" = "etcd" ]; then
            utils::log "INFO" "etcd Cluster status：$(check_etcd_status "$hostIP" "$cacert" "$cert" "$key")"
        fi

        # 输出应用的状态
        case $status in
        Running)
            utils::log "INFO" "System application $app is running, host IP: $hostIP"
            ;;
        Pending)
            utils::log "INFO" "System application $app is in a waiting state, host IP: $hostIP"
            ;;
        Succeeded)
            utils::log "OK" "System application $app has successfully completed, host IP: $hostIP"
            ;;
        Failed)
            utils::log "FATAL" "System application $app failed, host IP: $hostIP"
            ;;
        Unknown)
            utils::log "ERROR" "Unknown status of system application $app, host IP: $hostIP"
            ;;
        *)
            utils::log "ERROR" "Unknown status of system application $app, host IP: $hostIP"
            ;;
        esac
    done
done
