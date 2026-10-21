# 自动扩缩容（HPA）与 RBAC 管理指南

蓝鲸容器管理平台（BCS）在【资源视图】中提供了 **HPA（Horizontal Pod Autoscaler，水平 Pod 自动扩缩容）** 与 **RBAC（ServiceAccounts，服务账号）** 的可视化管理。

---

## 一、水平 Pod 自动扩缩容（HPA）

HPA 能根据监控到的 CPU 利用率、内存使用率或自定义指标，自动动态调整目标工作负载（如 `Deployment`、`StatefulSet`、`GameDeployment`、`GameStatefulSet`）的 Pod 副本数量。

> **扩展提示**：若除了指标触发外，您还需要按 **Crontab 定时扩缩容**、**Webhook 自定义业务策略扩缩容**（如游戏房间余量缓冲）或 **Kafka/Redis 事件驱动扩缩容**，请使用 BCS 自研的增强组件 [GeneralPodAutoscaler（GPA）](../06-custom-addons-and-game/general-pod-autoscaler.md)。

### 1. HPA 生效前提条件
1. 集群内已部署 `metrics-server` 组件（可通过 `kubectl top pod` 验证能否正常获取 Pod CPU/内存指标）。
2. **关键要求**：目标工作负载的容器配置中**必须明确声明 `resources.requests`（如 `requests.cpu` / `requests.memory`）**。若 Pod 未设置 `requests`，HPA 将无法计算资源使用率百分比（显示 `<unknown>/80%`），导致自动扩缩容无法生效。

### 2. 创建与配置 HPA
进入【资源视图】→【HPA】，选择目标集群与命名空间，点击【创建】：

#### 示例 1：针对原生 Deployment 的 CPU 与内存双指标 HPA（`autoscaling/v2`）

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: nginx-demo-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx-demo
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

#### 示例 2：针对 BCS 游戏负载 `GameDeployment` 的 HPA

HPA 同样支持对 BCS 扩展工作负载进行自动扩缩容，并且缩容时会触发 `GameDeployment` 配置的 `preDeleteUpdateStrategy` 钩子以实现优雅缩容（避免强杀在线对局）：

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gameserver-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: tkex.tencent.com/v1alpha1
    kind: GameDeployment
    name: test-gamedeployment
  minReplicas: 2
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 75
```

---

## 二、RBAC 服务账号管理（ServiceAccounts）

在 Kubernetes 中，**ServiceAccount（服务账号）** 为 Pod 内运行的进程提供访问集群 Kubernetes API Server 的身份标识。

### 1. 管理 ServiceAccount
进入【资源视图】→【RBAC】→【ServiceAccounts】，可查看、创建或删除命名空间下的服务账号：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-custom-sa
  namespace: default
imagePullSecrets:
  - name: my-registry-secret
```

### 2. ServiceAccount 常见用途
1. **Pod 内部调用 K8S API 授权**：
   - 为需要访问 K8S API 的应用（如自研 Operator、服务发现组件、Spring Cloud Kubernetes）创建专属 `ServiceAccount`。
   - 通过 `RoleBinding`（命名空间级）或 `ClusterRoleBinding`（集群级）将该 `ServiceAccount` 与最小权限的 `Role` / `ClusterRole` 绑定。
   - 在工作负载的 `spec.template.spec.serviceAccountName` 中指定该 `ServiceAccount`。
2. **统一绑定私有镜像拉取凭据（`imagePullSecrets`）**：
   - 将私有镜像仓库的 Secret 绑定到命名空间的 `default`（或自定义）`ServiceAccount` 的 `imagePullSecrets` 字段中，该命名空间下使用此 ServiceAccount 的所有 Pod 将自动获得镜像拉取凭据，无需在每个 Deployment 中重复编写 `imagePullSecrets`。
