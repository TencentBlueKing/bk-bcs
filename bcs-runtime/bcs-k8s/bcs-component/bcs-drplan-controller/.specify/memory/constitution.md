<!--
Sync Impact Report
==================
Version Change: N/A → 1.0.0 (Initial)
Bump Rationale: MAJOR - Initial constitution ratification

Modified Principles: N/A (Initial creation)
Added Sections:
  - Core Principles (5 principles)
  - Technical Constraints
  - Development Workflow
  - Governance

Templates Status:
  - .specify/templates/plan-template.md: ✅ Compatible
  - .specify/templates/spec-template.md: ✅ Compatible
  - .specify/templates/tasks-template.md: ✅ Compatible

Follow-up TODOs: None
-->

# bcs-drplan-controller Constitution

## Core Principles

### I. Kubernetes Operator 模式

遵循 Kubernetes Operator 最佳实践：

- 所有容灾策略必须通过 CRD（Custom Resource Definition）定义
- 使用 Reconcile 循环驱动状态收敛，确保声明式配置
- Controller 必须是无状态的，所有状态存储在 Kubernetes API Server
- 支持多副本部署，通过 Leader Election 保证单一活跃实例
- 必须正确处理 Finalizer，确保资源清理完整

### II. 故障恢复可靠性

容灾动作的可靠性是核心要求：

- 所有容灾动作必须是幂等的，支持安全重试
- 支持部分失败后的继续执行，避免从头开始
- 必须记录执行进度和状态，支持断点恢复
- 关键操作必须有超时控制和失败重试策略
- 确保最终一致性，避免中间状态长期存在

### III. 可观测性优先

所有操作必须可追踪和可审计：

- 容灾动作必须发送 Kubernetes Events，记录关键状态变更
- 使用结构化日志（JSON 格式），包含请求 ID、资源标识等上下文
- 暴露 Prometheus 指标：执行计数、延迟、成功/失败率
- Status 字段必须反映当前状态和最近错误信息
- 支持通过 kubectl describe 快速定位问题

### IV. 测试覆盖

关键功能必须有测试保障：

- 核心业务逻辑必须有单元测试，覆盖正常和异常路径
- Reconcile 逻辑必须有集成测试，使用 envtest 框架
- 容灾场景需要模拟测试（故障注入）验证
- Webhook 验证逻辑必须有测试覆盖
- 测试代码与生产代码同等重要，需要代码审查

### V. 渐进式执行

容灾计划支持灵活的执行策略：

- 支持分阶段（Phase）执行，每个阶段可独立验证
- 支持人工审批节点，关键操作需要确认
- 提供回滚机制，支持撤销已执行的动作
- 支持模拟执行（Dry-run），预览变更而不实际执行
- 执行速率可控，避免雪崩效应

## Technical Constraints

**语言与框架**：
- 使用 Go 语言开发，遵循 BCS 项目 Go 版本要求
- 基于 controller-runtime 框架构建
- 使用 kubebuilder 生成脚手架代码

**兼容性要求**：
- 支持 Kubernetes 1.20+ 版本
- CRD 使用 apiextensions.k8s.io/v1
- 与 BCS 其他组件 API 兼容

**代码规范**：
- 遵循 BCS 项目 .golangci.yml 配置
- 代码必须通过 go vet、golangci-lint 检查
- 导出的类型和函数必须有 godoc 注释

**依赖管理**：
- 使用 Go Modules 管理依赖
- 优先使用 BCS 公共库（bcs-common）
- 第三方依赖需要评估许可证兼容性

## Development Workflow

**代码审查**：
- 所有代码变更必须通过 Pull Request
- 至少一名核心成员 Approve 后方可合并
- CI 检查必须全部通过

**版本管理**：
- 遵循语义化版本（Semantic Versioning）
- Breaking Change 必须在 CHANGELOG 中明确标注
- CRD 版本变更需要提供迁移方案

**部署策略**：
- 支持 Helm Chart 部署
- 提供完整的 RBAC 配置
- 支持配置热加载，避免频繁重启

**文档要求**：
- API 变更必须更新对应文档
- 提供运维手册和故障排查指南
- 重要设计决策需要记录 ADR（Architecture Decision Record）

## Governance

- 本 Constitution 是项目开发的最高准则，所有 PR 和代码审查必须验证合规性
- 修订 Constitution 需要：明确的修改理由、团队讨论、文档化的迁移计划
- 复杂性必须有正当理由：如果简单方案能解决问题，不应引入复杂设计
- 运行时开发指导参考项目 README 和 docs/ 目录

**Version**: 1.0.0 | **Ratified**: 2026-01-30 | **Last Amended**: 2026-01-30
