# drplan-gen 使用指南

## 概述

`drplan-gen` 用于将 `helm template` / `helmfile template` 的渲染结果转换为 DRPlan 编排文件（`DRPlan + DRWorkflow + DRPlanExecution`）。

渲染 YAML 模式当前默认仍为**单 workflow 模式**，但会按是否存在 hook 分成两条路径：

- 无 hook：生成简化的 `execute` stage 和 `workflow-execute.yaml`
- 有任意 hook：生成统一的 `install` stage 和 `workflow-install.yaml`
- 两种模式都只生成一个 workflow
- hook-aware 路径通过 `when + waitReady` 模拟 Helm install/upgrade/delete/rollback hook 语义

## 构建

```bash
make build-gen
ls -la bin/drplan-gen
./bin/drplan-gen --version
./bin/drplan-gen version
```

## 命令行参数

| 参数 | 短写 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `--name` | | 是 | | Release 名称 |
| `--namespace` | | 否 | `default` | 目标命名空间 |
| `--file` | `-f` | 否 | stdin | 输入 YAML 文件路径 |
| `--output` | `-o` | 否 | `.` | 输出目录 |
| `--version` | | 否 | `false` | 输出版本信息并退出 |

## 快速开始

### 1. 直接从 helm template 输入

```bash
helm template my-app ./my-chart | drplan-gen --name my-app --namespace production
```

### 2. 从文件输入

```bash
helm template my-app ./my-chart > rendered.yaml
drplan-gen --name my-app --namespace production -f rendered.yaml -o ./drplan-output/
```

### 3. 从 helmfile template 输入

```bash
helmfile -f helmfile.yaml template | drplan-gen --name my-release --namespace production
```

### 4. 使用项目测试 chart 生成

```bash
helm template demo-app ./testdata/charts/demo-app \
  | drplan-gen --name demo-app --namespace default -o ./output/
```

## 输出文件

| 文件名 | 说明 |
|---|---|
| `drplan.yaml` | DRPlan 资源，默认只包含一个 stage；是否为 `execute/install` 取决于是否存在 hook |
| `workflow-execute.yaml` | 无 hook 时生成的简化 DRWorkflow，仅包含主资源正向 action |
| `workflow-install.yaml` | 存在 hook 时生成的统一 DRWorkflow，包含 hooks + main resources |
| `drplanexecution-install.yaml` | Install 执行样例，带 `mode: Install` |
| `drplanexecution-upgrade.yaml` | Upgrade 执行样例，带 `mode: Upgrade` |
| `drplanexecution-delete.yaml` | Delete 执行样例，带 `mode: Delete` |
| `drplanexecution-revert.yaml` | Revert 样例，带 `mode: Rollback` |

## 默认生成模型

### 无 hook：简化 execute workflow

当渲染结果中不存在任何 Helm hook 注解时，生成器会走简化路径：

- stage 名称为 `execute`
- workflow 文件为 `workflow-execute.yaml`
- 只保留一个主资源 `Subscription` action
- 主资源 action 使用 `subscription.operation: Apply`
- 不显式生成 `when`
- 不显式生成 delete action；`mode: Delete` 时由 workflow executor 逆序自动推导删除

对应的 `drplan.yaml` 形态如下：

```yaml
stages:
  - name: execute
    description: Simplified workflow for main resources without hooks
    workflows:
      - workflowRef:
          name: <release>-execute
```

### 有 hook：统一 install workflow

当存在任意 `pre/post install/upgrade/delete/rollback` hook 时，生成器回退到 hook-aware 路径：

```yaml
stages:
  - name: install
    description: Unified workflow for install/upgrade hooks and main resources
    workflows:
      - workflowRef:
          name: <release>-install
```

### 单 workflow

无论是否简化，所有 action 都只落在一个 workflow 中。  
其中 hook-aware 模式下，顺序不再是“固定线性列表”，而是由生成器构造成一个 DAG：

1. `pre-*` hook 按 hook 位置和 weight 生成分层 `dependsOn`
2. 主资源聚合 Subscription 依赖最后一层 `pre-*` hook
3. `post-*` hook 再依赖主资源 action

同一位置、同一 weight 的 hook 会落在同一层；install / upgrade 同位置 hook 会生成两个独立 action，各自保留自己的 `when` 和 weight。

这意味着：

- install 路径会执行 install 相关的 `pre-* -> 主资源 -> post-*`
- upgrade 路径会执行 upgrade 相关的 `pre-* -> 主资源 -> post-*`
- `pre-*` 和 `post-*` 仍然是不同 action，不能靠一个 `when` 同时表达前后位置

## Hook 映射规则

### 识别的注解

| Annotation | 作用 |
|---|---|
| `helm.sh/hook` | 识别 hook 类型 |
| `helm.sh/hook-weight` | 同类 hook 排序，升序执行 |
| `helm.sh/hook-delete-policy` | 当前映射为 hook `hookCleanup.beforeCreate/onSuccess/onFailure` |

### 生成规则

- hook 资源统一映射为独立 `Subscription` action
- 每个 hook 资源作为一个独立 subscription 的 feed
- hook action 自动加：
  - `waitReady: true`
  - `clusterExecutionMode: PerCluster`
  - mode 对应的 `when`
  - `hookCleanup`
- 主资源统一聚合到一个 `create-subscription` action
- 主资源 action 默认使用 `subscription.operation: Apply`
- hook action 当前统一使用 `subscription.operation: Create`
- `helm.sh/hook: test` / `test-success` 不生成 action
- 无 hook 简化模式下，主资源 action 不显式携带 `when`

### when 规则

当前 `when` 支持：

```yaml
when: mode == "install"
when: mode == "upgrade"
when: mode == "delete"
when: mode == "rollback"
when: mode == "install" || mode == "upgrade"
```

但 `when` 只负责“是否执行”，不能表达 hook 在 workflow 中的前后位置。  
为了保证串行和 `waitReady` 语义，`pre-install`、`post-install`、`pre-upgrade`、`post-upgrade` 仍需拆成不同 action。

## 示例输出

### 无 hook 时的 `drplan.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: demo-app
  namespace: default
spec:
  description: "Auto-generated from Helm template: demo-app"
  failurePolicy: Stop
  globalParams:
    - name: feedNamespace
      value: default
  stages:
    - name: execute
      description: Simplified workflow for main resources without hooks
      workflows:
        - workflowRef:
            name: demo-app-execute
```

### 无 hook 时的 `workflow-execute.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: demo-app-execute
  namespace: default
spec:
  failurePolicy: FailFast
  parameters:
    - name: feedNamespace
      type: string
      default: default
  actions:
    - name: create-subscription
      type: Subscription
      timeout: 5m
      subscription:
        operation: Apply
        name: demo-app-subscription
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: v1
              kind: ConfigMap
              name: release-name-config
              namespace: $(params.feedNamespace)
            - apiVersion: apps/v1
              kind: Deployment
              name: release-name-server
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}
```

### 有 hook 时的 `drplan.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: demo-app
  namespace: default
spec:
  description: "Auto-generated from Helm template: demo-app"
  failurePolicy: Stop
  globalParams:
    - name: feedNamespace
      value: default
  stages:
    - name: install
      description: Unified workflow for install/upgrade hooks and main resources
      workflows:
        - workflowRef:
            name: demo-app-install
```

### 有 hook 时的 `workflow-install.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: demo-app-install
  namespace: default
spec:
  failurePolicy: FailFast
  parameters:
    - name: feedNamespace
      type: string
      default: default
  actions:
    - name: release-name-db-migrate
      type: Subscription
      timeout: 5m
      waitReady: true
      clusterExecutionMode: PerCluster
      when: mode == "install"
      subscription:
        operation: Create
        name: release-name-db-migrate-sub
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: batch/v1
              kind: Job
              name: release-name-db-migrate
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}

    - name: create-subscription
      type: Subscription
      timeout: 5m
      dependsOn:
        - release-name-db-migrate-pre-install
      subscription:
        operation: Apply
        name: demo-app-subscription
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: v1
              kind: ConfigMap
              name: release-name-config
              namespace: $(params.feedNamespace)
            - apiVersion: apps/v1
              kind: Deployment
              name: release-name-server
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}

    - name: release-name-health-check-post-install
      type: Subscription
      timeout: 5m
      waitReady: true
      clusterExecutionMode: PerCluster
      dependsOn:
        - create-subscription
      subscription:
        operation: Create
        name: release-name-health-check-sub
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: batch/v1
              kind: Job
              name: release-name-health-check
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}
```

### `drplanexecution-install.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: demo-app-install-001
  namespace: default
spec:
  mode: Install
  operationType: Execute
  planRef: demo-app
```

## 生成后的建议调整

1. 补充 `subscribers.clusterAffinity`，指定目标子集群
2. 按实际时长调整 `timeout`
3. 如需集群差异化，下游手工增加 Localization action
4. 根据执行入口选择 install / upgrade / delete / revert 对应 execution 样例

## helmfile 子命令

`drplan-gen helmfile` 用于直接解析一个 helmfile release，并生成基于 `HelmChart + Globalization + Subscription` 的 DRPlan 编排文件。

这个子命令和上面的“渲染 YAML 输入模式”是两条独立链路：

- 普通模式：输入 `helm template` / `helmfile template` 的渲染结果
- `helmfile` 子命令：直接读取 helmfile，并解析 release、values、namespace、chart 信息

### 命令格式

```bash
drplan-gen helmfile \
  -f <helmfile.yaml.gotmpl> \
  -l name=<release-name> \
  --chart-repo <oci://repo-or-http-repo> \
  [-n <namespace>] \
  [--plain-http] \
  [--keep-full-values] \
  [-o <output-dir>]
```

### 参数说明

| 参数 | 短写 | 必填 | 默认值 | 说明 |
|---|---|---|---|---|
| `--file` | `-f` | 是 | | helmfile 文件路径 |
| `--selector` | `-l` | 否 | | release 选择器，例如 `name=bk-cmdb` |
| `--namespace` | `-n` | 否 | | 覆盖 helmfile release namespace |
| `--chart-repo` | | 是 | | 写入 `HelmChart.spec.repo` 的 chart 仓库地址 |
| `--plain-http` | | 否 | `false` | 指定后生成 `HelmChart.spec.plainHTTP: true` |
| `--keep-full-values` | | 否 | `false` | 保留完整 values；默认会和 chart 默认 values 做 diff，只保留差异 |
| `--output` | `-o` | 否 | `.` | 输出目录 |

### 使用示例

```bash
drplan-gen helmfile \
  -f base-blueking.yaml.gotmpl \
  -l name=bk-cmdb \
  -n blueking \
  --chart-repo oci://registry.example.com/charts \
  -o ./drplan-output
```

如果 chart 仓库走 HTTP，可以显式加上：

```bash
drplan-gen helmfile \
  -f base-blueking.yaml.gotmpl \
  -l name=bk-cmdb \
  -n blueking \
  --chart-repo http://registry.example.com/charts \
  --plain-http \
  -o ./drplan-output
```

### 生成结果

helmfile 子命令当前会生成以下文件：

| 文件名 | 说明 |
|---|---|
| `drplan.yaml` | DRPlan 资源，stage 名称为 `execute` |
| `workflow-execute.yaml` | Helmfile 专用 workflow，包含 `HelmChart / Globalization / Subscription` action |
| `drplanexecution-install.yaml` | Install 执行样例，带 `mode: Install` |
| `drplanexecution-upgrade.yaml` | Upgrade 执行样例，带 `mode: Upgrade` |
| `drplanexecution-delete.yaml` | Delete 执行样例，带 `mode: Delete` |
| `drplanexecution-revert.yaml` | Revert 执行样例，带 `mode: Rollback` |

### 生成模型

helmfile 子命令生成的是单个 `execute` stage，对应一个 `workflow-execute.yaml`。默认 action 顺序如下：

1. `HelmChart`：创建或更新 HelmChart 资源
2. `Globalization`：如果存在 values 差异，则生成 values 覆盖
3. `Subscription`：订阅 HelmChart

其中：

- `HelmChart` / `Globalization` / `Subscription` 的安装与升级路径默认使用 `operation: Apply`
- helmfile 子命令默认不生成显式 `rollback`
- helmfile 子命令默认不生成显式 `when`
- helmfile 子命令默认不生成显式 `delete-*` actions
- `mode: Delete` 时，会由 workflow executor 基于正向 action 自动按逆序推导删除动作：
  - `Subscription -> Globalization -> HelmChart`
  - 如果 workflow 中手工写了显式 delete action，则仍按显式定义执行
- `Globalization` 只有在存在非空 values 时才会生成

### Namespace 参数化

为了便于后续调整，helmfile 子命令生成的 workflow 默认会把 namespace 写成参数变量，而不是直接写死：

- `$(params.feedNamespace)`
- `$(params.targetNamespace)`

对应默认值会同时写入：

- `DRPlan.spec.globalParams`
- `DRWorkflow.spec.parameters`

例如：

```yaml
spec:
  globalParams:
    - name: feedNamespace
      value: blueking
    - name: targetNamespace
      value: blueking
```

```yaml
helmChart:
  namespace: $(params.feedNamespace)
  spec:
    targetNamespace: $(params.targetNamespace)
```

### values 处理规则

- 默认行为：将 helmfile release 最终渲染出的 values 与 chart 默认 values 做 diff，只保留差异项写入 `Globalization`
- 指定 `--keep-full-values`：保留完整最终 values
- 当前默认 diff 依赖本地 chart 目录或本地 `.tgz`；如果不能读取 chart 默认 values，可改用 `--keep-full-values`

### Helm 语义映射

- `wait`：优先取 `release.wait`，否则取 `helmDefaults.wait`
- `waitForJob`：优先取 `release.waitForJobs`，否则取 `helmDefaults.waitForJobs`
- `atomic`：优先取 `release.atomic`，否则取 `helmDefaults.atomic`
- 当 `atomic=true` 时，生成器会同时写入：
  - `HelmChart.spec.atomic`
  - `HelmChart.spec.upgradeAtomic`
- 这样可以同时对齐 Helm install 和 Helm upgrade 的原子性语义

## 注意事项

1. 输入必须是已渲染 YAML，工具不会替你执行 `helm template`
2. `waitReady` 依赖 controller 侧能力，生成器只负责写入字段
3. 当前默认策略是“一个 stage + 一个 workflow”，不是多 stage DAG
4. 新增配置推荐使用 `$(params.xxx)` 语法；历史 `{{ .params.xxx }}` 写法继续兼容
5. 为保证 hook 顺序语义，`pre-*` 与 `post-*` 不会合并为同一个 action
