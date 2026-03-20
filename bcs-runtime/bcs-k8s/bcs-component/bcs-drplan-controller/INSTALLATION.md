# BCS DR Plan Controller - 安装指南

## 📋 前置要求

- Kubernetes 集群 1.19+
- kubectl 已配置并可访问集群
- (可选) Helm 3.0+ - 用于 Helm 安装方式

## 🚀 安装方式

### 方式 1: 使用 Helm Chart（推荐）

#### 快速安装

```bash
# 创建命名空间
kubectl create namespace bcs-system

# 使用 Helm 安装
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system
```

#### 自定义安装

```bash
# 创建自定义 values 文件
cat > custom-values.yaml <<EOF
image:
  repository: your-registry.com/bcs-drplan-controller
  tag: v1.0.0

controller:
  logLevel: 4  # 启用 Debug 日志
  reconcileFrequency: 30s

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# 启用 Prometheus 监控
metrics:
  serviceMonitor:
    enabled: true
    interval: 30s
EOF

# 安装
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  -f custom-values.yaml
```

#### 生产环境推荐配置

```yaml
# prod-values.yaml
replicaCount: 2  # 高可用

controller:
  logLevel: 0  # Info 级别
  leaderElection: true
  reconcileFrequency: 30s

resources:
  requests:
    cpu: 200m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Pod 反亲和性，避免调度到同一节点
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: bcs-drplan-controller
        topologyKey: kubernetes.io/hostname

# 监控配置
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
    additionalLabels:
      prometheus: kube-prometheus
```

### 方式 2: 使用 Kustomize

```bash
# 安装 CRDs
make install

# 构建镜像
make docker-build IMG=your-registry/bcs-drplan-controller:v1.0.0
make docker-push IMG=your-registry/bcs-drplan-controller:v1.0.0

# 部署到集群
make deploy IMG=your-registry/bcs-drplan-controller:v1.0.0
```

### 方式 3: 手动部署

```bash
# 1. 创建命名空间
kubectl create namespace bcs-system

# 2. 安装 CRDs
kubectl apply -f config/crd/bases/

# 3. 创建 ServiceAccount 和 RBAC
kubectl apply -f config/rbac/ -n bcs-system

# 4. 创建 Webhook 证书
# 使用 Helm 安装方式会自动生成证书（推荐）
# 如果手动部署，需要手动创建证书:
kubectl create secret tls bcs-drplan-controller-webhook-cert \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace bcs-system

# 5. 部署 Controller
kubectl apply -f config/manager/ -n bcs-system
```

## ✅ 验证安装

### 1. 检查 CRD 安装

```bash
kubectl get crd | grep dr.bkbcs.tencent.com
```

预期输出：
```
drplanexecutions.dr.bkbcs.tencent.com   2026-02-03T02:19:01Z
drplans.dr.bkbcs.tencent.com            2026-02-03T02:19:01Z
drworkflows.dr.bkbcs.tencent.com        2026-02-03T02:19:01Z
```

### 2. 检查 Controller 运行状态

```bash
# 使用 Helm 安装时
kubectl get pods -n bcs-system -l app.kubernetes.io/name=bcs-drplan-controller

# 使用 Kustomize 安装时
kubectl get pods -n bcs-drplan-controller-system
```

预期输出：
```
NAME                                      READY   STATUS    RESTARTS   AGE
bcs-drplan-controller-xxxxxxxxxx-xxxxx    1/1     Running   0          1m
```

### 3. 检查 Controller 日志

```bash
kubectl logs -n bcs-system -l control-plane=controller-manager -f
```

预期看到：
```
I0203 02:19:05.123456       1 main.go:95] Starting DR Plan Controller
I0203 02:19:05.123456       1 main.go:96] Logging configuration: Info=default, Debug=V(4)
...
I0203 02:19:06.123456       1 main.go:238] All action executors registered successfully
I0203 02:19:06.123456       1 main.go:243] Executors initialized successfully
I0203 02:19:06.123456       1 main.go:271] Controllers and webhooks registered successfully
```

### 4. 检查 Webhook 配置

```bash
kubectl get mutatingwebhookconfiguration | grep drplan
kubectl get validatingwebhookconfiguration | grep drplan
```

### 5. 测试创建资源

```bash
# 创建示例 DRWorkflow
kubectl apply -f config/samples/drworkflow-http.yaml

# 检查状态
kubectl get drworkflow http-healthcheck -o jsonpath='{.status.phase}'
# 预期输出: Ready
```

## 🔄 升级

### 使用 Helm 升级

```bash
# 升级到新版本
helm upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  -f your-values.yaml

# 查看变更
helm diff upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  -f your-values.yaml
```

### 使用 Kustomize 升级

```bash
# 更新镜像版本
make deploy IMG=your-registry/bcs-drplan-controller:v1.1.0
```

## 🗑️ 卸载

### 使用 Helm 卸载

```bash
# 卸载 Controller（保留 CRD）
helm uninstall bcs-drplan-controller --namespace bcs-system

# 删除 CRD（注意：会删除所有 DRWorkflow、DRPlan、DRPlanExecution 资源）
kubectl delete crd drworkflows.dr.bkbcs.tencent.com
kubectl delete crd drplans.dr.bkbcs.tencent.com
kubectl delete crd drplanexecutions.dr.bkbcs.tencent.com

# 删除命名空间
kubectl delete namespace bcs-system
```

### 使用 Kustomize 卸载

```bash
# 卸载 Controller
make undeploy

# 卸载 CRDs
make uninstall
```

## 🔧 故障排查

### Controller 无法启动

1. **检查镜像是否可拉取**
```bash
kubectl describe pod -n bcs-system -l control-plane=controller-manager
```

2. **检查 RBAC 权限**
```bash
kubectl get clusterrole | grep drplan
kubectl get clusterrolebinding | grep drplan
```

3. **检查日志错误**
```bash
kubectl logs -n bcs-system -l control-plane=controller-manager --tail=100
```

### Webhook 证书问题

1. **检查证书是否存在**
```bash
kubectl get secret -n bcs-system | grep webhook-cert
```

2. **检查证书内容**
```bash
kubectl get secret bcs-drplan-controller-webhook-cert -n bcs-system -o yaml
```

3. **重新生成证书**（使用 Helm upgrade）
```bash
# Helm 会在每次 upgrade 时重新生成证书（如果 autoGenerate=true）
helm upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system
```

4. **使用自定义证书**
```bash
# 生成自签名证书
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout tls.key -out tls.crt \
  -days 365 \
  -subj "/CN=bcs-drplan-controller-webhook-service.bcs-system.svc"

# Base64 编码
CERT_PEM=$(cat tls.crt | base64 -w 0)
KEY_PEM=$(cat tls.key | base64 -w 0)
CA_BUNDLE=$(cat tls.crt | base64 -w 0)  # 自签名证书，CA 和证书相同

# 使用 Helm 部署
helm install bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --namespace bcs-system \
  --set webhook.certificate.autoGenerate=false \
  --set webhook.certificate.certPem="${CERT_PEM}" \
  --set webhook.certificate.keyPem="${KEY_PEM}" \
  --set webhook.certificate.caBundle="${CA_BUNDLE}"
```

### CRD 验证失败

1. **检查 CRD 版本**
```bash
kubectl get crd drworkflows.dr.bkbcs.tencent.com -o yaml | grep "version:"
```

2. **重新安装 CRD**
```bash
kubectl replace -f config/crd/bases/dr.bkbcs.tencent.com_drworkflows.yaml
```

### 执行失败问题

1. **检查 DRWorkflow 状态**
```bash
kubectl get drworkflow <name> -o yaml
```

2. **检查 DRPlanExecution 状态和事件**
```bash
kubectl describe drplanexecution <name>
kubectl get events --field-selector involvedObject.name=<execution-name>
```

3. **启用 Debug 日志**
```bash
# 更新 Helm values
helm upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --set controller.logLevel=4 \
  --namespace bcs-system
```

## 🔒 安全考虑

### RBAC 最小权限原则

默认 ClusterRole 包含以下权限：
- 管理 DR CRDs (drworkflows, drplans, drplanexecutions)
- 创建/删除 Jobs
- 管理 Clusternet CRs (localizations, subscriptions)
- 发送 Events
- 读写 ConfigMaps 和 Secrets（用于 KubernetesResource 动作）

如需添加额外权限，使用 `rbac.additionalRules`：

```yaml
rbac:
  additionalRules:
  - apiGroups: ["custom.io"]
    resources: ["customresources"]
    verbs: ["get", "list", "create"]
```

### 网络策略

建议为 Controller 配置 NetworkPolicy：

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: bcs-drplan-controller
  namespace: bcs-system
spec:
  podSelector:
    matchLabels:
      control-plane: controller-manager
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 9443  # Webhook
    - protocol: TCP
      port: 8080  # Metrics
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 6443  # API Server
    - protocol: TCP
      port: 443   # HTTPS
```

## 📊 监控配置

### Prometheus ServiceMonitor

```bash
helm upgrade bcs-drplan-controller ./install/helm/bcs-drplan-controller \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.serviceMonitor.additionalLabels.prometheus=kube-prometheus \
  --namespace bcs-system
```

### 指标列表

- `controller_runtime_reconcile_total` - Reconcile 总次数
- `controller_runtime_reconcile_errors_total` - Reconcile 错误次数
- `controller_runtime_reconcile_time_seconds` - Reconcile 耗时
- 自定义指标（待实施）

## 💡 最佳实践

1. **使用 Leader Election**: 生产环境始终启用 `controller.leaderElection=true`
2. **配置资源限制**: 根据集群规模调整 CPU/Memory
3. **启用监控**: 集成 Prometheus ServiceMonitor
4. **日志级别**: 生产环境使用 `logLevel=0`，调试时使用 `logLevel=4`
5. **Webhook 证书**: 
   - 默认自动生成证书（有效期 10 年）
   - 生产环境建议使用企业 CA 签发的证书
   - 证书过期前使用 `helm upgrade` 重新部署
6. **备份 CRD**: 在升级前备份所有 DRWorkflow 和 DRPlan 定义

## 🔗 相关链接

- [用户指南](docs/user-guide.md)
- [快速开始](specs/001-drplan-action-executor/quickstart.md)
- [架构设计](specs/001-drplan-action-executor/spec.md)
- [Helm Chart 文档](install/helm/bcs-drplan-controller/README.md)
