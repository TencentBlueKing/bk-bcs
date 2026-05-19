# drplan-gen Helmfile 模式设计

## 目标

新增一个 `drplan-gen` 生成模式，直接基于 `helmfile` 配置解析单个 release 的最终部署信息，生成一套可执行的 `DRPlan + DRWorkflow + DRPlanExecution`。

该模式不再以 workload rendered YAML 作为输入，也不再单独输出 `HelmChart`、`Globalization`、`Subscription` 资源文件，而是将这三类对象以内联 action spec 的方式嵌入生成的 workflow 中。

## 范围

首版范围：

- 输入来源为 `helmfile`，解析方式使用 helmfile Go 包，不调用 `helmfile` CLI
- 只支持单个 release
- 只生成一个公共 `Globalization`，不生成 `Localization`
- `HelmChart.spec.repo` 由用户通过 CLI 显式传入
- workflow 中支持专用 `HelmChart` action，并与 `Globalization`、`Subscription` action 一起编排

首版不做：

- 多 release 聚合生成
- `Localization` 自动拆分
- 依赖 rendered workload YAML 的 hook/workload feed 生成
- 额外输出 `helmchart.yaml`、`globalization.yaml`、`subscription.yaml`

## CLI 设计

保留现有模式不变：

```bash
drplan-gen --name my-app --namespace default -f rendered.yaml
```

新增子命令：

```bash
drplan-gen helmfile -f base-blueking.yaml.gotmpl -l name=bk-paas -n blueking --chart-repo oci://10.0.208.12:5000/charts
```

参数：

- `-f, --file`
  - helmfile 文件路径，必填
- `-l, --selector`
  - helmfile label selector，可重复
- `-n, --namespace`
  - 目标 namespace，同时作为 helmfile release 选择条件和默认 release namespace
- `--chart-repo`
  - 必填，写入 `HelmChart.spec.repo`
- `--plain-http`
  - 可选，写入 `HelmChart.spec.plainHTTP`
- `-o, --output`
  - 输出目录，保留现有语义

## 生成模型

### 输出文件

新模式只输出：

- `drplan.yaml`
- `workflow-install.yaml`
- `drplanexecution-execute.yaml`
- `drplanexecution-revert.yaml`

不输出：

- `helmchart.yaml`
- `globalization.yaml`
- `subscription.yaml`

### Workflow 结构

当 release 有非空最终 values 时，生成的 workflow 固定包含 3 个主 action：

1. `apply-helmchart`
2. `apply-globalization`
3. `apply-subscription`

依赖关系：

- `apply-globalization` depends on `apply-helmchart`
- `apply-subscription` depends on `apply-globalization`

当最终 values 为空时：

- 省略 `apply-globalization`
- `apply-subscription` 直接 depends on `apply-helmchart`

删除链路固定为逆序：

1. `delete-subscription`
2. `delete-globalization`（如果存在）
3. `delete-helmchart`

## Helmfile 解析

### Release 选择

通过 helmfile Go 包加载 state，并应用 `-f/-l/-n`。

解析结果要求：

- 命中 0 个 release：报错
- 命中多于 1 个 release：报错
- 命中 1 个 release：继续生成

### 必要的 Release 信息

从最终选中的 release 提取：

- `release.name`
- `release.namespace`
- `release.chart`
- `release.version`
- release 对应的最终 merge 后 values
- 以及可映射到 `HelmChartSpec` 的 helm 选项：
  - `wait`
  - `waitForJobs`
  - `atomic`
  - `createNamespace`
  - `timeout`

### Chart 名称提取

对于渲染后的 `release.chart`，首版采用如下规则生成 `HelmChart.spec.chart`：

1. 取最后一个路径段
2. 去掉 `.tgz` 扩展名
3. 如果尾部是 `-<release.version>`，剥离该后缀
4. 如果无法剥离，则退化为“去扩展名后的 basename”

示例：

```text
./charts/bcs-services-stack-1.2.3.tgz + version=1.2.3
-> bcs-services-stack
```

## 资源映射

### HelmChart Action

新增专用 `HelmChart` action 类型，供 workflow 使用。

生成时默认使用：

- `type: HelmChart`
- `operation: Apply`

字段映射：

- `metadata.name` <- `release.name`
- `metadata.namespace` <- `release.namespace`
- `spec.repo` <- `--chart-repo`
- `spec.chart` <- 由 `release.chart` 解析出的 chart 名称
- `spec.version` <- `release.version`
- `spec.targetNamespace` <- `release.namespace`
- `spec.wait` <- release/helmDefaults 合并结果
- `spec.waitForJob` <- release/helmDefaults 合并结果
- `spec.atomic` <- release/helmDefaults 合并结果
- `spec.createNamespace` <- release/helmDefaults 合并结果
- `spec.timeoutSeconds` <- release/helmDefaults 合并结果
- `spec.plainHTTP` <- `--plain-http`

### Globalization Action

生成时默认使用：

- `type: Globalization`
- `operation: Apply`

字段规则：

- `name` <- `release.name`
- `feed` 固定指向同一个 `HelmChart`
- `priority` 固定使用 `600`
- `overridePolicy` 固定使用 `ApplyNow`
- `overrides[0].type` 固定使用 `Helm`
- `overrides[0].value` <- 最终 merge 后 values YAML

当 values 为空时，不生成该 action。

### Subscription Action

生成时默认使用：

- `type: Subscription`
- `operation: Apply`

字段规则：

- `name` <- `release.name + "-subscription"`
- `namespace` <- `release.namespace`
- `feeds` 只引用同一个 `HelmChart`
- `subscribers` 首版固定为 `clusterAffinity: {}`
- `schedulingStrategy` 固定为 `Replication`

首版默认不自动生成 `waitReady: true`，避免把 `HelmChart` feed 误判为 workload-ready。

## DRPlan 结构

新模式生成的 `DRPlan` 保持现有单 stage / 单 workflow 结构，不引入新的 stage 组织方式。

- `DRPlan.metadata.name` <- `release.name`
- `DRWorkflow.metadata.name` <- `release.name + "-install"`
- `DRPlanExecution` 样例仍输出 `execute` 和 `revert`

## 失败场景

以下情况直接报错：

- helmfile 文件加载失败
- selector/namespace 过滤后没有 release
- selector/namespace 过滤后多于一个 release
- `--chart-repo` 未提供
- release 缺少 name / namespace / chart / version 等关键字段
- 最终 chart 名无法从 `release.chart` 提取出非空值

以下情况允许生成，但会改变输出：

- 最终 values 为空：省略 `Globalization` action

## 代码落点

### drplan-gen CLI

- `cmd/drplan-gen/main.go`
  - 增加 `helmfile` 子命令
  - 复用现有输出目录和版本信息逻辑

### Generator

新增一组 helmfile 模式专用文件，避免污染现有 rendered-YAML 路径：

- `internal/generator/helmfile_loader.go`
  - 封装 helmfile Go 包加载与单 release 选择
- `internal/generator/helmfile_types.go`
  - 定义解析后的 release 中间结构
- `internal/generator/helmfile_planner.go`
  - 生成 `DRPlan/DRWorkflow/DRPlanExecution`
- `internal/generator/helmfile_planner_test.go`
  - 覆盖主要规划逻辑

现有 `internal/generator/parser.go`、`classifier.go`、`planner.go` 保持用于 workload-YAML 模式，不强行混用。

### Controller / API

新增 `HelmChart` action 全链路：

- `api/v1alpha1`
  - action type 常量
  - `HelmChartAction`
  - `ActionOutputs.HelmChartRef`
- `internal/executor`
  - `helmchart_executor.go`
  - `helmchart_executor_test.go`
- `internal/webhook`
  - default / validate
- RBAC / manifests
  - 增加 `helmcharts` 资源权限

## 测试策略

### drplan-gen

- `helmfile` 子命令参数校验
- 0 / 1 / 多 release 选择校验
- chart basename 提取
- values 为空时省略 `Globalization`
- values 非空时输出 `HelmChart -> Globalization -> Subscription`
- 删除链路逆序正确

### Controller

- `HelmChart` action executor：
  - `Create`
  - `Apply`
  - `Patch`
  - `Delete`
  - rollback
- webhook：
  - 默认值
  - `Apply/Patch` 需要 rollback
  - `Delete` 不要求 spec

## 实施顺序

1. 新增 `HelmChart` action 类型、executor、webhook、RBAC、测试
2. 新增 `drplan-gen helmfile` 子命令和参数定义
3. 接入 helmfile Go 包，完成单 release 解析
4. 实现 helmfile planner，生成新的 workflow 结构
5. 补充单测与集成验证
