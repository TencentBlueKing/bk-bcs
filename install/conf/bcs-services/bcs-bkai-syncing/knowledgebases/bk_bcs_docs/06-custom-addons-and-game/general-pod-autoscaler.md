# 通用 Pod 自动扩缩容（GeneralPodAutoscaler，GPA）指南

`GeneralPodAutoscaler`（简称 **GPA**，API Group: `autoscaling.bkbcs.tencent.com/v1alpha1`）是蓝鲸容器管理平台（BCS）基于 Kubernetes HPA v2 扩展开发的**通用型 Pod 自动弹性伸缩组件**，既适用于常规 Web/微服务，也深度适配游戏服务器（GameServer）及自定义 CRD 工作负载。

---

## 一、GPA 与原生 HPA 的核心差异与优势

1. **兼容 HPA 全量能力，且不强依赖 K8S 版本**：以独立 CRD + Operator 方式运行，升级灵活，无需重启 K8S 核心控制面组件。
2. **支持四大弹性驱动模式（可单独使用，也可多模式混合驱动）**：
   - **`MetricMode`（指标驱动）**：与原生 HPA 一致，支持内置资源指标（CPU、内存）及自定义指标（如 `memory_rss`、网卡流量、QPS）。
   - **`TimeMode`（Crontab 定时驱动）**：基于标准 Crontab 语法，在指定高峰时间段（如每晚 20:00~23:00）自动将副本数保障在预设水位。
   - **`WebhookMode`（Webhook 自定义策略驱动）**：定期调用业务自定义的 Webhook 接口获取期望副本数（如游戏服始终保持 N 个空闲对局房间 Buffer）。
   - **`EventMode`（外部事件源驱动）**：支持对接 `Kafka`、`Redis` 等外部队列深度触发弹性扩缩。
3. **YAML 结构差异**：
   - 原生 HPA 的 `metrics` 直接位于 `spec.metrics`。
   - GPA 将指标列表嵌套在 `spec.metric.metrics` 下，与 `spec.time`、`spec.webhook`、`spec.event` 并列。

---

## 二、部署与常用命令

- **组件安装**：在 BCS【插件管理】→【组件库】中启用 `bcs-general-pod-autoscaler`，或使用 Helm 安装：
  ```bash
  helm upgrade bcs-general-pod-autoscaler helm/bcs-general-pod-autoscaler -n bcs-system --install
  ```
- **查看 GPA 状态（资源简写为 `pa` 或 `gpa`）**：
  ```bash
  kubectl get pa -n <namespace>
  ```

---

## 三、四大伸缩模式配置示例

### 1. Metric 模式（内置 CPU/内存指标 + 自定义 Pod 指标）

```yaml
apiVersion: autoscaling.bkbcs.tencent.com/v1alpha1
kind: GeneralPodAutoscaler
metadata:
  name: gpa-metric-demo
  namespace: default
spec:
  minReplicas: 2
  maxReplicas: 10
  scaleTargetRef:
    apiVersion: tkex.tencent.com/v1alpha1
    kind: GameDeployment
    name: test-gamedeployment
  metric:
    metrics:
      # 内置 CPU 指标
      - type: Resource
        resource:
          name: cpu
          target:
            type: Utilization
            averageUtilization: 75
      # 自定义 Pod 级别指标（如 memory_rss）
      - type: Pods
        pods:
          metric:
            name: memory_rss
          target:
            type: AverageValue
            averageValue: 500Mi
```

### 2. Time 模式（Crontab 定时扩缩容）

基于 Crontab 表达式定义多个时间窗口的期望副本数（当多模式混合触发时，GPA 会取各模式计算出的最大推荐副本数以保障业务容量）：

```yaml
apiVersion: autoscaling.bkbcs.tencent.com/v1alpha1
kind: GeneralPodAutoscaler
metadata:
  name: gpa-crontab-demo
  namespace: default
spec:
  minReplicas: 2
  maxReplicas: 20
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-service
  time:
    ranges:
      # 每天 12:00 ~ 13:59 午间高峰期维持至少 8 个副本
      - schedule: "*/1 12-13 * * *"
        desiredReplicas: 8
      # 每天 20:00 ~ 22:59 晚间高峰期维持至少 15 个副本
      - schedule: "*/1 20-22 * * *"
        desiredReplicas: 15
```

### 3. Webhook 模式（业务自定义伸缩决策，支持与 Crontab 混合）

业务可提供集群内 Service 或集群外 HTTP URL，GPA 会携带当前工作负载的 `name`、`namespace`、`currentReplicas` 及自定义 `parameters` 发送 `AutoscaleReview` 请求，由业务服务返回 `{"response": {"uid": "...", "scale": true, "replicas": 6}}`：

```yaml
apiVersion: autoscaling.bkbcs.tencent.com/v1alpha1
kind: GeneralPodAutoscaler
metadata:
  name: gpa-webhook-mix-demo
  namespace: default
spec:
  minReplicas: 2
  maxReplicas: 20
  scaleTargetRef:
    apiVersion: tkex.tencent.com/v1alpha1
    kind: GameDeployment
    name: test-gamedeployment
  # 可同时结合定时模式与 Webhook 模式
  time:
    ranges:
      - schedule: "*/1 18-23 * * *"
        desiredReplicas: 6
  webhook:
    parameters:
      buffer: "2"
    # 集群内 Service 方式（若在集群外可使用 url: "http://example.com:8080/scale"）
    service:
      namespace: kube-system
      name: gpa-webhook-svc
      port: 8000
      path: scale
```

---

## 四、精细控制扩缩容速率与防抖动（`behavior`）

为防止流量瞬时抖动导致 Pod 频繁扩缩容，可通过 `spec.behavior` 配置扩容（`scaleUp`）与缩容（`scaleDown`）的冷却窗口及步长策略（语法与原生 HPA v2 完全一致）：

```yaml
spec:
  minReplicas: 2
  maxReplicas: 20
  behavior:
    scaleDown:
      # 缩容稳定窗口（默认 300 秒，取过去 300 秒内的最高推荐副本数）
      stabilizationWindowSeconds: 300
      selectPolicy: Max
      policies:
        # 每 60 秒最多只允许缩容 1 个 Pod（或使用 type: Percent 按百分比限制）
        - type: Pods
          value: 1
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
        - type: Percent
          value: 100
          periodSeconds: 60
```
