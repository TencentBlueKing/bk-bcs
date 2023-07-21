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



check_dependency() {
    if ! command -v $1 &>/dev/null; then
        echo "Error: $1 is not installed. Please install it and try again."
        exit 1
    fi
}


install_nginx_ingress() {
    local ver=$NGINX_INGRESS_VER
    local namespace=$NAMESPACE
    
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    helm repo update
    helm install ingress-nginx ingress-nginx/ingress-nginx  --version $ver -n $namespace
    
}


main {

    # 检查kubectl是否安装
    check_dependency kubectl
    # 检查helm是否安装
    check_dependency helm

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
