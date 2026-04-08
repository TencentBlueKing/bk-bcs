## Why

当前用户要将一个 Helm chart 部署为 DRPlan 编排时，需要手动编写 DRPlan、DRWorkflow、DRPlanExecution 等多个 YAML 文件。这个过程重复、易出错，且要求用户深入理解 Helm Hook annotation 到 DRPlan Stage/Action 的映射关系。

需要一个 CLI 工具自动化这个转换过程：输入 `helm template` / `helmfile template` 渲染后的 YAML，输出完整的 DRPlan 编排文件。

## What Changes

新增独立 CLI 工具 `drplan-gen`：

- **输入**：`helm template` 渲染输出的 YAML（支持 stdin 管道和文件）
- **解析**：自动识别 Helm Hook annotation（`helm.sh/hook`、`helm.sh/hook-weight`、`helm.sh/hook-delete-policy`）
- **分类**：将资源分为 Hook 资源（pre-install/post-install 等）和主资源
- **生成**：
  - DRPlan：Stage DAG 编排（pre-install → install → post-install）
  - DRWorkflow：每个 Stage 对应的 Workflow（Hook → 独立 Subscription Action，主资源 → 聚合 Subscription feeds）
  - DRPlanExecution：Execute 和 Revert 样例
- **CLI 参数**：`--name`、`--namespace`、`-f`（输入文件）、`-o`（输出目录）、`--wait`（生成 waitReady: true）

新增文件：
- `cmd/drplan-gen/main.go`：CLI 入口
- `internal/generator/parser.go`：YAML 多文档解析
- `internal/generator/classifier.go`：资源分类（Hook vs 主资源）
- `internal/generator/planner.go`：DRPlan/DRWorkflow 生成
- `internal/generator/writer.go`：YAML 文件输出
- `internal/generator/types.go`：内部数据结构
- Makefile 新增 `build-gen` target

## Impact

- **新增代码**：约 800-1000 行 Go 代码 + 测试
- **不影响现有 controller**：工具是独立二进制，与 controller manager 无耦合
- **不引入新外部依赖**：仅使用项目已有的 `unstructured`、`sigs.k8s.io/yaml`、`cobra`
- **API 类型**：只读引用 `api/v1alpha1/` 中的类型定义用于生成 YAML，不修改
