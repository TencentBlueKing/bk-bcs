# drplan-gen 使用指南

## 概述

`drplan-gen` 用于将 `helm template` / `helmfile template` 的渲染结果转换为 DRPlan 编排文件（`DRPlan + DRWorkflow + DRPlanExecution`）。

当前默认生成标准为**单 workflow 模式**：

- 默认只生成一个 stage：`install`
- 默认只生成一个 workflow：`workflow-install.yaml`
- hook action 与主资源 action 都放在同一个 workflow 中
- 通过 `when + waitReady` 模拟 Helm install/upgrade hook 语义

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
| `drplan.yaml` | DRPlan 资源，默认只包含一个 `install` stage |
| `workflow-install.yaml` | 统一 DRWorkflow，包含 hooks + main resources |
| `drplanexecution-install.yaml` | Install 执行样例，带 `mode: Install` |
| `drplanexecution-upgrade.yaml` | Upgrade 执行样例，带 `mode: Upgrade` |
| `drplanexecution-delete.yaml` | Delete 执行样例，带 `mode: Delete` |
| `drplanexecution-revert.yaml` | Revert 样例，带 `mode: Rollback` |

## 默认生成模型

### 单 stage

生成器默认只生成一个 stage：

```yaml
stages:
  - name: install
    description: Unified workflow for install/upgrade hooks and main resources
    workflows:
      - workflowRef:
          name: <release>-install
```

### 单 workflow

所有 action 都落在一个 workflow 中，但顺序不再是“固定线性列表”，而是由生成器构造成一个 DAG：

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
  - install / upgrade 路径对应的 `when`
  - `hookCleanup`
- 主资源统一聚合到一个 `create-subscription` action
- 主资源 action 默认使用 `subscription.operation: Apply`
- hook action 当前统一使用 `subscription.operation: Create`
- `helm.sh/hook: test` / `test-success` 不生成 action

### when 规则

当前 `when` 支持：

```yaml
when: mode == "install"
when: mode == "upgrade"
when: mode == "install" || mode == "upgrade"
```

但 `when` 只负责“是否执行”，不能表达 hook 在 workflow 中的前后位置。  
为了保证串行和 `waitReady` 语义，`pre-install`、`post-install`、`pre-upgrade`、`post-upgrade` 仍需拆成不同 action。

## 示例输出

### `drplan.yaml`

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

### `workflow-install.yaml`

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

## 注意事项

1. 输入必须是已渲染 YAML，工具不会替你执行 `helm template`
2. `waitReady` 依赖 controller 侧能力，生成器只负责写入字段
3. 当前默认策略是“一个 stage + 一个 workflow”，不是多 stage DAG
4. 新增配置推荐使用 `$(params.xxx)` 语法；历史 `{{ .params.xxx }}` 写法继续兼容
5. 为保证 hook 顺序语义，`pre-*` 与 `post-*` 不会合并为同一个 action
