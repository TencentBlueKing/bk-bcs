# 容器日志采集指南

由于容器生命周期短暂，容器销毁重建后其内部日志会随之丢失。蓝鲸容器管理平台（BCS）打通了**蓝鲸日志平台（BKLog）**，提供对业务容器**零侵入**的**标准输出日志（Stdout/Stderr）**与**容器内文本文件日志**采集、清洗、检索和告警能力。

---

## 一、容器日志分类与采集原理

1. **标准输出日志（Stdout / Stderr）**：
   - 容器内进程直接打印到标准输出和标准错误的日志，由容器运行时（Docker / Containerd）以落盘文件形式保存在宿主机目录（如 `/var/lib/docker/containers/<id>/<id>-json.log` 或 `/var/log/pods/...`）。
2. **容器内文本文件日志（非标准输出）**：
   - 业务进程写入容器内部指定路径的日志文件（例如 `/data/home/logs/app.log` 或 `/var/log/nginx/access.log`），采集器可在宿主机上通过 `/proc/<pid>/root/<容器内路径>` 自动定位并实时采集。
3. **自动打标与采集方式**：
   - 集群节点以 DaemonSet 方式运行日志采集器（`bkunifylogbeat` / `logbeat`），采集时自动附加 `cluster_id`、`namespace`、`pod_name`、`container_name`、Pod Labels 及自定义 Tags，上报至蓝鲸日志平台。

---

## 二、配置日志采集的方式

### 方式 1：通过 BCS 控制台【日志采集】可视化创建（推荐）

1. 进入顶部导航栏【插件管理】→【日志采集】，选择目标集群。
2. 点击【新建采集规则】，配置以下信息：
   - **规则名称**与**所属命名空间**（支持全命名空间或指定命名空间）。
   - **关联工作负载 / 选择器**：可按工作负载类型（如 `Deployment`、`StatefulSet`、`GameDeployment`）和名称匹配，或按 **Pod Label Selector** 匹配目标 Pod，也可指定容器名（`containerName`）。
   - **日志源类型**：
     - **标准输出**：直接勾选开启标准输出采集。
     - **文件日志**：填写容器内的日志绝对路径（支持通配符，如 `/data/logs/*.log`）及文件编码格式（如 `UTF-8`）。
   - **附加日志标签（Log Tags）**：自定义键值标签，或勾选自动上报 Pod Labels。
3. 保存后，可在规则列表点击【前往日志检索】一键跳转到**蓝鲸日志平台**实时查询和分析采集到的容器日志。

---

### 方式 2：通过 `BcsLogConfig` CRD 声明式管理

在启用了 `bcs-webhook-server` 与日志控制器的集群中，也支持直接通过 `BcsLogConfig`（API Group: `bkbcs.tencent.com/v1`）声明采集规则：

#### 1. 集群默认日志采集配置（`configType: default`）

作用于命名空间下未匹配特定自定义规则的所有普通业务容器：

```yaml
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: default-log-conf
  namespace: default
spec:
  configType: default
  appId: "20000"
  clusterId: BCS-K8S-15049
  stdDataId: "20001"
  nonStdDataId: "20002"
  # 采集容器标准输出
  stdout: true
  # 采集容器内指定路径下的文本日志
  logPaths:
    - /data/home/logs
    - /data/logs/default
  logTags:
    platform: bcs
```

#### 2. 针对特定工作负载或标签的自定义采集（`configType: custom`）

```yaml
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: custom-app-log-conf
  namespace: default
spec:
  configType: custom
  clusterId: BCS-K8S-15049
  appId: "20000"
  workloadType: Deployment
  # 支持正则表达式匹配工作负载名称
  workloadName: "python-web.*"
  workloadNamespace: default
  stdout: true
  stdDataId: "20001"
  nonStdDataId: "20002"
  logPaths:
    - /data/home/logs/access.log
  podLabels: true
  logTags:
    app: python-web
  # 可选：通过 Label Selector 进一步精确过滤目标 Pod
  selector:
    matchLabels:
      env: prod
    matchExpressions:
      - key: tier
        operator: In
        values: ["backend", "api"]
```

#### 3. 针对 Pod 内特定容器（如 Sidecar）精细化配置（`containerConfs`）

```yaml
apiVersion: bkbcs.tencent.com/v1
kind: BcsLogConfig
metadata:
  name: sidecar-log-conf
  namespace: default
spec:
  configType: custom
  clusterId: BCS-K8S-15049
  appId: "20000"
  containerConfs:
    - containerName: istio-proxy
      stdout: true
      stdDataId: "1001"
      nonStdDataId: "1002"
      logPaths:
        - /var/log/istio
      logTags:
        component: sidecar
```

#### 4. 系统命名空间组件日志采集说明（`configType: bcs-system`）
默认情况下，系统命名空间（`kube-system`、`kube-public`、`bcs-system`）下的 Pod 不会注入默认业务日志采集规则。若需采集系统组件日志，需创建 `configType: bcs-system` 的 `BcsLogConfig`，并在目标系统组件 Pod 的 `metadata.annotations` 中添加 `webhook.inject.bkbcs.tencent.com: "true"`。
