# GameDeployment 增强型无状态工作负载使用指南

`GameDeployment`（API Group: `tkex.tencent.com/v1alpha1`）是蓝鲸容器管理平台（BCS）针对游戏服务器（GameServer）及复杂业务场景自研的**增强版无状态工作负载**。它在完全兼容 Kubernetes 原生 `Deployment` 语义的基础上，直接管理 Pod 生命周期（不通过多层 ReplicaSet 跳转），提供了容器原地更新、镜像热更新、分步自动化灰度、优雅删除/更新 Hook、指定 Pod 缩容及 Pod 唯一序号注入等核心特性。

---

## 一、核心特性对比（GameDeployment vs 原生 Deployment）

| 能力维度 | 原生 `Deployment` | BCS `GameDeployment` |
| :--- | :--- | :--- |
| **滚动更新 (`RollingUpdate`)** | 支持（支持 `maxSurge` / `maxUnavailable`） | 支持（支持 `maxSurge` / `maxUnavailable`） |
| **原地重启更新 (`InplaceUpdate`)** | 不支持（更新镜像必须销毁重建 Pod，IP 与共享内存丢失） | **支持**（仅重启容器，保持 Pod IP、所在节点及 IPC 共享内存不变，支持 `gracePeriodSeconds` 平滑摘流与容器更新顺序策略） |
| **镜像热更新 (`HotPatchUpdate`)** | 不支持 | **支持**（配合定制 kubelet/dockerd，不重启容器仅替换镜像版本） |
| **`partition` 灰度发布** | 不支持（需手动维护多个 Deployment） | **支持**（通过数字或百分比精确控制保留旧版本 Pod 数量） |
| **分步骤自动化灰度 (`canary`)** | 不支持 | **支持**（支持按步骤灰度、定时暂停、手动暂停及外部 Hook 指标校验） |
| **删除/更新前优雅 Hook** | 仅支持简单的容器内 `preStop` | **支持**（联动 `bcs-hook-operator`，支持 `preDelete` / `preInplace` / `postInplace` 多阶段 Webhook/Prometheus/K8S 校验） |
| **指定 Pod 缩容 (`podsToDelete`)** | 不支持 | **支持**（在 `scaleStrategy.podsToDelete` 中指定优先删除的 Pod 名称或按 Cost 权重排序删除） |
| **Pod 注入唯一序号 (`POD_INDEX`)** | 不支持 | **支持**（通过注解开启，自动向容器环境变量注入指定区间内的唯一序号） |
| **CRD / CR 防误删保护** | 无 | **支持**（通过 Admission Webhook 与 `io.tencent.bcs.dev/deletion-allow` 标签防止误删） |

---

## 二、组件安装

在使用 `GameDeployment` 前，请确保目标集群已安装 `bcs-gamedeployment-operator`（若需使用 Hook 能力，还需安装 `bcs-hook-operator`）：
- **控制台安装**：进入【插件管理】→【组件库】→ 选择目标集群，启用 `bcs-gamedeployment-operator` 与 `bcs-hook-operator`。
- **Helm 命令行安装**：
  ```bash
  helm upgrade bcs-gamedeployment-operator helm/bcs-gamedeployment-operator -n bcs-system --install
  ```

---

## 三、典型 CRD 配置样例

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameDeployment
metadata:
  name: test-gamedeployment
  namespace: default
  labels:
    app: test-gamedeployment
    # 防误删标签：Always 允许直接删除；Cascading 无级联 Pod 时允许删除；不填则禁止删除
    io.tencent.bcs.dev/deletion-allow: Always
  annotations:
    # 开启 Pod 唯一序号注入（通过环境变量 POD_INDEX 读取）
    tkex.bkbcs.tencent.com/gamedeployment-index-on: "true"
    tkex.bkbcs.tencent.com/gamedeployment-index-range: '{"podStartIndex": 100, "podEndIndex": 110}'
spec:
  replicas: 5
  minReadySeconds: 5
  disableReadinessGate: false
  selector:
    matchLabels:
      app: test-gamedeployment
  # 指定优先删除的 Pod 列表（可选）
  scaleStrategy:
    podsToDelete:
      - test-gamedeployment-ghrwg
  # 更新策略配置
  updateStrategy:
    # 可选值：RollingUpdate | InplaceUpdate | HotPatchUpdate
    type: InplaceUpdate
    # 灰度保留旧版本实例数量或百分比（0 表示全量更新）
    partition: 0
    maxUnavailable: 25%
    maxSurge: 0
    paused: false
    inPlaceUpdateStrategy:
      # 原地更新前提前将 Pod 设为 UnReady 以摘除 Service 流量的缓冲秒数
      gracePeriodSeconds: 30
      # 多容器原地更新顺序策略：DisOrdered（默认） | OrderedReady | OrderedUpdated
      policy: DisOrdered
  # 优雅删除与原地更新前后 Hook（需配合 HookTemplate 使用）
  preDeleteUpdateStrategy:
    hook:
      templateName: test-delete-hook
  preInplaceUpdateStrategy:
    hook:
      templateName: test-pre-inplace-hook
  postInplaceUpdateStrategy:
    hook:
      templateName: test-post-inplace-hook
  template:
    metadata:
      labels:
        app: test-gamedeployment
      annotations:
        # 可选：自定义缩容刪除优先级代价值计算指标与权重
        io.tencent.bcs.dev/pod-cost-metrics-name: "cpu;memory"
        io.tencent.bcs.dev/pod-cost-metrics-weight: "2;0.001"
        io.tencent.bcs.dev/pod-deletion-cost-sort-method: "ascend"
    spec:
      containers:
        - name: gameserver
          image: hub.bktencent.com/blueking/game-demo:0.0.1
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 8080
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
```

---

## 四、核心特性详解与操作指引

### 1. 滚动更新（`RollingUpdate`）与原地更新（`InplaceUpdate`）

- **`RollingUpdate`**：与 K8S 原生滚动更新一致，销毁旧版本 Pod 并创建新版本 Pod。可配合 `maxUnavailable`（每批最大不可用数，默认 `25%`）和 `maxSurge`（更新前最多额外扩建的新 Pod 数，默认 `0` 或 `25%`）保障更新期间的服务容量。
- **`InplaceUpdate`（核心推荐）**：
  - 更新时**不销毁 Pod、不重新调度**，仅原地重启镜像发生变更的容器，Pod IP 和共享内存保持不变。
  - **字段限制**：`InplaceUpdate` 仅支持以原地重启容器的方式更新 `image` 字段，以及以不重启的方式更新 `labels` / `annotations` 字段；若修改了容器端口、环境变量或资源配额等其他字段，原地更新不会生效（需切换为 `RollingUpdate`）。
  - **流量无损切换（`gracePeriodSeconds`）**：设置 `inPlaceUpdateStrategy.gracePeriodSeconds: 30` 后，Operator 在重启容器前会先通过 `InPlaceUpdateReady` ReadinessGate 将 Pod 状态置为 `UnReady`，等待 30 秒让 K8S Service / CLB 将该 Pod 从后端 Endpoints 剔除，随后再重启容器；容器恢复正常后再将 Pod 重新置为 `Ready` 接回流量。
  - **多容器更新顺序（`inPlaceUpdateStrategy.policy`）**：
    - `DisOrdered`（默认）：不保证同一 Pod 内多个容器的更新顺序。
    - `OrderedReady`：按 `spec.containers` 定义顺序依次更新，上一个容器更新且 `Ready` 后才更新下一个容器。
    - `OrderedUpdated`：按定义顺序更新，上一个容器完成重启后即开始更新下一个容器。

### 2. 使用 `partition` 进行灰度发布

`spec.updateStrategy.partition` 表示**保留旧版本 Pod 的数量（整数）或百分比（如 `40%`）**，默认为 `0`：
- 若 `replicas: 8`，设置 `partition: 3`，则修改镜像后控制器只会将 `8 - 3 = 5` 个 Pod 更新到新版本，其余 3 个保留旧版本。
- 验证新版本正常后，将 `partition` 改为 `0`，即可完成剩余全部实例的更新。

### 3. 指定 Pod 缩容与缩容优先级排序

- **精确指定删除某个 Pod（`scaleStrategy.podsToDelete`）**：
  - 将目标 Pod 名称填入 `spec.scaleStrategy.podsToDelete` 列表并同时调小 `replicas`，控制器会优先删除列表中指定的 Pod（删除后自动从列表中移除该名称）。
  - **注意**：若仅把 Pod 名字加入 `podsToDelete` 但未调小 `replicas`，控制器会删掉该指定 Pod 并立即新建一个替代 Pod（相当于单实例定向重建）。
- **默认缩容排序与 Cost 权重**：
  - 未指定 `podsToDelete` 时，控制器按 `NotReady < Ready`、`Unscheduled < Scheduled`、`Pending < Running` 顺序优先淘汰非健康实例，并支持结合 Pod 注解 `pod-cost-metrics-name` / `pod-cost-metrics-weight` 按业务负载高低排序删除。

### 4. Pod 注入唯一序号（`POD_INDEX`）

无状态 GameDeployment 的 Pod 名称后缀是随机字符串，若业务进程需要固定区间内的唯一数字编号（如 `100`~`110`），可在 `metadata.annotations` 中添加：

```yaml
metadata:
  annotations:
    tkex.bkbcs.tencent.com/gamedeployment-index-on: "true"
    tkex.bkbcs.tencent.com/gamedeployment-index-range: '{"podStartIndex": 100, "podEndIndex": 110}'
```
开启后，每个 Pod 启动时会自动注入环境变量 `POD_INDEX`（取值为 `100`~`110` 之间未占用的唯一整数）。

### 5. 防误删保护（`deletion-allow`）

为防止误删 CRD 或生产环境 `GameDeployment` 实例导致所有游戏服 Pod 被级联销毁，Admission Webhook 默认拦截未携带许可标签的删除请求：

| 标签 Key | 标签 Value | 删除行为效果 |
| :--- | :--- | :--- |
| `io.tencent.bcs.dev/deletion-allow` | `Always` | 允许直接删除该 GameDeployment 实例 |
| `io.tencent.bcs.dev/deletion-allow` | `Cascading` | **安全推荐**：仅当实例下已无级联 Pod（即已先缩容 `replicas: 0`）时才允许删除 |
| 未配置该标签 | - | **默认拒绝删除**，防止误操作 |

### 6. ReadinessGate 可选开关（`disableReadinessGate`）

在 BCS 1.27.0+ 版本中，可通过 `spec.disableReadinessGate: true` 关闭默认注入的 `InPlaceUpdateReady` ReadinessGate（默认为 `false` 开启）。
> **注意**：修改 `disableReadinessGate` 涉及 Pod ReadinessGate 结构变更，不能在 `InplaceUpdate` 模式下直接生效，需先以 `RollingUpdate` 重建 Pod 后再切回 `InplaceUpdate`。

---

## 五、常用运维命令

```bash
# 查看集群中的 GameDeployment 列表及更新进度
kubectl get gamedeployments.tkex.tencent.com -n <namespace>

# 修改镜像触发原地更新或滚动更新
kubectl patch gamedeployment <name> -n <namespace> --type='json' \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"<new_image>"}]'

# 扩缩容 GameDeployment
kubectl scale gamedeployment/<name> --replicas=10 -n <namespace>

# 暂停或恢复发布（paused: true / false）
kubectl patch gamedeployment <name> -n <namespace> --type='json' \
  -p='[{"op": "replace", "path": "/spec/updateStrategy/paused", "value": true}]'
```
