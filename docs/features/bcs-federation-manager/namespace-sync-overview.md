# bcs-federation-manager namespace 同步方案说明

## 背景

`bcs-federation-manager` 中的 `syncnamespacequota.go` 负责按周期扫描联邦 host 集群中的 namespace，并将目标 namespace 同步到对应的子集群。虽然代码命名中包含 `quota`，但从当前这条链路来看，核心行为主要是：

1. 周期性扫描 host 集群中的 namespace。
2. 根据 namespace 上声明的 `cluster-range` 计算目标子集群范围。
3. 为每个目标子集群派发异步 task。
4. 由 task 在子集群中创建或更新 namespace。
5. 回写 host 集群中的 namespace 状态和 task 信息。

本文档描述的是当前代码实现的整体链路、边界条件，以及从代码审查视角识别出的潜在问题。

## 相关代码

- 控制器入口：`bcs-services/bcs-federation-manager/handler/syncnamespacequota.go`
- 普通子集群 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/check_in_normal_step.go`
- 混部子集群 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/check_in_hunbu_step.go`
- 太极子集群 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/check_in_taiji_step.go`
- 算力子集群 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/check_in_suanli_step.go`
- 获取 namespace 信息 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/get_namespace_quota_step.go`
- quota 参数校验 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/check_namespace_quota_step.go`
- quota 处理 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/handle_federation_namespace_quota_step.go`
- 状态回写 step：`bcs-services/bcs-federation-manager/internal/task/steps/sync_namespace_steps/update_federation_namespace_status_step.go`
- task 定义：`bcs-services/bcs-federation-manager/internal/task/tasks/sync_namespace_task.go`
- 集群客户端（常量 + 接口）：`bcs-services/bcs-federation-manager/internal/clients/cluster/client.go`
- step 公共常量：`bcs-services/bcs-federation-manager/internal/task/steps/types.go`

## 整体设计

整体上，这是一套两级定时驱动的 fan-out 同步模型：

1. `FedNamespaceControllerManager.StartLoop()` 每 3 分钟扫描一次联邦集群列表。
2. 对每个联邦集群维护一个 `SyncNamespaceQuotaController`。
3. 每个 controller 每 5 分钟扫描一次对应 `hostClusterID` 下的 namespace。
4. 每个 namespace 根据 `cluster-range` 拆分为多个目标子集群。
5. 为每个子集群分别派发一个 task。
6. task 在子集群中执行 namespace 创建或更新，并最终回写 host namespace 状态。

## 时序图

```
  FedNamespaceControllerManager     SyncNamespaceQuotaController      HostCluster        Store         TaskManager       SubCluster
            |                                  |                          |                |                |                |
            |--- 每3分钟拉取联邦集群列表 ------->|                          |                |                |                |
            |                                  |                          |                |                |                |
            |--- 为每个联邦集群维护 controller -->|                          |                |                |                |
            |                                  |                          |                |                |                |
            |                                  |-- 每5分钟 ListNamespace ->|                |                |                |
            |                                  |<-- namespaceList --------|                |                |                |
            |                                  |                          |                |                |                |
            |                        +---------+---[遍历每个 namespace]---+                |                |                |
            |                        |         |                          |                |                |                |
            |                        |         |-- 查 task 状态 --------->|                |-> GetTaskWithID |                |
            |                        |         |                          |                |                |                |
            |                        |    [若 task RUNNING/INITIALIZING => 跳过本 namespace]                |                |
            |                        |         |                          |                |                |                |
            |                        |    [否则继续]                      |                |                |                |
            |                        |         |                          |                |                |                |
            |                        |         |-- 解析 cluster-range --->|                |                |                |
            |                        |         |-- 查询真实子集群 ------->|                |                |                |
            |                        |         |<-- validSubClusterIDs ---|                |                |                |
            |                        |         |                          |                |                |                |
            |                        |    +----+--[遍历每个子集群]--------+                |                |                |
            |                        |    |    |                          |                |                |                |
            |                        |    |    |-- GetNamespace ----------+----------------|--------------->|                |
            |                        |    |    |                          |                |                |                |
            |                        |    |    | [若子集群已存在 namespace => 跳过该子集群]  |                |                |
            |                        |    |    |                          |                |                |                |
            |                        |    |    | [若不存在]               |                |                |                |
            |                        |    |    |-- GetManagedCluster ---->|                |                |                |
            |                        |    |    |<-- labels --------------|                |                |                |
            |                        |    |    |                          |                |                |                |
            |                        |    |    | [clusterType=taiji/suanli] => 跳过        |                |                |
            |                        |    |    | [mixercluster="true"]                     |                |                |
            |                        |    |    |     => Dispatch SyncHbNamespaceQuotaTask(含 host annotations) -->|          |
            |                        |    |    | [其他]                                    |                |                |
            |                        |    |    |     => Dispatch SyncNormalNamespaceQuotaTask ------------->|                |
            |                        |    |    |                          |                |                |                |
            |                        |    |    | [taskID 非空]            |                |                |                |
            |                        |    |    |-- UpdateNamespace ------>|                |                |                |
            |                        |    |    |   (写入 taskId)          |                |                |                |
            |                        |    +----+                          |                |                |                |
            |                        +---------+                          |                |                |                |
```

## 流程图

```
 ┌──────────────────────────────────────────┐
 │ StartLoop 每3分钟扫描联邦集群             │
 └─────────────────────┬────────────────────┘
                       │
                       ▼
 ┌──────────────────────────────────────────┐
 │ 为每个联邦集群启动/维护                   │
 │ SyncNamespaceQuotaController             │
 └─────────────────────┬────────────────────┘
                       │
                       ▼
 ┌──────────────────────────────────────────┐
 │ controller 每5分钟执行 processNamespaces  │
 └─────────────────────┬────────────────────┘
                       │
                       ▼
                ┌──────────────┐
                │hostClusterID │    是
                │  是否为空？   ├──────────► [结束本轮]
                └──────┬───────┘
                       │ 否
                       ▼
 ┌──────────────────────────────────────────┐
 │ ListNamespace(hostClusterID)              │
 └─────────────────────┬────────────────────┘
                       │
                       ▼
                ┌──────────────┐
                │namespaceList │    是
                │  是否为空？   ├──────────► [结束本轮]
                └──────┬───────┘
                       │ 否
                       ▼
 ┌──────────────────────────────────────────┐
 │ 遍历每个 namespace                       │◄─────────────────────────────────┐
 └─────────────────────┬────────────────────┘                                  │
                       │                                                       │
                       ▼                                                       │
                ┌──────────────────┐                                           │
                │ namespace 名称   │   是                                      │
                │   是否为空？     ├────────────────────────────────────────────┤
                └──────┬───────────┘                                           │
                       │ 否                                                    │
                       ▼                                                       │
                ┌──────────────────┐                                           │
                │ CreateNamespace  │   是                                      │
                │ TaskId 对应任务  ├────────────────────────────────────────────┤
                │   是否在运行？   │                                           │
                └──────┬───────────┘                                           │
                       │ 否                                                    │
                       ▼                                                       │
 ┌──────────────────────────────────────────┐                                  │
 │ 解析 cluster-range                       │                                  │
 └─────────────────────┬────────────────────┘                                  │
                       │                                                       │
                       ▼                                                       │
 ┌──────────────────────────────────────────┐                                  │
 │ 从 store 过滤合法子集群                   │                                  │
 └─────────────────────┬────────────────────┘                                  │
                       │                                                       │
                       ▼                                                       │
                ┌──────────────────┐                                           │
                │validSubClusterIDs│   是                                      │
                │  是否为空？      ├────────────────────────────────────────────┤
                └──────┬───────────┘                                           │
                       │ 否                                                    │
                       ▼                                                       │
 ┌──────────────────────────────────────────┐                                  │
 │ 遍历每个 subClusterID                    │◄────────────────────────┐        │
 └─────────────────────┬────────────────────┘                         │        │
                       │                                              │        │
                       ▼                                              │        │
                ┌──────────────────┐                                  │        │
                │ 子集群已存在     │   是                              │        │
                │ 同名 namespace？ ├──────────────────────────────────►│        │
                └──────┬───────────┘                                  │        │
                       │ 否                                           │        │
                       ▼                                              │        │
 ┌──────────────────────────────────────────┐                         │        │
 │ 读取 ManagedCluster labels               │                         │        │
 └─────────────────────┬────────────────────┘                         │        │
                       │                                              │        │
                       ▼                                              │        │
                ┌──────────────────┐                                  │        │
                │ clusterType      │                                  │        │
                │ =taiji/suanli?   │   是                             │        │
                └──┬───────────────┴──────────────────────────────────►        │
                   │ 否（default 分支）                               │        │
                   ▼                                                  │        │
                ┌──────────────────┐                                  │        │
                │ mixercluster     │                                  │        │
                │ =="true"?        │                                  │        │
                └──┬───────┬───────┘                                  │        │
            是     │       │ 否                                       │        │
          ┌────────┘       └────────┐                                 │        │
          │                         │                                 │        │
          ▼                         ▼                                 │        │
 ┌────────────────────┐  ┌──────────────────┐                         │        │
 │ 序列化 host ns     │  │ 创建 SyncNormal..│                         │        │
 │ annotations        │  │ NamespaceTask    │                         │        │
 │ 创建 SyncHb...     │  └───────┬──────────┘                         │        │
 │ NamespaceTask      │          │                                    │        │
 │ (含 labels +       │          │                                    │        │
 │  hostAnnotations)  │          │                                    │        │
 └───────┬────────────┘          │                                    │        │
         │                       │                                    │        │
         └─────┬─────────────────┘                                    │        │
               │                                                      │        │
               ▼                                                      │        │
        ┌──────────────┐                                              │        │
        │  taskID      │   是                                         │        │
        │  是否为空？  ├─────────────────────────────────────────────►│        │
        └──────┬───────┘                                              │        │
               │ 否                                                   │        │
               ▼                                                      │        │
 ┌──────────────────────────────────────────┐                         │        │
 │ 更新 host namespace 的                   │                         │        │
 │ CreateNamespaceTaskId                    ├────────────────────────►│        │
 └──────────────────────────────────────────┘                         │        │
                                                    [遍历下一个子集群]─┘        │
                                                                               │
                                                            [遍历下一个 namespace]
```

## namespace 同步主链路说明

### 1. 控制器管理层

`FedNamespaceControllerManager` 是 namespace 同步的外层管理器：

- 周期性查询所有联邦集群。
- 为每个联邦集群维护一个独立 controller。
- 当联邦集群不存在时，停止对应 controller。
- 当联邦集群新增时，创建新的 controller 并在后台运行。

这意味着 namespace 同步逻辑是按联邦集群隔离运行的。

### 2. namespace 扫描阶段

`SyncNamespaceQuotaController.processNamespaces()` 的行为如下：

1. 校验 `hostClusterID` 是否为空。
2. 从 host 集群拉取全部 namespace。
3. 遍历 namespace 列表。
4. 对每个 namespace 执行 `processSingleNamespace()`。

当前实现并不会在这里优先过滤 `is-federated-namespace=true` 的 namespace，而是直接扫描 host 集群全部 namespace，再通过后续的 `cluster-range` 和 task 状态判断是否继续处理。

### 3. 单个 namespace 的判定逻辑

`processSingleNamespace()` 主要分为三步：

1. `shouldProcessNamespace()`：判断当前 namespace 是否应该继续处理。
2. `extractSubClusterIDs()`：从 annotation `federation.bkbcs.tencent.com/cluster-range` 中提取目标子集群列表。
3. `getValidSubClusterIDs()`：将提取到的子集群列表与 store 中真实存在的子集群做交集。

如果最后得到的 `validSubClusterIDs` 为空，则本轮不会向任何子集群发起同步。

### 4. 子集群分发逻辑

`syncNamespaceToSubClusters()` 会对每个目标子集群执行：

1. 先检查子集群中是否已经存在同名 namespace（`checkSubClusterNamespace`）。
2. 若不存在，则读取 `ManagedCluster` 的 labels（`getManagedClusterAndBuildTask`）。
3. 根据 labels 进行两级判断来决定创建哪种 task（`buildSubClusterTask`）：
   - **第一级**：读取 `subscription.bkbcs.tencent.com/clustertype`，若为 `taiji` 或 `suanli` 则直接跳过（返回空 taskID）。
   - **第二级**：对于其余所有类型（包括 `normal`、`hunbu` 及其他），再读取 `subscription.bkbcs.tencent.com/mixercluster` label：
     - 若 `mixercluster == "true"`，则将 host namespace annotations 序列化后，连同 ManagedCluster labels 一起传入，创建 `SyncHbNamespaceQuotaTask`。
     - 否则创建 `SyncNormalNamespaceQuotaTask`。
4. task 派发成功后，把 taskID 回写到 host namespace 的 annotation `federation.bkbcs.tencent.com/create-namespace-taskId` 中。

这一层是同步入口的核心 fan-out 行为。

## task 设计说明

### 普通子集群 task

普通子集群走 `SyncNormalNamespaceQuotaTask`，step 顺序如下：

1. `GetNamespaceQuotaStep`
2. `CheckInNormalStep`
3. `UpdateFederationNamespaceStatusStep`

task 构建时通过 `CommonParams` 传递：`namespace`、`hostClusterID`、`subClusterID`。

```
 ┌───────────────────────────────────────────────────┐
 │          SyncNormalNamespaceQuotaTask              │
 └──────────────────────┬────────────────────────────┘
                        │
                        ▼
 ┌───────────────────────────────────────────────────┐
 │ Step 1: GetNamespaceQuotaStep                     │
 │                                                   │
 │  从 host 集群读取 namespace 对象                   │
 │  -> 序列化后写入 SyncNamespaceQuotaKey             │
 │                                                   │
 │  从 host 集群读取 quotaList                        │
 │  -> 序列化后写入 NamespaceQuotaListKey             │
 └──────────────────────┬────────────────────────────┘
                        │
                        ▼
 ┌───────────────────────────────────────────────────┐
 │ Step 2: CheckInNormalStep                         │
 │                                                   │
 │  反序列化 host namespace                           │
 │  查询子集群 namespace                              │
 │                                                   │
 │       ┌──────────────────────────┐                │
 │       │ 子集群 namespace 存在？   │                │
 │       └────┬────────────┬────────┘                │
 │         否 │            │ 是                      │
 │            ▼            ▼                          │
 │  ┌──────────────┐ ┌──────────────────────┐        │
 │  │buildNormal   │ │ buildNormal          │        │
 │  │SubCluster    │ │ SubClusterAnnotations│        │
 │  │Annotations() │ │ 合并到子集群 ns      │        │
 │  │+ CreateNs()  │ │ 然后 UpdateNamespace │        │
 │  └──────────────┘ └──────────────────────┘        │
 │  * 注：update 分支从当前入口基本走不到             │
 └──────────────────────┬────────────────────────────┘
                        │
                        ▼
 ┌───────────────────────────────────────────────────┐
 │ Step 3: UpdateFederationNamespaceStatusStep       │
 │                                                   │
 │  回写 host namespace annotations（带 retry）：     │
 │  - CreateNamespaceTaskId = 当前 taskId            │
 │  - HostClusterNamespaceStatus = Success           │
 │  - NamespaceUpdateTimestamp = 当前时间            │
 └───────────────────────────────────────────────────┘
```

#### GetNamespaceQuotaStep

这个 step 会从 host 集群中读取：

- namespace 本身，并序列化后写入 `syncNamespaceQuota`
- namespace 下的 quota 列表，并序列化后写入 `quotaList`

从当前链路来看，namespace 信息会被后续 step 使用，而 `quotaList` 虽然被取出来，但在普通子集群的后续 step 中并没有被消费（`quotaList` 主要被 taiji/suanli 的独立 task 链路使用）。

#### CheckInNormalStep

这个 step 会反序列化前一步存下来的 host namespace，然后在目标子集群中：

- 如果 namespace 不存在，则通过 `buildNormalSubClusterAnnotations()` 从 host namespace annotations 中按优先级提取 projectcode/businessid，再调用 `CreateClusterNamespace()` 创建。
- 如果 namespace 已存在，则通过 `buildNormalSubClusterAnnotations()` 生成新的 annotations，**合并**（非替换）到子集群 namespace 的现有 annotations 中，再执行更新。

需要注意的是，虽然 step 自身支持"已存在则更新"，但 controller 入口层已经在 task 派发前做过一次存在性检查，因此这条 update 分支在当前主链路中基本不会真正走到。

#### UpdateFederationNamespaceStatusStep

这个 step 会回写 host 集群中的 namespace annotation：

- `CreateNamespaceTaskId`
- `HostClusterNamespaceStatus=Success`
- `NamespaceUpdateTimestamp=当前时间`

这一步带有 retry 逻辑（最多 5 次，起始延迟 1 分钟，指数退避，最大延迟 10 分钟），属于最终状态回写阶段。

### 混部子集群 task

混部子集群走 `SyncHbNamespaceQuotaTask`，step 顺序如下：

1. `CheckInHunbuStep`
2. `UpdateFederationNamespaceStatusStep`

task 构建时通过 `CommonParams` 传递：`namespace`、`hostClusterID`、`subClusterID`、`managedClusterLabels`（JSON 序列化后的 ManagedCluster labels）、`hostNamespaceAnnotations`（JSON 序列化后的 host namespace annotations，可选）。

```
 ┌───────────────────────────────────────────────────┐
 │          SyncHbNamespaceQuotaTask                  │
 └──────────────────────┬────────────────────────────┘
                        │
                        ▼
 ┌───────────────────────────────────────────────────┐
 │ Step 1: CheckInHunbuStep                          │
 │                                                   │
 │  从 CommonParams 反序列化：                        │
 │  - ManagedCluster labels                          │
 │  - host namespace annotations                     │
 │                                                   │
 │  根据 labels 生成混部 annotations (buildHbReq)     │
 │  根据 host annotations 生成 projectcode/businessid │
 │        (buildNormalSubClusterAnnotations)          │
 │  查询子集群 namespace                              │
 │                                                   │
 │       ┌──────────────────────────┐                │
 │       │ 子集群 namespace 存在？   │                │
 │       └────┬────────────┬────────┘                │
 │         否 │            │ 是                      │
 │            ▼            ▼                          │
 │  ┌──────────────┐ ┌──────────────────────┐        │
 │  │ 混部 annot.  │ │ 将混部 annotations   │        │
 │  │ + project    │ │ + projectcode/       │        │
 │  │ annotations  │ │ businessid           │        │
 │  │ CreateNs()   │ │ 合并到子集群 ns      │        │
 │  └──────────────┘ │ 然后 UpdateNamespace  │        │
 │                   └──────────────────────┘        │
 │  * 注：update 分支从当前入口基本走不到             │
 └──────────────────────┬────────────────────────────┘
                        │
                        ▼
 ┌───────────────────────────────────────────────────┐
 │ Step 2: UpdateFederationNamespaceStatusStep       │
 │                                                   │
 │  回写 host namespace annotations（带 retry）：     │
 │  - CreateNamespaceTaskId = 当前 taskId            │
 │  - HostClusterNamespaceStatus = Success           │
 │  - NamespaceUpdateTimestamp = 当前时间            │
 └───────────────────────────────────────────────────┘
```

它与普通子集群最大的区别是：

- 不需要 `GetNamespaceQuotaStep`（不从 host 集群重新读取 namespace），而是直接使用 controller 派发时通过 `CommonParams` 传入的 `managedClusterLabels` 和 `hostNamespaceAnnotations`。
- annotations 由两部分组成：
  1. 混部专有 annotations：通过 `buildHbReq()` 从 ManagedCluster labels 动态生成。
  2. projectcode/businessid annotations：通过 `buildNormalSubClusterAnnotations()` 从 host namespace annotations 按优先级提取（与普通子集群共享同一函数）。

混部专有 annotations 示例：

- `mixer.kubernetes.io/is-mixer-namespace`
- `tke.cloud.tencent.com/networks`
- `mixer.kubernetes.io/preemption-policy`
- `mixer.kubernetes.io/priority-class`
- `mixer.kubernetes.io/priority-value`

### taiji / suanli 子集群 task（周期扫描链路中未启用）

虽然在周期扫描链路中 `buildSubClusterTask()` 对 `taiji` 和 `suanli` 直接返回空 taskID（跳过），但代码中存在完整的 task 和 step 实现，可通过其他链路（如 `HandleNamespaceQuotaTask`）使用：

**SyncTjNamespaceQuotaTask** (太极):

1. `GetNamespaceQuotaStep`
2. `CheckInTaijiStep` — 通过第三方服务查询/创建/更新太极 namespace quota
3. `UpdateFederationNamespaceStatusStep`

**SyncSlNamespaceQuotaTask** (算力):

1. `GetNamespaceQuotaStep`
2. `CheckInSuanliStep` — 通过第三方服务查询/创建/更新算力 namespace quota
3. `UpdateFederationNamespaceStatusStep`

两者都依赖 `GetNamespaceQuotaStep` 获取的 `quotaList` 信息来向第三方服务提交 quota 配置。

## 子集群 namespace annotations 构建逻辑

### 核心原则

同步到子集群的 namespace **不会**携带 host cluster namespace 上的所有 annotations。子集群 namespace 只携带经过筛选和转换后的特定 annotations。

### projectcode / businessid 优先级规则

创建或更新子集群 namespace 时，由 `buildNormalSubClusterAnnotations()` 按照以下优先级从 host cluster namespace annotations 中提取项目/业务标识（匹配到任一规则后立即返回，不再继续匹配）：

**优先级 1：CMDB 二级业务 ID**

如果 host namespace annotations 中存在 `federation.bkbcs.tencent.com/obs-cmdb-business-id={cmdb二级业务id}`，则子集群 namespace 添加：

```
io.tencent.bcs.businesslevel2id={cmdb二级业务id}
```

**优先级 2：计费项目 code**

如果 host namespace annotations 中存在 `federation.bkbcs.tencent.com/bill-projectcode={项目code}`（表示真实计费的项目 code），则子集群 namespace 添加：

```
io.tencent.bcs.projectcode={项目code}
```

**优先级 3：原始项目 code**

如果 host namespace annotations 中存在 `io.tencent.bcs.projectcode={项目code}`，则子集群 namespace 添加：

```
io.tencent.bcs.projectcode={项目code}
```

### 普通子集群 annotations

普通子集群的 namespace 只携带上述 projectcode/businessid 规则产生的 annotations，不携带其他 host namespace annotations。

### 混部子集群 annotations

混部子集群（`mixercluster=="true"` 的集群）的 namespace 携带两部分 annotations：

1. 混部专有 annotations（由 `buildHbReq()` 根据 ManagedCluster labels 生成）：
   - `mixer.kubernetes.io/is-mixer-namespace`：当 `subscription.bkbcs.tencent.com/mixercluster=="true"` 时设置
   - `tke.cloud.tencent.com/networks`：取自 label `subscription.bkbcs.tencent.com/mixercluster-tke-networks`（若非空）
   - `mixer.kubernetes.io/preemption-policy`：当 label `subscription.bkbcs.tencent.com/mixercluster-low-priority=="true"` 时设为 `"Never"`
   - `mixer.kubernetes.io/priority-class`：同上条件，设为 `"offline-pod-priority"`
   - `mixer.kubernetes.io/priority-value`：同上条件，设为 `"-100"`

2. projectcode/businessid annotations（由 `buildNormalSubClusterAnnotations()` 从 host namespace annotations 按优先级提取，与普通子集群使用相同规则）

```
 ┌──────────────────────────────────────────────────┐
 │     host namespace annotations                    │
 └──────────────────────┬───────────────────────────┘
                        │
           ┌────────────┼────────────┐
           │            │            │
           ▼            ▼            ▼
 ┌──────────────┐ ┌──────────┐ ┌──────────────┐
 │ obs-cmdb-    │ │ bill-    │ │ projectcode  │
 │ business-id  │ │ project  │ │              │
 │ (优先级 1)   │ │ code     │ │ (优先级 3)   │
 │              │ │ (优先级2)│ │              │
 └──────┬───────┘ └────┬─────┘ └──────┬───────┘
        │              │              │
        ▼              ▼              ▼
 ┌──────────────────────────────────────────────────┐
 │ buildNormalSubClusterAnnotations()                │
 │ 按优先级匹配，返回单一 annotation                 │
 └──────────────────────┬───────────────────────────┘
                        │
           ┌────────────┴────────────┐
           │                         │
           ▼                         ▼
 ┌──────────────────┐    ┌──────────────────────┐
 │ 普通子集群       │    │ 混部子集群           │
 │ namespace        │    │ namespace            │
 │                  │    │                      │
 │ 仅 projectcode/ │    │ buildHbReq()         │
 │ businessid      │    │ + projectcode/       │
 │                  │    │   businessid         │
 └──────────────────┘    └──────────────────────┘
```

## 边界条件

## 1. 服务启动后不会立即执行一次同步

当前实现是纯 ticker 驱动：

- 管理器每 3 分钟跑一次
- controller 每 5 分钟跑一次

服务刚启动时不会立刻同步，首次执行存在延迟。

## 2. `cluster-range` 为空时不会默认同步到所有子集群

当前实现只是读取 annotation 中声明的 `cluster-range`，然后与 store 中已有子集群做交集。若 `cluster-range` 为空，则不会补齐"全部子集群"这一默认行为。

## 3. `cluster-range` 中的值会 trim 空格并转大写

`extractSubClusterIDs()` 对 `strings.Split()` 的结果先执行 `strings.TrimSpace()` 再执行 `strings.ToUpper()`，因此以下写法可以正确解析：

```text
BCS-K8S-001, BCS-K8S-002
```

## 4. 子集群已存在同名 namespace 时会被视为冲突

controller 入口中的 `checkSubClusterNamespace()` 只要发现子集群已有同名 namespace，就直接返回错误并跳过该子集群。这是当前同步入口的重要约束。

## 5. task 查询失败时仍会继续处理

`shouldProcessNamespace()` 在读取历史 task 状态失败时，默认返回 `true`，即继续尝试同步。这种策略偏向"尽量推进"，但也带来重复派发风险。

## 6. taiji 与 suanli 子集群在周期扫描链路中被跳过

`buildSubClusterTask()` 中对 `clusterType` 为 `taiji` 和 `suanli` 的子集群直接返回空 taskID，不在周期扫描链路中做同步。但代码中存在完整的 `SyncTjNamespaceQuotaTask` 和 `SyncSlNamespaceQuotaTask` 实现（包含 `CheckInTaijiStep`、`CheckInSuanliStep`），这些 task 可通过其他链路（如 `HandleNamespaceQuotaTask`）被调用。

## 7. 混部判断依据是 mixercluster label 而非 clusterType

`buildSubClusterTask()` 中判断一个子集群是否为混部集群，不是依据 `subscription.bkbcs.tencent.com/clustertype` 的值是否为 `hunbu`，而是依据 `subscription.bkbcs.tencent.com/mixercluster` label 是否为 `"true"`。这意味着一个 clusterType 为 `normal` 但 `mixercluster=="true"` 的集群也会走混部同步逻辑。
