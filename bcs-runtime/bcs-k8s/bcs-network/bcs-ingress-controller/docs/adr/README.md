# 架构决策记录（ADR）

> 记录 BCS Ingress Controller 的重要架构决策及其背景，供 Agent 和开发者在做类似决策时参考。

## 格式

- 命名：`NNNN-标题.md`（四位序号 + 短标题）
- 模板：背景 → 决策 → 后果（正面/负面）
- 状态：`已接受` / `已废弃` / `已替代`

## 索引

| 编号 | 标题 | 状态 | 日期 |
|------|------|------|------|
| [0001](0001-namespace-scope-exemption.md) | Namespace Scope 豁免机制 | 已接受 | 2026 |
| [0002](0002-hostnet-port-pool-allocation.md) | HostNetPortPool 动态端口分配 | 已接受 | 2026 |

## 何时新增 ADR

- 引入新的跨模块数据流或全局配置行为
- 选择某种云适配模式或缓存策略
- 对现有约束做有意的例外（如豁免名单）
- 技术栈大版本升级（如 controller-runtime 迁移）

功能设计细节仍放 `specs/{feature-id}/`；ADR 记录「为什么这样设计」。
