#!/bin/bash

# BCS Mesh Proxy 部署脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        print_error "$1 命令未找到，请先安装"
        exit 1
    fi
}

# 检查必要的命令
print_info "检查必要的命令..."
check_command kubectl
check_command docker

# 构建镜像
print_info "构建 Docker 镜像..."
docker build -t mesh-proxy:latest .

# 部署到 Kubernetes
print_info "部署到 Kubernetes..."

# 检查命名空间是否存在
if ! kubectl get namespace istio-system &> /dev/null; then
    print_warn "istio-system 命名空间不存在，正在创建..."
    kubectl create namespace istio-system
fi

# 部署 mesh-proxy
print_info "部署 mesh-proxy..."
kubectl apply -f deploy/kubernetes/deployment.yaml

# 等待 Pod 就绪
print_info "等待 Pod 就绪..."
kubectl wait --for=condition=ready pod -l app=mesh-proxy -n istio-system --timeout=300s

# 注册 APIService
print_info "注册 APIService..."
kubectl apply -f deploy/kubernetes/apiservice.yaml

# 检查部署状态
print_info "检查部署状态..."
kubectl get pods -n istio-system -l app=mesh-proxy
kubectl get apiservice | grep istio

print_info "部署完成！"
print_info "您可以使用以下命令检查服务状态："
echo "  kubectl get pods -n istio-system -l app=mesh-proxy"
echo "  kubectl get apiservice | grep istio"
echo "  kubectl get virtualservices -n istio-system" 