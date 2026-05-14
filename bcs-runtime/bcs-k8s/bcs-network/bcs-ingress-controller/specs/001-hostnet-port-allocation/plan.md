# Implementation Plan: HostNetwork 动态端口分配

**Branch**: `001-hostnet-port-allocation` | **Date**: 2026-04-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-hostnet-port-allocation/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

本功能为 bcs-ingress-controller 添加 HostNetPortPool 子系统，为使用 `hostNetwork` 模式的 Pod 提供动态端口段分配能力。Controller 通过监听 Pod、Node 和 HostNetPortPool CR 的事件，在 Pod 调度完成后动态分配端口段，并将结果注入 Pod annotation。关键架构决策：采用三个独立 Reconciler 分别处理不同资源类型，使用 Cache 而非 Annotation 进行幂等性检查。

## Technical Context

**Language/Version**: Go 1.20+  
**Primary Dependencies**: controller-runtime v0.6.3, kubebuilder, prometheus client  
**Storage**: Kubernetes etcd (CRD 存储), 内存 Cache (端口分配状态)  
**Testing**: go test (table-driven tests with fake client)  
**Target Platform**: Linux (Kubernetes cluster)
**Project Type**: Kubernetes Controller/Operator  
**Performance Goals**: 单次 Reconcile 秒级完成，支持并发 Pod 调度到同一 Node  
**Constraints**: 端口段 per-Node 隔离，Controller 重启后缓存重建不丢失状态  
**Scale/Scope**: 几十到几百 Node，每 Node 数十到数百端口段

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原则 | 状态 | 备注 |
|------|------|------|
| I. Go 语言工程规范 | PASS | 项目使用标准 Go 1.20，遵循 gofmt/goimports |
| II. Kubernetes 控制器开发规范 | PASS | 使用 controller-runtime，需拆分为三个独立 Reconciler |
| III. 代码质量 | PASS | 需确保拆分后单个 Reconciler 圈复杂度 < 15 |
| IV. 英文代码注释 | PASS | 所有导出类型/函数添加英文 GoDoc 注释 |
| V. 中文沟通语言 | PASS | 计划文档使用中文 |

**架构决策要点**（基于 Clarifications Session 2026-04-09）：
1. **Controller 拆分**: 使用三个独立 Reconciler（`HostNetPortPoolReconciler`、`PodReconciler`、`NodeReconciler`），符合 Constitution II 控制器规范，避免使用 `__node__` 特殊 Namespace 判断事件类型
2. **幂等性检查**: 使用 Cache 查询而非 Pod Annotation，避免 APIServer 延迟导致的竞态问题
3. **错误处理**: 非法 portCount 时直接报错（非默认值），显式处理错误并记录 Warning Event
4. **指标暴露**: 新增 Counter 指标 `hostnet_pool_shrink_conflict_total` 记录端口池缩小冲突

## Project Structure

### Documentation (this feature)

```text
specs/001-hostnet-port-allocation/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
bcs-ingress-controller/
├── hostnetportcontroller/           # HostNetPortPool 控制器（需要重构）
│   ├── pool_controller.go           # HostNetPortPoolReconciler - 处理 CRD 生命周期
│   ├── pod_controller.go            # PodReconciler - 处理 Pod 端口分配
│   ├── node_controller.go           # NodeReconciler - 处理 Node 缓存清理
│   └── *_test.go                    # 单元测试（表驱动，fake client）
├── internal/
│   ├── hostnetportpoolcache/        # 内存缓存层
│   │   ├── cache.go                 # 端口段分配管理
│   │   └── types.go                 # 类型定义
│   ├── metrics/
│   │   └── hostnetportpool.go       # Prometheus 指标（新增 shrink conflict counter）
│   ├── constant/
│   │   └── constant.go              # 常量定义（annotation keys, finalizer）
│   └── httpsvr/
│       └── hostnetportpool.go       # HTTP API - 查询分配结果
└── main.go                          # 注册三个 Reconciler
```

**Structure Decision**: 采用三个独立 Reconciler 结构，替代现有的单一 `controller.go` + `__node__` Namespace 前缀判断模式。每个 Reconciler 独立注册到 Manager，分别处理 HostNetPortPool CR、Pod、Node 资源事件，符合 controller-runtime 最佳实践和 Constitution II 规范。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 三个独立 Reconciler | 代码文件数量增加（从 1 个 controller.go 变为 3 个 controller 文件） | 保持单一 Reconciler 使用 `__node__` 前缀判断事件类型——拒绝原因：不符合 controller-runtime 最佳实践，Namespace 前缀魔术值难以维护，单一 Reconciler 职责不清晰 |
| Cache 查询幂等性 | 新增 Cache API 方法 `IsPodAllocated`，增加 Cache 层复杂度 | 直接读取 Pod Annotation 判断——拒绝原因：APIServer 压力大时 Annotation 更新延迟，可能导致竞态条件和重复分配 |
| 非法 portCount 报错 | 增加错误处理分支，需要记录 Event 和更新 status | 使用默认值 1 segment——拒绝原因：静默容错掩盖用户配置错误，不符合 Constitution "错误 MUST 被显式处理" 原则 |
| 端口池缩小 Counter | 新增指标类型，需定义 label 维度 | 仅记录 Event——拒绝原因：Event 不适合告警，Counter 指标可用于 PromQL 告警规则，提升可观测性 |
