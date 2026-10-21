# bcs-ingress-controller 与负载均衡端口池（PortPool）指南

`bcs-ingress-controller`（API Group: `networkextension.bkbcs.tencent.com/v1`）是蓝鲸容器管理平台（BCS）自研的多云四层/七层负载均衡控制器，目前已支持 **腾讯云 TKE（CLB）**、**亚马逊云 EKS（ALB/NLB）** 和 **谷歌云 GKE**。

相比社区原生 Ingress，它不仅支持 HTTP/HTTPS 七层路由，还深度支持 **TCP/UDP 四层转发**、**NodePort 与直通 Pod 双模式**、**多 Service / 同 Service 内多版本 Subset 权重配比（蓝绿/灰度切流）**、**StatefulSet/GameStatefulSet 端口段映射**、**随机 HostPort 注入**、**跨地域绑定（SnatPro）** 以及面向游戏 Dedicated Server 的 **PortPool 动态端口池分配**。

---

## 一、组件部署与权限要求（以腾讯云 CLB 为例）

### 1. 腾讯云 CAM 账号权限要求
部署 `bcs-ingress-controller` 所使用的腾讯云子账号 `SecretID` / `SecretKey` 需具备：
- **预设策略推荐**：直接绑定 `CLB 全读写权限（QcloudCLBFullAccess）` + `CVM 只读权限（QcloudCVMReadOnlyAccess）`。
- **安全组要求**：
  - **CLB 安全组**：放通外部客户端访问 CLB 监听端口。
  - **CVM 节点安全组**：放通 CLB 访问节点 NodePort 端口（或直通模式下的 Pod 容器端口）。

### 2. 在 BCS 组件库中启用
1. 将 `SecretID` 和 `SecretKey` 进行 Base64 编码：
   ```bash
   echo -n "<SecretID>" | base64
   echo -n "<SecretKey>" | base64
   ```
2. 进入 BCS 控制台【插件管理】→【组件库】→ 选择目标集群，找到 `BcsIngressController` 点击【启用】，填入集群所在地域（如 `ap-guangzhou`）及 Base64 编码后的密钥完成部署。

---

## 二、`Ingress`（`networkextension.bkbcs.tencent.com/v1`）典型场景配置

### 场景 1：CLB 转发至 Service NodePort（适用于 Overlay 网络集群）

将 CLB 的 `38080` TCP 端口流量转发至集群内 `test-svc` 的 NodePort，再路由至 Pod：

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: ingress-nodeport-demo
  namespace: default
  annotations:
    # 关联一个或多个腾讯云 CLB 实例 ID（支持跨地域格式 ap-shanghai:lb-xxxx）
    networkextension.bkbcs.tencent.com/lbids: "lb-xxxxxxx"
spec:
  rules:
    - port: 38080
      protocol: TCP
      listenerAttribute:
        sessionTime: 100
        lbPolicy: WRR
        healthCheck:
          enabled: true
      services:
        - serviceName: test-svc
          serviceNamespace: default
          servicePort: 8080
```

### 场景 2：CLB 直通 Pod（`isDirectConnect: true`，适用于 VPC-CNI Underlay 网络）

绕过 NodePort 转发开销，CLB 直接将流量转发给 Pod IP（注意：对应工作负载的容器配置中必须显式声明 `ports[].containerPort`）：

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: ingress-direct-pod
  namespace: default
  annotations:
    networkextension.bkbcs.tencent.com/lbids: "lb-xxxxxxx"
spec:
  rules:
    - port: 39090
      protocol: TCP
      services:
        - serviceName: test-svc
          serviceNamespace: default
          servicePort: 8080
          isDirectConnect: true
```

### 场景 3：七层 HTTP / HTTPS（单向认证）域名与路径路由

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: ingress-https-demo
  namespace: default
  annotations:
    networkextension.bkbcs.tencent.com/lbids: "lb-xxxxxxx"
spec:
  rules:
    - port: 443
      protocol: HTTPS
      certificate:
        mode: UNIDIRECTIONAL
        certID: "cert-xxxxxxx"
      layer7Routes:
        - domain: game.example.com
          path: /api
          listenerAttribute:
            lbPolicy: LEAST_CONN
          services:
            - serviceName: api-svc
              serviceNamespace: default
              servicePort: 8080
```

### 场景 4：多 Service 或同 Service 内多 Subset 权重配比（蓝绿/金丝雀切流）

#### （1）不同 Service 之间按权重分流（如 `svc-v1` 80%，`svc-v2` 20%）

```yaml
spec:
  rules:
    - port: 38080
      protocol: TCP
      services:
        - serviceName: svc-v1
          serviceNamespace: default
          servicePort: 8080
          weight:
            value: 80
        - serviceName: svc-v2
          serviceNamespace: default
          servicePort: 8080
          weight:
            value: 20
```

#### （2）直通模式下同一 Service 内按 Pod Label 子集（`subsets`）设置不同 RS 权重

```yaml
spec:
  rules:
    - port: 8080
      protocol: TCP
      services:
        - serviceName: nginx-svc
          serviceNamespace: default
          servicePort: 8080
          isDirectConnect: true
          subsets:
            - labelSelector:
                version: v1
              weight:
                value: 10
            - labelSelector:
                version: v2
              weight:
                value: 20
```

### 场景 5：`StatefulSet` / `GameStatefulSet` 端口段固定映射（`portMappings`）

适用于有状态游戏服每个固定序号 Pod 需要独立暴露一个或一段公网 TCP/UDP 端口的场景（例如 `gameserver-0` 对应 `30000~30004`，`gameserver-1` 对应 `30005~30009`）：

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: Ingress
metadata:
  name: sts-port-mapping
  namespace: game
  annotations:
    networkextension.bkbcs.tencent.com/lbids: "lb-xxxxxxx"
spec:
  portMappings:
    - startPort: 30000
      protocol: UDP
      # [startIndex, endIndex) 左闭右开区间，0~6 代表映射序号 0~5 共 6 个 Pod
      startIndex: 0
      endIndex: 6
      segmentLength: 5
      workloadKind: StatefulSet # 支持 StatefulSet 或 GameStatefulSet
      workloadName: gameserver
      workloadNamespace: game
      rsStartPort: 8080
```

---

## 三、负载均衡动态端口池（`PortPool`）使用指南

在无状态/动态扩缩的游戏开房间场景中，无需事先规划固定序号映射，可通过 **`PortPool`（端口池）** 为任意新创建的 Pod 自动分配 CLB 端口（或端口段），并将分配到的公网 VIP 与端口注入到容器环境变量与 DownwardAPI 挂载文件中。

### 第一步：创建 `PortPool` 端口池资源

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: PortPool
metadata:
  name: game-portpool
  namespace: game
spec:
  poolItems:
    - itemName: telecom-item
      loadBalancerIDs: ["lb-00001"]
      # 协议可选 TCP, UDP；腾讯云不填默认同时监听 TCP,UDP；AWS 必须显式指定单一协议
      protocol: "TCP"
      startPort: 31000
      endPort: 31100
      # 可选：若需为每个 Pod 分配连续 5 个端口的端口段，可设置 segmentLength: 5
      segmentLength: 0
```

创建后执行 `kubectl describe portpool game-portpool -n game`，等待 `status.poolItems[].status` 变为 **`Ready`**。

### 第二步：为目标命名空间打上端口池注入标签（必做！）

```bash
kubectl label ns game bcs-ingress-controller-inject=true
```

### 第三步：在工作负载的 `spec.template.metadata.annotations` 中声明使用端口池

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gameserver-room
  namespace: game
spec:
  replicas: 2
  selector:
    matchLabels:
      app: gameserver-room
  template:
    metadata:
      labels:
        app: gameserver-room
      annotations:
        # 1. 必选：开启端口池注入
        portpools.networkextension.bkbcs.tencent.com: "true"
        # 2. 必选：指定 "{端口池名} {协议(TCP/UDP/TCP_UDP)} {容器端口或端口名}"
        ports.portpools.networkextension.bkbcs.tencent.com: "game-portpool TCP 8080"
        # 3. 可选：在 CLB 端口绑定变为 Ready 前将 Pod ReadinessGate 保持为 NotReady
        readinessgate.portpools.networkextension.bkbcs.tencent.com: "true"
        # 4. 可选：同名 Pod（如 StatefulSet）重建后复用原 VIP 和端口的保留时长
        keepduration.portbinding.bkbcs.tencent.com: "5m"
    spec:
      containers:
        - name: server
          image: hub.bktencent.com/blueking/game-demo:0.0.1
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: podinfo
              mountPath: /etc/podinfo
      volumes:
        - name: podinfo
          downwardAPI:
            items:
              - path: annotations
                fieldRef:
                  fieldPath: metadata.annotations
```

### 第四步：业务容器内读取分配到的公网 VIP 与端口

Pod 创建后，`bcs-ingress-controller` 会自动完成两项注入：
1. **环境变量注入**：自动向容器注入 `BCS_PORTPOOL_PORT_VIPLIST_<协议>_<容器端口>`（例如 `BCS_PORTPOOL_PORT_VIPLIST_TCP_8080=1.1.1.1:31000`）。
2. **Pod 注解与 DownwardAPI 文件通知**：
   - `status.portpools.networkextension.bkbcs.tencent.com`：绑定成功后由 `NotReady` 变为 **`Ready`**。
   - `poolbindings.portpool.networkextension.bkbcs.tencent.com`：包含负载均衡器 ID、VIP 列表、分配的 `startPort`/`endPort` 等完整 JSON 信息。业务进程可监听 `/etc/podinfo/annotations` 文件感知绑定完成。

> **删除端口池注意事项**：若某个 `PortPool` 已有端口被 Pod 占用，禁止直接删除。需先执行 `kubectl get portbinding --all-namespaces` 查看占用端口的 Pod，清理对应工作负载释放端口后再删除 `PortPool`。

---

## 四、随机 HostPort 注入方案（`randhostport`）

在无云负载均衡器或裸金属自建集群中，若希望将 Pod 端口随机映射到宿主机空闲端口并实现同节点端口防冲突调度，可在 Pod Template 中添加以下注解（需集群已启用 `bcs-webhook-server` 的 `randhostport` 插件）：

```yaml
metadata:
  annotations:
    randhostport.webhook.bkbcs.tencent.com: "true"
    ports.randhostport.webhook.bkbcs.tencent.com: "8080,metric"
```
Webhook 会自动为容器端口分配宿主机随机 `hostPort`，自动注入 `podAntiAffinity` 防止同端口 Pod 调度到同一节点，并通过环境变量 `BCS_RANDHOSTPORT_HOSTIP` 和 `BCS_RANDHOSTPORT_FOR_CONTAINER_PORT_8080` 告知容器内业务进程。

---

## 五、`bcs-ingress-controller` 常用排查命令与常见问题

### 1. 常用诊断命令

```bash
# 查看 Ingress、监听器、端口池与端口绑定状态
kubectl get ingress.networkextension.bkbcs.tencent.com -n <namespace>
kubectl get listener -n <namespace>
kubectl get portpool -n <namespace>
kubectl get portbinding -n <namespace>

# 查看控制器运行状态与错误日志
kubectl get pod -n bcs-system | grep ingress-controller
kubectl logs <ingress-controller-pod> -n bcs-system --tail=200
```

### 2. 常见报错速查表

| 现象或日志报错 | 根本原因 | 解决方法 |
| :--- | :--- | :--- |
|Workload 声明了端口池注解，但 Pod 未分配到 CLB 端口 | 1. 命名空间未打注入标签；<br/>2. 注解写在了 Deployment 顶层 `metadata` 而非 `spec.template.metadata` 下 | 1. 执行 `kubectl label ns <ns> bcs-ingress-controller-inject=true`；<br/>2. 将 `portpools` 注解移至 `spec.template.metadata.annotations` |
| 控制器 Pod 启动失败：`setenv: invalid argument: unknown` 或日志报 `invalid header field value "TC3-HMAC-SHA256..."` | 启用组件时填写的腾讯云 `SecretID` / `SecretKey` 未做 Base64 编码或含换行符 | 使用 `echo -n "<SecretID>" \| base64` 重新编码并在组件库更新配置 |
| 日志报错 `has no permission` | 云账号缺少 CLB 或 CVM 查询权限 | 为子账号绑定 CLB 全读写权限及 `cvm:DescribeInstances` 权限 |
| 日志报错 `loadbalancer not found` | 控制器默认地域与 CLB 实际所在地域不一致 | 修改组件 Values 中的 `tencentcloudRegion`，或在 `lbids` 注解中显式加上地域前缀 `ap-xxx:lb-xxx` |
| 日志报错 `is not add to global route` 或 `Not support to register other ip before enable SnatPro` | CLB 与集群不在同一 VPC/地域，且未开启跨域绑定 | 使用与集群同 VPC 的 CLB，或为 CLB 开启 **SnatPro（跨域绑定 2.0）** 并配置 SNAT IP |
| 日志报错 `uin not allow to create port range listener` | 腾讯云账号未开通公网 CLB 端口段监听器白名单 | 联系腾讯云技术支持为账号 UIN 添加端口段白名单 |
| CLB 监听器已绑定后端且健康检查正常，但客户端无法通过 VIP 访问 | CLB 安全组未放通客户端 IP，或 CVM 节点安全组未放通 CLB 流量 | 检查并放通 CLB 安全组入站规则及 CVM 安全组规则 |
