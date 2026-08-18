# Harness Engineering 规范

> 本目录定义了项目的 AI Agent 运行环境规范，是 Agent 理解项目边界、工具能力和行为约束的来源。

## 项目概述

- **项目名称**：bk-bcs
- **技术栈**：Go 1.21+ monorepo；控制面微服务 + K8s Operator；控制台 Vue 2 + webpack；接口以 Protobuf / gRPC-Gateway 为主
- **Agent 适用场景**：单组件功能开发、缺陷修复、需求流水线（TAPD）、代码评审与安全红线检查、Harness 文档维护

## 规范导航

| 组件 | 文档 | 概要 |
|------|------|------|
| 上下文工程 | [context-engineering.md](context-engineering.md) | 知识来源、上下文结构、动态数据接入 |
| 架构约束 | [architectural-constraints.md](architectural-constraints.md) | 分层模型、依赖规则、与工作单元入口的关系 |
| 熵管理 | [entropy-management.md](entropy-management.md) | 文档园艺、技术债追踪、一致性检测 |
| 工具能力 | [tooling.md](tooling.md) | Skill / MCP / CLI 契约清单 |
| 执行与验证 | [execution-verification.md](execution-verification.md) | 执行循环、验证清单、可观测性 |

## 关联文档

| 类型 | 入口 |
|------|------|
| 项目入口 | [`../../AGENTS.md`](../../AGENTS.md) |
| 技术开发规范 | [`../standards/README.md`](../standards/README.md) |
| 业务规范（按需） | [`../business-standards/README.md`](../business-standards/README.md) |
| 词汇表 | [`../glossary.md`](../glossary.md) |
| 架构总览 | [`../overview/README.md`](../overview/README.md) |

## 使用说明

1. Agent 首次接触项目时，先读根 `AGENTS.md`，再读本文件获取全局视图
2. 修改某路径前，阅读向上最近的工作单元 `AGENTS.md`（局部优先）
3. 执行具体任务时，按需深入对应组件文档与 `docs/standards/` 相关章节
4. 规范更新后需同步检查关联组件的一致性（可触发「文档巡检」）

## 版本记录

| 版本 | 日期 | 变更说明 |
|------|------|---------|
| 1.0.0 | 2026-08-17 | 仓库根首次生成五大组件 + standards + 业务规范空间 |
