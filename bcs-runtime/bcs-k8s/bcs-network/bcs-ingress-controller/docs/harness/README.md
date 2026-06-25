# Harness Engineering 规范

> 本目录定义 BCS Ingress Controller 的 AI Agent 运行环境规范，是 Agent 理解项目边界、工具能力和行为约束的权威来源。

## 项目概述

- **项目名称**：bcs-ingress-controller
- **技术栈**：Go 1.20+、controller-runtime v0.6.3、go-restful、Prometheus、多云 LB SDK
- **Agent 适用场景**：K8s Operator 功能开发、CRD 控制器扩展、云适配器修改、HTTP API 新增、Webhook 逻辑、单元测试与文档维护

## 规范导航

| 组件 | 文档 | 概要 |
|------|------|------|
| 上下文工程 | [context-engineering.md](context-engineering.md) | 知识来源、渐进式披露、动态数据接入 |
| 架构约束 | [architectural-constraints.md](architectural-constraints.md) | 分层模型、依赖规则、Controller 模式 |
| 熵管理 | [entropy-management.md](entropy-management.md) | 文档园艺、技术债追踪、一致性检测 |
| 工具能力 | [tooling.md](tooling.md) | Skill/MCP/CLI 清单、环境状态、稳定性策略 |
| 执行与验证 | [execution-verification.md](execution-verification.md) | Agent Loop、预完成检查、可观测性 |

## 关联文档

| 类型 | 入口 |
|------|------|
| 技术开发规范 | [../standards/README.md](../standards/README.md) |
| 开发地图 | [../dev-map/README.md](../dev-map/README.md) |
| 词汇表 | [../glossary.md](../glossary.md) |
| 项目入口 | [../../AGENTS.md](../../AGENTS.md) |

## 使用说明

1. Agent 首次接触项目时，先读 `AGENTS.md` 获取全局视图
2. 执行具体任务前，按需深入本目录对应组件文档
3. 涉及代码实现时，同步加载 `docs/standards/` 中相关技术规范
4. 评估变更影响时，查阅 `docs/dev-map/` 模块与依赖索引
5. 规范更新后，运行「文档巡检」检查关联组件一致性

## 版本记录

| 版本 | 日期 | 变更说明 |
|------|------|---------|
| 1.0.0 | 2026-06-08 | 初始版本，覆盖五大组件 + dev map + standards |
