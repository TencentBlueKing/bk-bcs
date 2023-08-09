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

NGINX_INGRESS_VER=${NGINX_INGRESS_VER:-"4.2.5"}
NGINX_INGRESS_URL=${NGINX_INGRESS_URL:-"https://kubernetes.github.io/ingress-nginx"}

TIMEOUT=180s
NAMESPACE=ingress-nginx
SELF_DIR=$(dirname "$(readlink -f "$0")")
ROOT_DIR="${SELF_DIR}/.."

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

# 加载公共函数变量
source_files=("${ROOT_DIR}/functions/utils.sh" "${ROOT_DIR}/env/bcs.env")
for file in "${source_files[@]}"; do
  safe_source "$file"
done

check_dependency() {
  if ! command -v "$1" &>/dev/null; then
    echo "Error: $1 not found. Please install it and try again."
    exit 1
  fi
}

install_nginx_ingress() {
  local ver=$NGINX_INGRESS_VER
  local namespace=$NAMESPACE

  helm repo add ingress-nginx "$NGINX_INGRESS_URL"
  helm repo update
  helm install ingress-nginx ingress-nginx/ingress-nginx --version "$ver" -n $namespace

}

check_k8s_status() {
  if ! kubectl cluster-info 2>/dev/null; then
    utils::log "FATAL" "fail to get k8s cluster info"
  fi
  return 0

}

main() {

  # 检查kubectl是否安装
  check_dependency kubectl
  # 检查helm是否安装
  check_dependency helm
  # 检查集群状态
  check_k8s_status

  install_nginx_ingress

  echo "waiting Nginx Ingress to be install..."
  kubectl wait --namespace $NAMESPACE --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller --timeout=$TIMEOUT

  # 显示部署结果
  echo "Deploy Nginx Ingress sucess, detail:"
  kubectl get svc -n $NAMESPACE

}

main

exit $?
