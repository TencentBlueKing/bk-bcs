#!/bin/bash

# mesh-proxy TLS 证书设置脚本

set -e

NAMESPACE="istio-system"
SECRET_NAME="bcs-mesh-proxy-tls"

echo "=== mesh-proxy TLS 证书设置 ==="

# 检查参数
if [ $# -eq 0 ]; then
    echo "用法: $0 <证书文件路径> <密钥文件路径>"
    echo "示例: $0 ./certs/tls.crt ./certs/tls.key"
    exit 1
fi

CERT_FILE="$1"
KEY_FILE="$2"

# 检查文件是否存在
if [ ! -f "$CERT_FILE" ]; then
    echo "错误: 证书文件不存在: $CERT_FILE"
    exit 1
fi

if [ ! -f "$KEY_FILE" ]; then
    echo "错误: 密钥文件不存在: $KEY_FILE"
    exit 1
fi

echo "正在读取证书文件: $CERT_FILE"
echo "正在读取密钥文件: $KEY_FILE"

# 读取并 base64 编码证书和密钥
CERT_B64=$(cat "$CERT_FILE" | base64 -w 0)
KEY_B64=$(cat "$KEY_FILE" | base64 -w 0)

# 创建 Secret YAML
cat > /tmp/mesh-proxy-tls-secret.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: $SECRET_NAME
  namespace: $NAMESPACE
  labels:
    app: bcs-mesh-proxy
type: kubernetes.io/tls
data:
  tls.crt: $CERT_B64
  tls.key: $KEY_B64
EOF

echo "正在创建/更新 Secret: $SECRET_NAME"
kubectl apply -f /tmp/mesh-proxy-tls-secret.yaml

echo "Secret 已创建/更新成功！"
echo "现在可以部署启用 HTTPS 的 mesh-proxy:"
echo "kubectl apply -f deploy/kubernetes/deployment-https.yaml"

# 清理临时文件
rm -f /tmp/mesh-proxy-tls-secret.yaml

echo "=== 设置完成 ===" 