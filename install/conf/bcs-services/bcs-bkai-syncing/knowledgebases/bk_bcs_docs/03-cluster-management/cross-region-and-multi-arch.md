# 跨云区域与多架构异构集群纳管指南

在复杂企业环境中，常面临**跨网络隔离区域（需通过云区域 Proxy 中转）**以及**同时包含 `x86_64` 与 `arm64` 节点的异构 K8S 集群**纳管场景。本文提供完整的方案说明与操作步骤。

---

## 一、多架构（`x86_64` / `arm64`）异构集群纳管

### 1. 原理与组件说明
当一套 BCS 同时纳管 `x86_64`（amd64）和 `arm64` 架构的 K8S 集群（或同一集群内混合两种架构节点）时，核心依赖两个集群侧组件的跨架构兼容：
- **`bcs-kube-agent`**：BCS 集群代理，负责与 `bcs-cluster-manager` 建立 WebSocket 隧道并上报集群 APIServer 地址。
- **`bcs-k8s-watch`**：负责监听采集集群内 K8S 资源变更并同步至 `bcs-storage`。

BCS 采用 Docker **多架构清单（Manifest List）**机制：同一个镜像 Tag 指向一个包含 `linux/amd64` 和 `linux/arm64` 两个平台镜像 Digest 的索引清单，节点拉取镜像时会自动匹配自身 CPU 架构。

### 2. 构建并推送多架构 Manifest 镜像步骤

若项目私有镜像仓库尚未包含多架构清单，可按以下步骤构建（注意：Docker 19.x 推送 HTTP 仓库存在 `--insecure` 失效的已知 Bug，建议使用 HTTPS 仓库或将 Docker CLI 升级至 20.x+）：

```bash
# 1. 设置目标私有镜像仓库地址与版本
REGISTRY_ADDR="registry.example.com:5000"
VERSION="v1.29.2"

# 2. 将 x86_64 与 arm64 架构的镜像分别打 Tag 并推送到目标仓库
# 涉及组件：bcs-k8s-watch、bcs-kube-agent、bcs-busybox
for comp in bcs-k8s-watch bcs-kube-agent; do
  docker tag ${comp}:${VERSION}-amd64 ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-x86
  docker tag ${comp}:${VERSION}-arm64 ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-arm
  docker push ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-x86
  docker push ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-arm

  # 3. 创建并推送多架构 Manifest 清单（若为 HTTP 仓库可追加 --insecure）
  docker manifest create ${REGISTRY_ADDR}/blueking/${comp}:${VERSION} \
    --amend ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-x86 \
    --amend ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}-arm
  docker manifest push ${REGISTRY_ADDR}/blueking/${comp}:${VERSION}
done

# 4. 查看与验证多架构清单
docker manifest inspect ${REGISTRY_ADDR}/blueking/bcs-kube-agent:${VERSION}
```

### 3. 纳管异构集群时的配置调整

- **方式 A（全局配置）**：修改 BCS 服务端 `bcs-cluster-manager` Deployment 的环境变量 `bcsImageRegistry` 指向上述支持多架构 Manifest 的镜像仓库地址，此后直接通过 Kubeconfig 纳管异构集群即可自动拉取对应架构镜像。
- **方式 B（插件管理单集群更新）**：若通过 Kubeconfig 纳管后 `bcs-k8s-watch` 因镜像架构不匹配启动失败，可进入 BCS 控制台【插件管理】→【组件库】→ 找到该集群的 `bcs-k8s-watch` 点击【更新】，将 `values.yaml` 中的 `image.registry` 和 `global.busybox.registry` 修改为多架构镜像仓库地址并更新。

---

## 二、手动对接跨云区域（Proxy 隔离网）K8S 集群

当被纳管的业务 K8S 集群位于非直连网络区域（如专有云区域、隔离 VPC），无法直接连接 BCS 服务端时，可通过**云区域 Proxy 节点部署 Nginx/OpenResty 反向代理**完成手动纳管。

### 1. 网络策略开通要求

| 源地址 | 目的地址 | 协议 | 目的端口 | 用途说明 |
| :--- | :--- | :--- | :--- | :--- |
| 业务 K8S 集群节点 IP | 云区域 Proxy IP | TCP | `31443` | 必选，`bcs-kube-agent` 建立 WSS 隧道连接 BCS 网关 |
| 业务 K8S 集群节点 IP | 云区域 Proxy IP | TCP | `31024` | 必选，`bcs-k8s-watch` 上报资源数据至 `bcs-storage` |
| 业务 K8S 集群节点 IP | 云区域 Proxy IP | TCP | `5000` | 可选，代理访问 BCS 侧私有镜像仓库（若集群有独立仓库可不开） |
| 云区域 Proxy IP | BCS 部署集群节点 IP | TCP | `31443` / `31024` / `5000` | 必选，Proxy 转发流量至 BCS 服务端 |

### 2. 在 BCS 服务端注册集群并提取凭证

在 **BCS 部署集群的 Master 节点**上执行：

```bash
# 1. 获取 BCS API Gateway Token
bcs_apigateway_token=$(kubectl get secret -n bcs-system bcs-password -o jsonpath="{.data.gateway_token}" | base64 -d)

# 2. 调用接口仅创建集群元数据记录（onlyCreateInfo: true），替换实际 projectID 和 businessID
curl -X POST -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${bcs_apigateway_token}" \
  "http://${BCS_DOMAIN}/bcsapi/v4/clustermanager/v1/cluster" \
  -d '{
    "clusterName": "跨云区域生产集群",
    "provider": "bluekingCloud",
    "region": "default",
    "projectID": "<PROJECT_ID>",
    "businessID": "<BIZ_ID>",
    "environment": "prod",
    "engineType": "k8s",
    "isExclusive": true,
    "creator": "admin",
    "onlyCreateInfo": true,
    "clusterType": "single"
  }'
# 从返回结果中记录生成的 clusterID（例如 BCS-K8S-40000）

# 3. 导出 BCS 客户端与 CA 证书（用于 Proxy 和集群组件双向 TLS 认证）
kubectl get secret -n bcs-system bcs-client-bcs-services-stack -o jsonpath='{.data.tls\.crt}' | base64 -d > /tmp/bcs-client.crt
kubectl get secret -n bcs-system bcs-client-bcs-services-stack -o jsonpath='{.data.tls\.key}' | base64 -d > /tmp/bcs-client.key
kubectl get secret -n bcs-system bcs-client-bcs-services-stack -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/bcs-ca.crt
```

### 3. 在云区域 Proxy 部署 OpenResty/Nginx 代理

将导出的 `bcs-client.crt`、`bcs-client.key`、`bcs-ca.crt` 传输至云区域 Proxy 机器的 `/data/bcs/cert/bcs/` 目录，并配置 OpenResty 代理 `31443`（WSS）和 `31024`（mTLS）端口：

```nginx
events {
    worker_connections 65535;
}
http {
    include       mime.types;
    default_type  application/octet-stream;
    sendfile      on;

    # 代理 bcs-kube-agent WSS 隧道 (31443)
    server {
        listen 31443 ssl;
        ssl_certificate     /data/bcs/cert/bcs/bcs-client.crt;
        ssl_certificate_key /data/bcs/cert/bcs/bcs-client.key;

        location / {
            proxy_pass https://bcs.monitor.com:31443/;
            proxy_set_header Host bcs.monitor.com;
            proxy_ssl_verify off;
            proxy_ssl_trusted_certificate /data/bcs/cert/bcs/bcs-client.crt;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_read_timeout 300s;
            proxy_send_timeout 300s;
            proxy_ssl_server_name on;
        }
    }

    # 代理 bcs-k8s-watch 数据上报 (31024)
    server {
        listen 31024 ssl;
        ssl_certificate        /data/bcs/cert/bcs/bcs-client.crt;
        ssl_certificate_key    /data/bcs/cert/bcs/bcs-client.key;
        ssl_client_certificate /data/bcs/cert/bcs/bcs-ca.crt;
        ssl_verify_client      on;

        location / {
            proxy_pass https://bcs.monitor.com:31024/;
            proxy_ssl_certificate         /data/bcs/cert/bcs/bcs-client.crt;
            proxy_ssl_certificate_key     /data/bcs/cert/bcs/bcs-client.key;
            proxy_ssl_verify              off;
            proxy_ssl_trusted_certificate /data/bcs/cert/bcs/bcs-ca.crt;
            proxy_set_header SSL-Client-Cert $ssl_client_cert;
            proxy_read_timeout 300s;
            proxy_send_timeout 300s;
            proxy_ssl_server_name on;
        }
    }
}
```

### 4. 在业务 K8S 集群部署 `bcs-kube-agent` 与 `bcs-k8s-watch`

在业务集群 Master 上生成 `bcs-kube-agent-values.yaml` 和 `bcs-k8s-watch-values.yaml`（填入集群 ID、Proxy IP、API Gateway Token 及上述证书内容），然后使用 Helm 安装：

```bash
# 安装 bcs-kube-agent
helm upgrade bcs-kube-agent --install --create-namespace \
  -f bcs-kube-agent-values.yaml bcs-kube-agent-1.29.1.tgz -n bcs-system

# 安装 bcs-k8s-watch
helm upgrade bcs-k8s-watch --install --create-namespace \
  -f bcs-k8s-watch-values.yaml bcs-k8s-watch-1.29.1.tgz -n bcs-system

# 检查 Pod 状态
kubectl get pod -n bcs-system
```

---

## 三、重装 BCS 后恢复已纳管集群连接（证书更新）

当 BCS 控制面重新部署但复用原有数据库时，项目、`clusterID` 和 `gateway_token` 均保持不变，但 **BCS 内部 CA 和客户端证书会重新签发**，导致存量跨云区域集群的 Proxy、`bcs-k8s-watch` 和 `bcs-kube-agent` 因证书不匹配离线。此时**无需重建集群**，按以下步骤更新证书即可恢复：

1. **重新导出新证书**：在重装后的 BCS 部署集群 Master 上重新提取 `bcs-client-bcs-services-stack` Secret 中的 `tls.crt`、`tls.key`、`ca.crt`，并用 `openssl x509 -in /tmp/bcs-client.crt -noout -subject -issuer -dates` 校验有效期。
2. **更新云区域 Proxy 证书**：将新证书覆盖到所有 Proxy 节点的 `/data/bcs/cert/bcs/` 目录，执行 `nginx -t && systemctl reload openresty`。
3. **更新业务集群证书并滚动重启**：
   - 将业务集群 `bcs-k8s-watch-values.yaml` 和 `bcs-kube-agent-values.yaml` 中的 `ca_crt`、`tls_crt`、`tls_key` 替换为新证书。
   - 执行 `helm upgrade` 更新 Secret，并显式触发 Pod 滚动重启：
     ```bash
     kubectl rollout restart deployment/bcs-k8s-watch -n bcs-system
     kubectl rollout restart deployment/bcs-kube-agent -n bcs-system
     ```
4. **验证恢复状态**：
   ```bash
   # 检查日志中是否还有 x509 / tls handshake 报错（无输出表示正常）
   kubectl logs -n bcs-system deployment/bcs-k8s-watch --tail=300 | grep -Ei 'x509|certificate|tls handshake' || true
   kubectl logs -n bcs-system deployment/bcs-kube-agent --tail=300 | grep -Ei 'x509|certificate|tls handshake' || true
   ```
