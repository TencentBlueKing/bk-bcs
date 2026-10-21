# 蓝鲸容器监控与 PromQL 查询指南

蓝鲸容器管理平台（BCS）与**蓝鲸监控平台（BKMonitor）**深度集成，提供集群、节点、命名空间、工作负载、Pod、容器及自定义业务 Metrics 的全栈可观测性能力。同时，BCS 网关开放了兼容 Prometheus 标准协议的查询接口（`bcs-monitor`），支持使用 PromQL 实时查询容器监控指标。

---

## 一、容器监控架构与核心依赖组件

在蓝鲸 7.x 体系中，被纳管集群要正常上报容器监控指标与事件，必须确保以下三个基础组件正常运行：

| 组件名称 | 部署位置 / 形态 | 核心作用 |
| :--- | :--- | :--- |
| **`bcs-k8s-watch`** | 集群 `bcs-system` 命名空间（Deployment） | 监听采集集群内 K8S 资源元数据并同步至 `bcs-storage`，为监控平台提供集群拓扑上下文。 |
| **`bkmonitor-operator-stack`** | 集群 `bkmonitor-operator` 命名空间 | 包含 `bkmonitor-operator`（管理 `ServiceMonitor` / `PodMonitor` / `DataID` CRD）、`bkmonitorbeat`（DaemonSet 指标采集器）、`kube-state-metrics`、`node-exporter` 和 `kubernetes-event-exporter`。 |
| **数据上报管道** | 节点级 `gse_agent` 或跨云区域 `bk-collector` | 直连环境依赖节点上的 `gse_agent`（可通过 `ps -ef \| grep gse_agent` 检查）；跨云区域无 Agent 环境可开启 `bkmonitorbeat` 的 `pushMode` 推送至 Proxy 的 `bk-collector`。 |

---

## 二、自定义业务指标采集（Metrics 管理）

若业务容器自身暴露了 Prometheus 格式的 `/metrics` 接口，可通过以下两种方式接入蓝鲸容器监控：

1. **在 BCS 控制台配置【Metrics 管理】**：
   - 先为业务工作负载创建关联的 `Service`。
   - 进入【插件管理】→【Metrics 管理】（或 ServiceMonitor 配置），点击【新建】，选择目标集群、命名空间、关联的 `Service`、端口名称、采集路径（如 `/metrics`）及采集周期（如 `30s`）。
2. **直接创建 `ServiceMonitor` 或 `PodMonitor` CRD**：
   - `bkmonitor-operator` 会自动监听集群内的 `ServiceMonitor` 与 `PodMonitor` 资源，并动态下发采集子任务给 `bkmonitorbeat`（下发后约 3~10 分钟即可在蓝鲸监控平台检索到指标）。

---

## 三、使用 PromQL 查询 BCS 监控指标

BCS 提供了兼容 Prometheus HTTP API 的查询接口（如即时查询 `query`、范围查询 `query_range`、时间序列 `series`、标签查询 `labels` 与 `label_values`）。

### 1. 必填标签约束（`cluster_id`）
在通过 BCS 监控接口执行 PromQL 查询（`query` / `query_range`）或时间序列查询（`series`）时，**PromQL 表达式的标签选择器中必须包含 `cluster_id="BCS-K8S-xxxxx"`**（或通过接口路径限定集群），以完成集群数据路由与权限校验。

### 2. 常用容器与节点 PromQL 查询示例

假设目标集群 ID 为 `BCS-K8S-40001`，命名空间为 `default`：

#### （1）查询命名空间下各 Pod 的 CPU 使用量（核数）
```promql
sum by (pod_name) (rate(container_cpu_usage_seconds_total{cluster_id="BCS-K8S-40001", namespace="default", container_name!="POD", container_name!=""}[5m]))
```

#### （2）查询命名空间下各 Pod 的内存使用量（Working Set 字节数）
```promql
sum by (pod_name) (container_memory_working_set_bytes{cluster_id="BCS-K8S-40001", namespace="default", container_name!="POD", container_name!=""})
```

#### （3）查询容器重启次数（排查 CrashLoopBackOff / OOMKilled）
```promql
sum by (namespace, pod_name, container_name) (kube_pod_container_status_restarts_total{cluster_id="BCS-K8S-40001", namespace="default"})
```

#### （4）查询集群节点的 CPU 与内存实际使用率
```promql
# 节点内存使用率
(1 - (node_memory_MemAvailable_bytes{cluster_id="BCS-K8S-40001"} / node_memory_MemTotal_bytes{cluster_id="BCS-K8S-40001"})) * 100
```

---

## 四、容器监控无数据排查清单

当在 BCS 控制台或蓝鲸监控页面发现某个集群无监控数据时，请按以下顺序逐步排查：

```bash
# 1. 检查 bcs-k8s-watch 是否 Running 且无报错
kubectl get pods -n bcs-system | grep bcs-k8s-watch
kubectl logs -n bcs-system -l app.kubernetes.io/name=bcs-k8s-watch --tail=100

# 2. 检查 bkmonitor-operator 命名空间下所有组件（operator、bkmonitorbeat、kube-state-metrics、node-exporter）是否全部 Running
kubectl get pods -n bkmonitor-operator -o wide

# 3. 检查 DataID 资源是否已由蓝鲸监控平台正常分配并同步到集群
kubectl get dataids -n bkmonitor-operator

# 4. 登录其中一个 bkmonitorbeat Pod，检查采集子任务配置是否已成功生成
kubectl exec <bkmonitorbeat-pod-name> -n bkmonitor-operator -- ls -la /data/bkmonitorbeat/config/child_configs

# 5. 检查节点上的 gse_agent 进程与通道状态（直连模式下）
ps -ef | grep gse_agent
```
> **提示**：首次安装 `bkmonitor-operator-stack` 后，由于蓝鲸监控后台分配 DataID 与轮询同步存在周期，预计首次看到监控数据需要等待 **10~20 分钟**。
