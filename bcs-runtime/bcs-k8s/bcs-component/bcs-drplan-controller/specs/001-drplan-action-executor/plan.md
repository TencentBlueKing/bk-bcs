# Implementation Plan: 容灾策略 CR 及动作执行器

**Branch**: `001-drplan-action-executor` | **Date**: 2026-01-30 | **Updated**: 2026-02-02 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-drplan-action-executor/spec.md`

## Summary

实现一个 Kubernetes Operator，提供三层 CRD（DRWorkflow、DRPlan、DRPlanExecution）来定义和执行容灾策略。支持 HTTP、Job、Localization、Subscription、KubernetesResource 五种动作类型，具备参数化工作流、Stage 编排、自动/自定义回滚、手动触发等能力。

**更新（2026-02-02）**：基于澄清会话新增性能、可靠性和可观测性要求。

## Technical Context

**Language/Version**: Go 1.21+（遵循 BCS 项目要求）  
**Primary Dependencies**: 
- controller-runtime v0.16+（Reconcile 框架）
- kubebuilder v3.x（脚手架生成）
- client-go（Kubernetes API 客户端）
- clusternet/apis（Localization/Subscription CRD，更新 2026-02-02）
- logr（结构化日志接口，更新 2026-02-02）

**Storage**: Kubernetes API Server（etcd），无额外存储  
**Testing**: 
- go test（单元测试）
- envtest（controller-runtime 集成测试框架）
- ginkgo/gomega（BDD 风格测试）
- gomock（executor 接口 mock，更新 2026-02-02）

**Target Platform**: Kubernetes 1.20+，部署在 Hub 集群  
**Project Type**: Kubernetes Operator（单一项目）  
**Performance Goals**: 
- 触发执行后 10 秒内开始第一个动作
- 状态变更 5 秒内反映到 Status
- 支持单工作流 20+ 动作
- **Reconcile 频率可配置，默认 30 秒**（新增 2026-02-02）

**Constraints**: 
- Controller 无状态，支持多副本 Leader Election
- 所有动作幂等，支持安全重试
- Controller 重启后 30 秒内恢复执行
- **网络分区超过 2 分钟标记为 Unknown 状态**（新增 2026-02-02）
- **Stage 并行执行采用 FailFast 策略**（新增 2026-02-02）
- **Clusternet 动作采用异步模型（CR 创建成功即完成）**（新增 2026-02-02）

**Observability Requirements**（新增 2026-02-02）:
- 使用结构化 JSON 日志格式
- Info 级别：关键事件（执行/Stage/Workflow/动作的开始/成功/失败）
- Debug 级别：参数替换详情、动作执行详情（HTTP 请求体、Job manifest 等）
- 所有操作通过 Kubernetes Events 记录（关联到 DRPlanExecution）

**Scale/Scope**: 
- 单集群 500 DRPlan（更新 2026-02-02）
- 单集群 5000 DRPlanExecution（历史记录，更新 2026-02-02）
- 单个 DRWorkflow 最多 20 个 Action

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Evidence |
|-----------|--------|----------|
| **I. Kubernetes Operator 模式** | ✅ PASS | 三层 CRD 设计，Reconcile 驱动，Controller 无状态 |
| **II. 故障恢复可靠性** | ✅ PASS | 幂等动作、断点恢复、超时重试、参数快照 |
| **III. 可观测性优先** | ✅ PASS | Events、结构化日志、Status 实时更新 |
| **IV. 测试覆盖** | ✅ PASS | 单元测试 + envtest 集成测试规划 |
| **V. 渐进式执行** | ✅ PASS | 分步执行、手动触发、回滚机制 |

**Technical Constraints**:
- ✅ Go 语言 + controller-runtime
- ✅ kubebuilder 脚手架
- ✅ CRD apiextensions.k8s.io/v1
- ✅ 遵循 BCS .golangci.yml

## Project Structure

### Documentation (this feature)

```text
specs/001-drplan-action-executor/
├── spec.md              # 功能规格说明
├── plan.md              # 本文件（实现计划）
├── research.md          # 技术研究（Phase 0）
├── data-model.md        # 数据模型（Phase 1）
├── quickstart.md        # 快速开始指南
├── contracts/           # CRD API 定义
│   ├── drworkflow.yaml
│   ├── drplan.yaml
│   └── drplanexecution.yaml
└── checklists/
    └── requirements.md  # 需求检查清单
```

### Source Code (repository root)

```text
bcs-drplan-controller/
├── api/
│   └── v1alpha1/                    # CRD 类型定义
│       ├── drworkflow_types.go
│       ├── drplan_types.go
│       ├── drplanexecution_types.go
│       ├── groupversion_info.go
│       └── zz_generated.deepcopy.go
│
├── internal/
│   ├── controller/                  # Reconciler 实现
│   │   ├── drworkflow_controller.go
│   │   ├── drplan_controller.go
│   │   ├── drplanexecution_controller.go
│   │   └── suite_test.go
│   │
│   ├── executor/                    # 动作执行器
│   │   ├── executor.go              # 执行器接口
│   │   ├── http_executor.go         # HTTP 动作
│   │   ├── job_executor.go          # Job 动作
│   │   ├── localization_executor.go # Localization 动作
│   │   ├── subscription_executor.go # Subscription 动作（新增 2026-02-02）
│   │   ├── k8s_resource_executor.go # KubernetesResource 动作（新增 2026-02-02）
│   │   ├── stage_executor.go        # Stage 编排执行（新增 2026-02-02）
│   │   └── rollback.go              # 回滚逻辑
│   │
│   ├── webhook/                     # Webhook 验证
│   │   ├── drworkflow_webhook.go
│   │   ├── drplan_webhook.go
│   │   └── drplanexecution_webhook.go
│   │
│   └── utils/                       # 工具函数
│       ├── template.go              # 参数模板替换
│       └── retry.go                 # 重试逻辑
│
├── config/
│   ├── crd/                         # CRD YAML
│   ├── rbac/                        # RBAC 配置
│   ├── manager/                     # Controller 部署
│   └── webhook/                     # Webhook 配置
│
├── cmd/
│   └── main.go                      # 入口
│
├── tests/
│   ├── e2e/                         # 端到端测试
│   └── integration/                 # 集成测试
│
├── Dockerfile
├── Makefile
├── go.mod
└── go.sum
```

**Structure Decision**: 采用标准 kubebuilder 项目结构，executor 包独立封装动作执行逻辑，便于扩展新动作类型。

## Complexity Tracking

> **无违规需要说明**

Constitution Check 全部通过，设计符合所有核心原则。

---

## Phase 0: Research Summary

详见 [research.md](./research.md)

## Phase 1: Design Artifacts

- **数据模型**: [data-model.md](./data-model.md)
- **CRD 合约**: [contracts/](./contracts/)
- **快速开始**: [quickstart.md](./quickstart.md)
