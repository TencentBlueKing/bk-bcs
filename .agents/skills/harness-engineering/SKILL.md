---
name: harness-engineering
slug: harness-engineering
version: 1.6.7
description: |
  Harness Engineering 编排器——为 AI Agent 构建运行环境规范。
  通过触发词路由到两个子 skill：harness-generating（规范文档生成/修正）
  和 harness-gardening（文档一致性巡检与修复）。
  Use this skill whenever the user mentions Harness Engineering, AI Agent 运行环境,
  驾驭工程, Agent 规范, 上下文工程, 架构约束, 熵管理, 工具能力, 执行与验证,
  生成 Harness 文档, Agent 治理, AI 治理规范, 构建 Agent 环境, Agent harness,
  AGENTS.md, 词汇表, glossary,
  文档园艺, 文档巡检, 检查文档一致性, 文档是否过时, harness gardening,
  扫描文档, 园艺扫描,
  开发地图, dev map, 生成开发地图, 更新开发地图, 项目地图, 代码索引, 源文件索引, 模块索引, 模块依赖,
  or any workflow involving Harness spec generation or document maintenance.
metadata:
  requires:
    skills: ["harness-generating", "harness-gardening"]
---

# Harness Engineering

## 核心理念

Harness Engineering 的本质是为 AI Agent 构建可靠的"运行环境"——不是优化模型本身，
而是通过系统级工具链解决模型在真实环境中的状态管理、工具调用、任务漂移、结果验证等问题。
当 Agent 表现不佳时，优先审视环境支持是否充分，而非归咎于模型能力不足。

## 路由与编排

收到用户消息后，按以下顺序判定路由目标，**必须**读取对应子 skill 的 SKILL.md 再执行：

1. 匹配**维护类**触发词 → `harness-gardening`
2. 匹配**生成类**触发词（或默认） → `harness-generating`
3. 无法判定 → 询问用户意图（生成规范 / 文档巡检）

| 路由目标 | 触发词（示例） | 调用方式 |
|---------|--------------|---------|
| **生成** | "生成 Harness 规范"、"AI 治理规范"、"修改架构约束"、"生成词汇表"、"生成开发地图" | `harness-generating`（mode=full 或 targeted） |
| **维护** | "文档园艺"、"文档巡检"、"检查文档一致性"、"harness gardening"、"扫描文档"、"更新开发地图" | `harness-gardening` |

### harness-generating（生成类）

读取 `harness-generating/SKILL.md`，传入 `mode`（`full` / `interactive` / `targeted`）、`target`（定向修正时，如 `context-engineering`）、`workspace_root`。

| mode | 判定条件 |
|------|---------|
| `full` | "生成 Harness 规范"、"创建 Agent 环境文档"、首次使用 |
| `interactive` | 用户信息不足以生成完整规范 |
| `targeted` | 指定修改某个组件，如"生成开发地图"（target=dev-map） |

### harness-gardening（维护类）

读取 `harness-gardening/SKILL.md`，传入 `mode`（`pr` / `full` / `targeted`）、`target`（targeted 时必填）、`workspace_root`。

| mode | 判定条件 |
|------|---------|
| `pr` | commit 后 hook / 用户指定轻量检查 |
| `full` | "文档巡检"、"检查文档一致性" / 默认全量扫描 |
| `targeted` | 指定检查某个组件，如"更新开发地图"（target=dev-map）、"检查词汇表"（target=glossary） |

## 参考资源

| 文件 | 用途 | 读取方 |
|------|------|-------|
| `harness-generating/SKILL.md` | 规范文档生成流程（含 dev map） | 父 skill（生成类路由） |
| `harness-gardening/SKILL.md` | 文档一致性巡检流程（含项目记忆等维度） | 父 skill（维护类路由） |
| `references/project-type-detection.md` | 项目类型识别与 Skill 安装根检测 | 两个子 skill（执行前必读） |
| `references/skill-install-root.md` | Skill 安装根探测与推荐布局 | generating / doctor / 文档表述 |
| `references/agents-merge.md` | 根 AGENTS 先理解再合并（保留项目记忆） | generating 第三步；gardening 记忆维度 |
| `references/agents-work-units.md` | 工作单元 `**/AGENTS.md` 发现与根索引 | generating 写根前；gardening 维度 12 |
| `references/project-owned-tools.md` | 项目自有工具（git 跟踪 ∩ 安装布局，含 monorepo 子树） | tooling 分节；doctor 必查 |
| `scripts/harness-doctor.sh` | 运行时探活（stdout，不入库） | generating 第一步；人工排查 |
| `references/tool-dependencies.md` | Agent 工具依赖权威清单 | generating（环境检查）；gardening（维度 7） |
| `references/best-practices.md` | 五大组件最佳实践详解 | 子 skill（生成/修复时） |
| `assets/` | 文档模板、`dev-map.gitignore`、技术规范预设库 | 子 skill（生成流程中） |
| `scripts/detect-standards.sh` | monorepo 技术栈 Level-1 探测（R2） | generating 第三步-B |
