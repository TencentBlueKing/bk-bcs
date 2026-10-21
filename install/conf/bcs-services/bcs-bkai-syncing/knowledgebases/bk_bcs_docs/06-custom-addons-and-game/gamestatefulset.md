# GameStatefulSet 增强型有状态工作负载使用指南

`GameStatefulSet`（API Group: `tkex.tencent.com/v1alpha1`）是蓝鲸容器管理平台（BCS）针对有状态游戏服务器及复杂有状态应用自研的**增强版 StatefulSet**。它在 100% 兼容 Kubernetes 原生 `StatefulSet`（固定序号 Pod 名称 `<name>-0, 1, 2...`、固定网络标识、独立 PVC 存储模板）的基础上，新增了**容器原地重启升级（InplaceUpdate）**、**镜像热更新（HotPatchUpdate）**、**自动并行滚动/原地更新（Parallel + maxUnavailable/maxSurge）**、**失联节点 Pod 自动强制漂移**、**分步自动化灰度（Canary）**以及**删除/更新前优雅 Hook** 等关键生产特性。

---

## 一、核心增强特性一览

1. **完全兼容原生 StatefulSet**：保留有序固定 Pod 名称、Headless Service 稳定域名及 `volumeClaimTemplates` 持久化卷绑定。
2. **四种更新策略（`updateStrategy.type`）**：
   - `RollingUpdate`：滚动重建更新。
   - `OnDelete`：手动删除 Pod 时才触发更新。
   - **`InplaceUpdate`（新增核心）**：原地重启容器更新镜像，不销毁 Pod、不改变 Pod IP、不卸载数据卷、不丢失共享内存，支持 `gracePeriodSeconds` 无损摘流与多容器顺序策略。
   - **`HotPatchUpdate`（新增）**：配合定制节点运行时，在不重启容器的情况下热替换容器镜像版本。
3. **自动并行更新（`podManagementPolicy: Parallel` 增强）**：
   - 原生 StatefulSet 设置 `Parallel` 时仅并行创建/删除 Pod，更新时依然只能按序逐个更新。
   - `GameStatefulSet` 在 `podManagementPolicy: Parallel` 下**支持并行更新**，并可结合 `rollingUpdate.maxUnavailable` 和 `rollingUpdate.maxSurge` 控制每批并发更新的窗口大小，大幅缩短百级副本有状态服务的发版耗时。
4. **Node 失联时 Pod 自动强制漂移（NodeLost Force Delete）**：
   - 原生 StatefulSet 在节点宕机失联（`NodeLost`）时，被驱逐的 Pod 会一直卡在 `Terminating` 状态，导致新节点上无法重建同名序号 Pod。
   - `GameStatefulSet` 支持通过注解开启失联节点 `Terminating` Pod 自动强制清理，促使业务 Pod 在健康节点上快速重建恢复。
5. **发布生命周期 Hook 与分步灰度**：
   - 联动 `bcs-hook-operator`，支持 `preDeleteUpdateStrategy`、`preInplaceUpdateStrategy`、`postInplaceUpdateStrategy` 及 `canary` 分步灰度发布。
6. **防误删保护**：支持 `io.tencent.bcs.dev/deletion-allow` 标签拦截误删操作。

---

## 二、典型 CRD 配置样例

```yaml
apiVersion: tkex.tencent.com/v1alpha1
kind: GameStatefulSet
metadata:
  name: test-gamestatefulset
  namespace: default
  labels:
    app: test-gamestatefulset
    # 防误删保护：Always | Cascading
    io.tencent.bcs.dev/deletion-allow: Always
spec:
  # 必填：关联的 Headless Service 名称
  serviceName: "test-headless-svc"
  replicas: 5
  # OrderedReady（默认，按序） | Parallel（并行创建/删除/更新）
  podManagementPolicy: Parallel
  selector:
    matchLabels:
      app: test-gamestatefulset
  updateStrategy:
    # RollingUpdate | OnDelete | InplaceUpdate | HotPatchUpdate
    type: InplaceUpdate
    paused: false
    rollingUpdate:
      # 保留旧版本 Pod 的数量或百分比（序号 >= partition 的 Pod 会被更新）
      partition: 0
      # 仅在 podManagementPolicy: Parallel 时生效
      maxUnavailable: 25%
      maxSurge: 0
    inPlaceUpdateStrategy:
      gracePeriodSeconds: 30
      policy: DisOrdered
  # 优雅删除与原地更新前后 Hook
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
        app: test-gamestatefulset
      annotations:
        # 开启节点失联 (NodeLost) 时自动强制删除 Terminating Pod 以快速漂移重建
        pod.gamestatefulset.bkbcs.tencent.com/node-lost-force-delete: "true"
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
```

---

## 三、关键配置项详解与注意事项

### 1. 原地更新（`InplaceUpdate`）注意事项
- **适用修改范围**：`InplaceUpdate` 仅支持以原地重启容器的方式更新 `image` 字段，以及以不重启的方式更新 `labels` / `annotations` 字段。
- **重要警告**：若在 `InplaceUpdate` 模式下修改了除 `image`、`labels`、`annotations` 以外的其他 Pod Template 字段（如环境变量、端口、挂载卷、资源配额），**此次原地更新会卡住**！如需修改其他字段，请将 `updateStrategy.type` 切换为 `RollingUpdate`。
- **流量平滑切换（`gracePeriodSeconds`）**：设置 `inPlaceUpdateStrategy.gracePeriodSeconds: 30` 后，控制器在原地重启容器前会先将 Pod 置为 `UnReady` 等待 30 秒，待 Service Endpoints 摘除流量后再重启容器。

### 2. 灰度发布（`rollingUpdate.partition`）
在 `GameStatefulSet` 中，无论是 `RollingUpdate`、`InplaceUpdate` 还是 `HotPatchUpdate`，灰度控制统一使用 `spec.updateStrategy.rollingUpdate.partition` 字段：
- 当 `partition` 为整数 `k` 时，只有**序号大于或等于 `k`**（即 `<name>-k` 到 `<name>-(replicas-1)`）的 Pod 会被更新为新版本，序号小于 `k` 的 Pod 保持旧版本不变。
- 也支持设置百分比，分批调小 `partition` 直至 `0` 即可完成全量发布。

### 3. 并行更新与批量控制（`Parallel` + `maxUnavailable` / `maxSurge`）
- `maxUnavailable`（默认 `25%`）和 `maxSurge`（默认 `0`）**仅在 `podManagementPolicy: Parallel` 且执行更新时生效**。
- 例如共有 8 个实例，设置 `podManagementPolicy: Parallel` 且 `maxUnavailable: 2`，则每次会同时并行滚动或原地更新 2 个实例，大幅提升发布速度。

### 4. 节点失联自动漂移（`node-lost-force-delete`）
在 `spec.template.metadata.annotations` 中添加：
```yaml
pod.gamestatefulset.bkbcs.tencent.com/node-lost-force-delete: "true"
```
当某个工作节点突发断电或网络失联进入 `NodeLost` 状态，且该节点上的 `GameStatefulSet` Pod 被驱逐进入 `Terminating` 状态时，Operator 会自动从 Etcd 中强制删除该 `Terminating` Pod，从而使其立即在集群其他健康节点上重建，大幅缩短单点硬件故障的业务恢复时间（MTTR）。

### 5. 防误删保护（`deletion-allow`）
与 `GameDeployment` 一致，默认禁止直接删除未带标签的 `GameStatefulSet`。删除前建议先缩容副本数为 `0` 并打上 `io.tencent.bcs.dev/deletion-allow: Cascading`（或确认强制删除时打上 `Always`）。

---

## 四、常用运维命令

```bash
# 查看集群中的 GameStatefulSet 状态
kubectl get gamestatefulsets.tkex.tencent.com -n <namespace>

# 扩缩容 GameStatefulSet
kubectl scale gamestatefulset/<name> --replicas=6 -n <namespace>

# 修改镜像版本触发更新
kubectl patch gamestatefulset <name> -n <namespace> --type='json' \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/image", "value":"<new_image>"}]'
```
