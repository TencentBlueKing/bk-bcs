## Context

bcs-drplan-controller 项目当前只有一个 controller manager 二进制（`cmd/main.go`），用户需要手动编写 DRPlan/DRWorkflow YAML。将 Helm chart 转为 DRPlan 编排需要深入理解 Hook annotation 到 Stage/Action 的映射关系，过程繁琐。

项目中 `helm.sh/helm/v3 v3.15.2` 作为 Clusternet 的传递依赖已存在于 go.mod 中，但不在业务代码中直接使用。`cobra` 和 `sigs.k8s.io/yaml` 等库也已作为依赖可用。

## Goals / Non-Goals

**Goals:**
- 提供 `drplan-gen` CLI 工具，自动从渲染后的 YAML 生成 DRPlan 编排文件
- 正确识别 Helm Hook annotation 并映射到 DRPlan Stage DAG
- 支持 stdin 管道和文件输入，兼容 `helm template` 和 `helmfile template` 输出
- 零新外部依赖，仅使用项目已有的库

**Non-Goals:**
- 不集成 Helm SDK 做 chart 渲染（用户自行运行 `helm template`）
- 不处理多集群差异化（Localization），用户后续手动添加
- 不处理 Helmfile 多 release 区分（聚焦单 release）
- 不实现 `waitReady` 运行时机制（仅在 YAML 中标记字段）

## Decisions

### D1: 采用方案 B - YAML 解析而非 Helm SDK

解析 `helm template` 渲染后的 YAML 而非直接集成 Helm SDK。

理由：
- 核心价值在于"资源到 DRPlan 的映射"而非"chart 渲染"
- 不引入重量级依赖，二进制体积不增长
- 更通用，可处理 helmfile template、kustomize build 甚至手写 YAML
- 后续如需 chart 直接输入，可加一层渲染前处理

### D2: 代码放在 internal/generator/ 而非 pkg/

遵循项目现有惯例（controller/executor 等都在 internal/），不暴露为公共 API。

### D3: Hook 资源通过独立 Subscription Action 分发（实现偏差）

~~原设计：Hook 资源映射为 Job Action。~~
**实际实现**：Hook 资源也通过 Subscription Action 分发到子集群，每个 hook 资源作为独立 Subscription 的 feed。

变更理由：在 Clusternet 多集群架构下，Job Action 会在管理集群直接创建 Job，但应用级别的 hook Job（如 db-migrate）需要在子集群执行。改为 Subscription 模式与主资源分发方式保持一致，确保 hook Job 能正确分发到目标子集群。

每个 hook 资源生成独立的 Subscription Action（而非聚合），Workflow 内的 Action 顺序执行自然保持了 hook-weight 排序语义。

### D4: 主资源通过 Subscription feeds 分发

非 Hook 资源聚合到一个 Subscription 的 feeds 列表中，不为每个资源单独创建 Action。这与 waitready-design.md 中的建议一致。

### D5: 默认不设置 waitReady

`waitReady` 默认为 false（与现有 Action 行为一致）。用户通过 `--wait` CLI 参数显式开启。这在 controller 尚未实现 waitReady 机制时也不会造成问题（字段被忽略）。

## Risks / Trade-offs

1. **Hook Kind 限制**：Helm Hook 通常是 Job/Pod，但工具对 Kind 无限制 — 所有带 hook annotation 的资源都会作为 Subscription feed 分发。非 Job 类 hook 在子集群的行为由 Kubernetes 本身决定。

2. **CRD 资源**：chart 中可能包含 CRD，CRD 需要在其他资源之前部署。v1 不做特殊处理，将 CRD 与其他主资源一起放入 Subscription feeds。

3. **waitReady 尚未实现**：生成的 YAML 中可能包含 `waitReady: true`，但 controller 尚未支持此字段。API 层面此字段为 optional，不影响 CRD 验证。

4. **Subscription spec 不完整**：生成的 Subscription 只有 feeds 和 schedulingStrategy，`subscribers.clusterAffinity` 为空，用户需要后续补充集群选择器。
