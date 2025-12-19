#!/bin/bash

# 生成自签名证书用于 mesh-proxy 测试

set -e

CERT_DIR="./certs"
CERT_FILE="$CERT_DIR/tls.crt"
KEY_FILE="$CERT_DIR/tls.key"

echo "=== 生成自签名证书 ==="

# 创建证书目录
mkdir -p "$CERT_DIR"

# 生成私钥
echo "正在生成私钥..."
openssl genrsa -out "$KEY_FILE" 2048

# 生成证书签名请求
echo "正在生成证书签名请求..."
openssl req -new -key "$KEY_FILE" -out /tmp/cert.csr -subj "/CN=mesh-proxy/O=mesh-proxy"

# 生成自签名证书
echo "正在生成自签名证书..."
openssl x509 -req -in /tmp/cert.csr -signkey "$KEY_FILE" -out "$CERT_FILE" -days 36500

# 清理临时文件
rm -f /tmp/cert.csr

echo "证书已生成:"
echo "  证书文件: $CERT_FILE"
echo "  密钥文件: $KEY_FILE"

echo ""
echo "现在可以使用以下命令创建 Kubernetes Secret:"
echo "  ./scripts/setup-tls.sh $CERT_FILE $KEY_FILE"

echo "=== 证书生成完成 ===" 