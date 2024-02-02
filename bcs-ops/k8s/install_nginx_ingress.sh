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
    echo "[ERROR]: FAIL to source, missing ${source_file}" >&2
    exit 1
  fi
}

# 加载公共函数变量
source_files=("${ROOT_DIR}/functions/utils.sh" "${ROOT_DIR}/functions/k8s.sh" "${ROOT_DIR}/env/bcs.env")
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
  local chart_path
    chart_path="$(find "${ROOT_DIR}/charts/" -name "ingress-nginx-${ver}.tgz" -type f)"
    if [[ -z $chart_path ]]; then
      utils::log "FATAL" "can't find ingress-nginx chart in ${ROOT_DIR}/version-${K8S_VER}/charts/"
    fi
  local registry
  if [[ -n ${BK_PUBLIC_REPO} ]]; then
    registry="${BK_PUBLIC_REPO}/registry.k8s.io"
  else
    registry="registry.k8s.io"
  fi

  utils::log "INFO" "installing ingress-nginx"

  cat <<EOF | helm upgrade --install ingress-nginx "${chart_path}" --version "$ver" -n $namespace --debug -f -
controller:
  metrics:
    enabled: true
  image:
    registry: "${registry}"
    digest: ""
  config:
    # log format is consistent with the filebeat collection configuration
    log-format-upstream: '\$remote_addr - \$remote_user [\$time_local] "\$request" \$status \$body_bytes_sent "\$http_referer" "\$http_user_agent" \$request_length \$request_time [\$proxy_upstream_name] [\$proxy_alternative_upstream_name] \$upstream_addr \$upstream_response_length \$upstream_response_time \$upstream_status \$req_id'
    # The number of requests that can be handled by a long connection maintained by nginx and the client, the default is 100
    # ref: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#keep-alive-requests
    keep-alive-requests: "10000"
    # The maximum number of idle connections between nginx and upstream to maintain a long connection, the default is 32
    # ref: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#upstream-keepalive-connections
    upstream-keepalive-connections: "200"
    # The maximum number of connections that each worker process can open, the default is 16384.
    # ref: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#max-worker-connections
    max-worker-connections: "65536"
    # upload file need
    proxy-body-size: "2G"
    proxy-read-timeout: "600"
  service:
    type: NodePort
    nodePorts:
      http: 32080
      https: 32443
  hostNetwork: false
  ingressClassResource:
      enabled: true
      default: true
  admissionWebhooks:
    patch:
      image:
        registry: ${registry}
        digest: ""
EOF
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
