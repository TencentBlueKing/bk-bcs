# drplan-gen Helmfile Release Hooks 迁移方案

## 背景

当前 `drplan-gen helmfile` 模式已经支持从单个 helmfile release 生成：

- `HelmChart` action
- `Globalization` action
- `Subscription` action

但还没有覆盖 helmfile `releases[].hooks`。

需要特别区分两类 hook：

1. chart rendered YAML 中的 `helm.sh/hook`
2. helmfile release 级别的 `hooks`

这两类 hook 不是同一类能力。

chart hook 的载体本身就是 Kubernetes 资源，因此现有普通 YAML 生成器可以把它们转成资源型 workflow action。

helmfile release hook 的载体是：

- `command`
- `args`
- `events`

本质上是在执行 helmfile 的机器上运行命令，不是直接下发 Kubernetes 资源。因此不能直接复用当前普通 YAML hook 的生成逻辑。

## Helmfile 原始语义

helmfile 官方对 release hooks 的定义如下：

- `prepare`
  - release 从 YAML 加载完成后、执行前触发
- `preapply`
  - `helmfile apply` 中，在 release uninstall/install/upgrade 前触发
- `presync`
  - release sync 到集群前触发，sync 语义为 install/upgrade
- `postsync`
  - release sync 到集群后触发，不论 sync 成功还是失败
- `cleanup`
  - release 处理完成后触发
- `preuninstall` / `postuninstall`
  - uninstall 前后触发

同时还有两个关键语义：

- 同一个 release 的 hooks 按 helmfile 中声明顺序执行
- hook 是否成功，本质上取决于命令退出码；非 0 退出会被视为失败
- `presync` 一旦失败，该 release 的 install/upgrade 不再继续
- `postsync` 是 “always run after sync result” 语义，不等价于普通的 post-install 成功后步骤
- `presync` 失败后，`postsync` 与 `cleanup` 仍然会触发

## 当前项目现状

### 已有能力

当前项目已经具备：

- `KubernetesResource` action
- `HelmChart` / `Globalization` / `Subscription` action
- `Subscription waitReady`
- `Subscription` 的 `PerCluster` 执行模式
- stage 串行执行能力
- stage 内 workflow 顺序执行能力
- workflow 内 action 顺序执行能力

当前 stage 的执行模型是：

- stage 之间按计划顺序串行执行
- 只有 stage 内的 workflows 才可能并行
- `Parallel` 不影响 stages 之间的顺序

因此从编排能力上，完全可以表达：

- 先执行一组 hook
- 再执行主部署 workflow
- 最后执行一组后置 hook

### 现有限制

当前 `Job` action 默认是在主控集群执行：

- controller 直接创建 host-cluster Job
- 不会自动下发到子集群

这和当前需求不一致。当前需求明确要求：

- hook 在每个目标子集群都执行一遍
- 不同子集群之间互不干扰

此外，当前普通 YAML hook 的 unified workflow 模型虽然支持：

- `when`
- `dependsOn`
- hook cleanup

但它主要面向资源型 hook，不适合直接映射 helmfile release hooks，原因包括：

1. helmfile release hook 的本体不是资源，而是命令
2. `postsync` 要求失败后仍然执行
3. 现有 unified workflow 默认 `FailFast`，主 action 失败时不会继续执行后置步骤

## 目标

新增 helmfile release hook 迁移能力，使 `drplan-gen helmfile` 能把 release hook 转成可执行的 DRPlan 编排。

目标包括：

- 支持从 helmfile Go 包读取 `release.Hooks`
- 支持将脚本类 hook 迁移为子集群执行的 `Manifest(template=Job)`
- 保持 helmfile 中同事件 hooks 的声明顺序
- 保持 `presync` / `preapply` 在主部署前执行
- 尽量逼近 `postsync` “无论成功失败都执行”的语义
- 支持每个目标子集群各执行一遍
- 保持不同子集群的执行资源、等待状态相互隔离

## 非目标

首版不做：

- 在 controller 容器中直接执行本地 shell 脚本
- 完整复刻 helmfile 的 `.Event.*`、`.Environment.*`、`.Release.*` 命令模板上下文
- 全量支持所有 release hook event

## 方案对比

### 方案一：在 controller 本地直接执行 shell 脚本

做法：

- 解析到 hook 后，controller 直接在自身容器里执行脚本

优点：

- 最接近 helmfile 原始执行方式
- 不需要额外 Job 资源

问题：

- controller 镜像必须带齐所有脚本和依赖
- 权限边界差，脚本直接运行在控制面容器中
- 可观测性、隔离性、失败重试都差
- 多租场景风险高

结论：

- 不建议采用

### 方案二：将 hooks 内联到主 workflow，作为 host-cluster Job actions 执行

做法：

- 在 `workflow-execute.yaml` 中插入若干 `Job` action
- 通过 action 顺序或 `dependsOn` 控制 hook 前后位置

优点：

- 结构简单
- 不需要新增 workflow/stage

问题：

- Job 只会在 host 集群执行，不会进入子集群
- `postsync` 很难表达“主流程失败后也必须执行”
- 主 workflow 一旦 `FailFast`，后置 action 会被跳过
- 语义容易与现有资源型 hook 混淆

结论：

- 不满足当前需求

### 方案三：使用 `Subscription` hook，将 feed 对象下发到子集群执行

做法：

- hook 不生成 `Job` action
- 先生成一个 hub 侧 feed 对象
- 再生成 `Subscription` action 引用该 feed
- 打开 `waitReady: true`
- 打开 `clusterExecutionMode: PerCluster`

优点：

- Job 真正运行在子集群
- 可以直接复用现有 `Subscription waitReady` 和 `PerCluster` 能力
- 每个子集群有独立 child Subscription 和独立等待状态
- 与当前普通 YAML hook 的资源型模型保持一致

问题：

- 生成器需要新增“把 helmfile command/args 组装成 Job manifest”的逻辑
- 最终 action 汇总状态仍会受 failurePolicy 影响

结论：

- 推荐采用

### 方案四：按 hook 事件组拆成独立 workflow / stage

做法：

- `presync` / `preapply` 拆成前置 workflow
- 主部署保持 `workflow-execute`
- `postsync` 拆成后置 workflow
- 每个逻辑 hook 在对应 workflow 内按固定 action 组顺序执行

优点：

- 语义清晰
- 能利用 stage 串行模型表达前后关系
- 更容易单独控制失败策略
- 更适合后续继续扩展 `preuninstall/postuninstall`

问题：

- 生成物结构会比当前 helmfile 模式复杂
- `postsync` 仍需要借助 plan 级别 failurePolicy 才能在失败后继续执行

结论：

- 推荐采用

## 推荐设计

首版采用：

- `scripts -> hook-runner 镜像 -> Job 模板`
- `Job 模板 -> Clusternet Manifest`
- `Manifest -> Subscription hook action -> 子集群执行`
- `Subscription hook action` 开启 `waitReady: true` 与 `clusterExecutionMode: PerCluster`
- `hook event -> 独立 workflow/stage`

### 脚本承载方式

约定将原有 repo 中的脚本统一打包到一个公共镜像，例如：

```text
hook-runner:<tag>
```

镜像中包含：

- 原 `scripts/` 目录
- `bash`
- `curl`
- `kubectl`
- `jq`
- 其他业务脚本运行所需依赖

每个 helmfile release hook 迁移为一个 Job 模板，其容器入口类似：

```bash
/bin/bash -c "./scripts/add_bkrepo_bucket.sh blueking http://... repo user pass true"
```

这样做的目的：

- 避免 controller 直接执行 shell
- 将 hook 的执行隔离到子集群中的独立 Pod
- 更容易复用镜像、权限和网络策略

### Feed 承载模型

这里不能直接在 hub 侧创建原生 `batch/v1 Job`。

原因是：

- 原生 `Job` 一旦创建在 hub 集群，就会在 hub 集群自己执行
- 这违反“只在子集群执行”的要求

因此首版使用：

- `apps.clusternet.io/v1alpha1 Manifest`

Manifest 只在 hub 侧保存 raw 模板，本身不会像原生 Job 一样在 hub 集群调度运行。
该 Manifest 需要按 Clusternet shadow 对象形态放在 `clusternet-reserved` 命名空间，并携带真实 Job feed 对应的 config labels。

推荐的单个 hook 展开模型为：

1. `apply-hook-manifest`
   - `type: KubernetesResource`
   - 在 hub 集群创建或更新一个 `Manifest`
   - `Manifest.template` 中存放目标 `Job`
2. `run-hook-subscription`
   - `type: Subscription`
   - `feeds` 引用真实的 `batch/v1 Job`
   - `clusterExecutionMode: PerCluster`
   - `waitReady: true`

这样可以同时满足：

- hub 侧只保存 Manifest 模板对象，不真正跑 Job
- Clusternet 按真实 Job feed 匹配 Manifest，并将模板中的 Job 下发到子集群
- hook 仍能复用当前 `Subscription + PerCluster` 的分发模型

### 执行域模型

本方案明确约定：

- hook 在每个目标子集群执行一遍
- 不在 host 集群执行
- 同一 hook action 会按目标集群拆成多个 child Subscription

也就是说，对一个 `presync` hook：

- 子集群 A 执行一份 Job
- 子集群 B 执行一份 Job
- 子集群 C 执行一份 Job

每个子集群使用独立的 child Subscription、独立的 Job、独立的等待链路。

### 事件支持范围

首版建议支持：

- `preapply`
- `presync`
- `postsync`

首版暂不支持：

- `prepare`
- `cleanup`
- `preuninstall`
- `postuninstall`

理由：

- `presync` 是当前用户最直接的迁移诉求
- `preapply` 与 `presync` 都属于前置变更类 hook，模型类似
- `postsync` 虽然更复杂，但在迁移时经常需要
- `prepare/cleanup` 更接近 helmfile 进程生命周期，而不是集群动作生命周期
- `preuninstall/postuninstall` 可在后续删除链路补齐

### 编排模型

假设 release 同时存在：

- `preapply`
- `presync`
- 主部署
- `postsync`

推荐生成：

1. 实际存在 `preapply` hook 时生成 `stage-preapply`
2. 实际存在 `presync` hook 时生成 `stage-presync`
3. 始终生成 `stage-execute`
4. 实际存在 `postsync` hook 时生成 `stage-postsync`

每个 stage 默认：

- `parallel: false`

每个 stage 下引用对应 workflow；没有 action 的 hook workflow 不生成：

- `workflow-preapply`
- `workflow-presync`
- `workflow-execute`
- `workflow-postsync`

stage 间顺序由 DRPlan 现有串行执行模型保证。

### Workflow 内部模型

同一个 hook 事件里的多个 hook，放在同一个 workflow 中。

例如 helmfile 中：

```yaml
hooks:
  - events: ["presync"]
    command: "./scripts/a.sh"
  - events: ["presync"]
    command: "./scripts/b.sh"
```

生成：

- `workflow-presync`
  - `presync-hook-1`
  - `presync-hook-2`

其执行顺序要求：

- 与 helmfile 中声明顺序一致

每个逻辑 hook 在 workflow 中展开为两个 action：

1. `apply-hook-manifest`
2. `run-hook-subscription`

其中实际下发到子集群的是第二个 `Subscription` action，它带：

- `waitReady: true`
- `clusterExecutionMode: PerCluster`

首版不强依赖 `dependsOn`，直接依赖 workflow 内 action 的天然顺序执行即可。

## PerCluster 语义

当前需求里“不同子集群互不干扰”有两层含义：

1. 资源层面互不干扰
2. 状态观察层面互不干扰

资源层面当前已有保障：

- `PerCluster` 会把一个 Subscription action 拆成多个 child Subscription
- 每个 child Subscription 只面向一个目标集群
- 每个目标集群里的 Job 独立创建、独立等待、独立失败

状态层面当前也已有基础能力：

- action status 中会记录 `ClusterStatuses`
- 每个 cluster status 单独反映对应子集群的执行结果

需要说明并明确写入实现要求的是：

- 最终 action 的汇总 phase 仍然应该是聚合结果
- 但聚合必须发生在“所有目标子集群都执行结束之后”
- 不能因为单个子集群失败就提前取消其他尚未完成的子集群执行

因此“互不干扰”在首版的准确含义应当是：

- 各子集群使用独立资源执行 hook
- 各子集群的 readiness/失败状态独立可见
- 某个子集群失败，不提前取消其他子集群
- 所有目标子集群执行结束后，再统一聚合结果
- 若任一 cluster 失败，则该 action 最终记为 `Failed`

这意味着首版实现不能直接复用当前 `PerCluster + FailFast` 的取消行为，而需要补一个 hook 场景下的聚合策略：

- `run-hook-subscription` 在 cluster fan-out 阶段必须等待所有目标子集群完成
- 最终 phase 再按聚合结果决定成功或失败

## 现有能力复用

首版不需要增强 `JobActionExecutor`。

原因是当前目标是：

- 通过 `KubernetesResource` 在 hub 侧 `clusternet-reserved` 命名空间创建 `Manifest(template=Job)`
- 再通过 `Subscription` 订阅真实的 `batch/v1 Job` feed，下发 Manifest 模板中的 Job
- 由 `Subscription waitReady` 去等待子集群中的 Job 完成

当前仓库里已经具备：

- `Subscription waitReady`
- 子集群 `Job` readiness 判定
- `PerCluster` child Subscription 执行模型

兼容已有手写 workflow 时仍保留一处小增强：

- 当 feed kind 为 `Manifest` 时，`waitReady` 需要能解出 `Manifest.template` 中的真实对象
- 对当前生成器输出，feed 已经直接指向 `batch/v1 Job`，`waitReady` 会直接检查子集群 Job

因此首版重点是：

- 生成器扩展
- `Subscription waitReady` 对真实 Job feed 的检查，以及对旧 Manifest feed 的兼容解析

## `postsync` 语义设计

这是本方案的关键点。

helmfile 的 `postsync` 语义是：

- release sync 完成后触发
- 无论 sync 成功还是失败都执行

当前 DRPlan 框架没有原生的 `finally` / `alwaysRun` 语义，因此首版采用近似实现：

- 将 `postsync` 放到独立后置 stage
- 将 `DRPlan.spec.failurePolicy` 设为 `Continue`

这里明确要求使用的是：

- `plan.Spec.FailurePolicy=Continue`

而不是：

- `stage.FailurePolicy=Continue`

原因是当前执行器真正控制“前一个 stage 失败后，下一个 stage 是否继续”的是 plan 级别 failurePolicy，stage 级别 failurePolicy 目前并未在执行路径中生效。

这样在 `stage-execute` 失败时：

- execution 不会立刻退出
- `stage-postsync` 仍然会继续执行

最终整次 execution 的 phase 仍会因为存在失败 stage 而保持 `Failed`。

这种行为与 helmfile `postsync` 最接近。

### 为什么不复用现有 unified workflow 模型

现有普通 YAML hook 的 unified workflow 设计更适合：

- 资源型 pre/post hook
- 主流程成功后再执行 post hook

但它不适合 release hooks 的 `postsync`，原因是：

- 主 action 失败时，workflow 默认 `FailFast`
- 后续 action 会被跳过
- 无法表达 “即便执行失败也要继续跑后置 hook”

因此 `postsync` 必须提升到 workflow/stage 级别处理，而不是继续作为主 workflow 里的末尾 action。

## 生成规则

### 命名规则

建议命名：

- `workflow-preapply.yaml`，仅存在 `preapply` hook 时生成
- `workflow-presync.yaml`，仅存在 `presync` hook 时生成
- `workflow-execute.yaml`
- `workflow-postsync.yaml`，仅存在 `postsync` hook 时生成

workflow 名称：

- `<release>-preapply`
- `<release>-presync`
- `<release>-execute`
- `<release>-postsync`

stage 名称：

- `preapply`
- `presync`
- `execute`
- `postsync`

### Hook 转子集群 Job 的映射

以如下 helmfile hook 为例：

```yaml
hooks:
  - events: ["presync"]
    command: "./scripts/add_bkrepo_bucket.sh"
    args:
      - "blueking"
      - "http://bkfile.{{`{{.Values.domain.bkDomain }}`}}"
      - "{{`{{.Values.bknodeman.bkrepo.repoName}}`}}"
      - "{{`{{.Values.bknodeman.bkrepo.username}}`}}"
      - "{{`{{.Values.bknodeman.bkrepo.password}}`}}"
      - "true"
```
映射为 drplan 中的两步：

第一步，创建 hub 侧 `Manifest`：

```yaml
- name: apply-presync-add-bkrepo-bucket-manifest
  type: KubernetesResource
  resource:
    operation: Apply
    manifest: |
      apiVersion: apps.clusternet.io/v1alpha1
      kind: Manifest
      metadata:
        name: jobs.$(params.targetNamespace).presync-add-bkrepo-bucket
        namespace: clusternet-reserved
        labels:
          apps.clusternet.io/config.group: batch
          apps.clusternet.io/config.version: v1
          apps.clusternet.io/config.kind: Job
          apps.clusternet.io/config.name: presync-add-bkrepo-bucket
          apps.clusternet.io/config.namespace: $(params.targetNamespace)
      template:
        apiVersion: batch/v1
        kind: Job
        metadata:
          name: presync-add-bkrepo-bucket
          namespace: $(params.targetNamespace)
        spec:
          backoffLimit: 0
          ttlSecondsAfterFinished: 300
          template:
            spec:
              restartPolicy: Never
              containers:
                - name: hook
                  image: <hook-runner>
                  command:
                    - /bin/bash
                      - -c
                      - ./scripts/add_bkrepo_bucket.sh ...
```

第二步，创建 `Subscription` hook：

```yaml
- name: presync-add-bkrepo-bucket
  type: Subscription
  waitReady: true
  clusterExecutionMode: PerCluster
  timeout: 5m
  hookCleanup:
    beforeCreate: true
  subscription:
    operation: Apply
    namespace: $(params.feedNamespace)
    name: presync-add-bkrepo-bucket-sub
    spec:
      schedulingStrategy: Replication
      feeds:
        - apiVersion: batch/v1
          kind: Job
          namespace: $(params.targetNamespace)
          name: presync-add-bkrepo-bucket
      subscribers:
        - clusterAffinity: {}
```

### 参数化约定

建议为 hook workflow 增加与主 workflow 一致的参数来源：

- `feedNamespace`
- `targetNamespace`

后续如果需要，还可以扩展：

- `hookImage`
- `hookServiceAccount`

### 命名与重跑策略

hook 需要支持后续重复执行，例如：

- 首次安装后的再次升级
- 同一 release 的多次 `Upgrade` execution

当前 DR 模板能力只支持：

- `$(params.xxx)`
- `$(planName)`
- `$(outputs.xxx)`

首版不具备 execution 级唯一变量，因此不建议依赖“每次执行都生成一个全新随机名称”的策略。

首版推荐使用稳定命名：

- `Manifest.metadata.name` 使用基于 `release + event + hook序号` 的稳定名称
- `Subscription.metadata.name` 使用与 hook 对应的稳定名称
- 子集群 `Job.metadata.name` 也使用稳定名称

为了避免重复执行时的旧资源冲突，首版增加如下约定：

- hook `Subscription` action 默认开启 `hookCleanup.beforeCreate=true`
- 每次重新执行 hook 前，先清理上一次残留的 child Subscription
- 清理 child Subscription 后，再重新下发新的 hook Job

这样可以在不引入 execution 级随机命名能力的前提下，保证 hook 可重复执行。

对于子集群中已完成的 Job，建议模板默认附带：

- `ttlSecondsAfterFinished`

其目的不是保证幂等，而是减少历史残留并保留短时间排障窗口。

### Job 成功判定

hook 被迁移到子集群 Job 后，成功判定不再直接看 controller 侧命令返回值，而是看子集群中 Job 的完成状态：

- `Job.status.conditions[type=Complete,status=True]` 视为成功
- `Job.status.conditions[type=Failed,status=True]` 视为失败

再往下一层，本质上仍然等价于脚本退出码：

- 容器主进程退出码为 `0`，Pod 成功，Job 最终进入 `Complete`
- 容器主进程退出码非 `0`，Pod 失败，Job 在重试耗尽后进入 `Failed`

为了尽量贴近 helmfile hook “执行失败立即失败”的语义，建议 hook Job 默认使用：

- `restartPolicy: Never`
- `backoffLimit: 0`

### 资源生命周期

单个 hook 在 workflow 中建议拆成两类资源：

1. hub 侧 `Manifest`
2. hub 侧 `Subscription`，其子集群侧会展开出 child Subscription 和最终 Job

首版生命周期建议如下：

- `Manifest` 使用 `operation: Apply`
  - Manifest 名称稳定，重跑时可复用同名 feed 并更新模板
- `Subscription` 使用 `operation: Apply`
  - 并依赖 `hookCleanup.beforeCreate=true` 保证重跑前先清理旧的 child Subscription

这样做的考虑是：

- `Manifest` 是稳定 feed 载体，适合在稳定名称下反复 `Apply`，避免重跑时因同名对象 `AlreadyExists` 失败
- `Subscription` 代表一次 hook 投递动作，适合在稳定名称下通过 `Apply` 持续驱动新一轮执行，并在投递前清理旧的 child Subscription

## 失败策略

### 前置 hooks

对于：

- `preapply`
- `presync`

建议默认：

- workflow `failurePolicy: FailFast`

这里的“失败”定义为：

- 所有目标子集群都执行结束后聚合得到的最终失败

而不是：

- 某个子集群一失败就立刻中断其他子集群

因此前置 hooks 的预期行为是：

- 子集群之间不互相取消
- 全部子集群执行完成后再聚合
- 若聚合结果为失败，则阻断主部署

其中 `presync` 的对齐语义应为：

- 每个目标子集群各执行一次 `presync` Job
- 全部子集群执行结束后再统一聚合结果
- 若任一子集群失败，则 `presync` action/workflow 最终失败
- 主 `execute` stage 不继续执行
- 若 plan 配置了 `failurePolicy: Continue`，则后续 `postsync` stage 仍可继续执行

### 后置 hooks

对于：

- `postsync`

建议：

- 仍使用单独 workflow
- workflow 内 action 顺序执行
- 但依赖 plan 级别 `failurePolicy: Continue` 保证 execute 失败后仍能运行到 postsync

## 与现有实现的关系

### 与普通 YAML hook 生成器的关系

普通 YAML 生成器现有的 hook 方案保留不变。

原因：

- 它处理的是 chart manifest hook
- 输入对象是资源 YAML
- 现有 `dependsOn + when + hookCleanup` 模型已足够

helmfile release hook 是另一条生成链路，应在 `drplan-gen helmfile` 内单独处理。

### 与当前 helmfile 简化模式的关系

当前 helmfile 模式在无 release hooks 时，维持简化输出：

- 单 `execute` stage
- 单 `workflow-execute.yaml`

一旦检测到 release hooks，则切换到 hook-aware 模式：

- 按实际 hook 事件生成 `preapply / presync / postsync` stage / workflow，并始终生成 `execute`
- 主部署 workflow 仍保留为 `workflow-execute.yaml`

## 首版边界

首版明确边界如下：

- 支持 `preapply` / `presync` / `postsync`
- 不支持 `prepare` / `cleanup` / `preuninstall` / `postuninstall`
- 若源 helmfile hook 带有 `showlogs`，首版直接忽略，不生成额外能力
- 不在 controller 容器中执行 shell
- 不做通用的 `alwaysRun` 框架能力
- 通过 `postsync` 独立 stage + `plan.failurePolicy=Continue` 近似对齐 helmfile `postsync`
- hook 的 `PerCluster` 执行必须满足“所有目标子集群都执行结束后再聚合”

## 后续演进

后续可以继续补：

1. `preuninstall/postuninstall`
   - 对齐 Delete mode
2. `prepare/cleanup`
   - 作为 helmfile 解析阶段或执行生命周期扩展
3. `finallyWorkflows` / `alwaysRun`
   - 作为框架级能力，消除 `postsync` 对 plan-level `Continue` 的依赖

## 结论

推荐方案是：

- 把 helmfile release hooks 视为“命令型 hook”，不要硬套现有资源型 hook 逻辑
- 用统一 `hook-runner` 镜像承载脚本
- 将脚本封装成 `Job` 模板，再由 hub 侧 `Manifest(template=Job)` 作为 shadow 模板承载
- 再用 `Subscription + PerCluster + waitReady` 订阅真实 Job feed，并把模板下发到每个目标子集群执行
- 用独立 stage/workflow 表达 `preapply`、`presync`、`postsync`
- 复用并小幅增强 `Subscription waitReady`，让它能跟踪子集群 Job 完成，并兼容旧的 Manifest feed 写法
- 调整 hook 场景下的 `PerCluster` 聚合语义：不因单个子集群失败而提前取消其他子集群，全部结束后再统一判定 action 成功/失败
- 通过 `postsync` 独立 stage + `plan.failurePolicy=Continue` 近似实现 “失败后也执行 postsync”

这是当前代码基础上最稳妥、改动面可控、且最接近 helmfile 原语义的迁移路径。
