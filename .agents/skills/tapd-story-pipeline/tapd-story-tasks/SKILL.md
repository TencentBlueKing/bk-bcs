---
name: tapd-story-tasks
slug: tapd-story-tasks
version: 3.0.0
description: |
  迭代执行流水线任务生成阶段。基于 plan.md / research.md 调用 /speckit.tasks 生成全量任务，
  随后调用 /speckit.analyze 验证产物合规。两段命令通过两次 subagent 隔离执行。
  命令与报告模板详见 ../references/subagent-prompt-template.md §2.3 + §2.4。
---

# 需求实现任务

## 前置条件

- 当前需求 phase 为 `researched`
- `${WORKDIR}/plan.md`、`research.md`、`data-model.md`（如有）已存在
- pipeline 主编排已按 `references/context-and-meta-template.md` 重写当前阶段的 `context.md`

## 输入

- `${WORKDIR}/meta.yaml`（读取 `attempts`）
- 当前需求 ID（由 pipeline 主编排传入）
- `${WORKDIR}/plan.md` / `research.md` / `data-model.md`
- `${WORKDIR}/context.md`
- `${WORKDIR}/iteration-patches/attempt-${N}.md`（若为回退重入，必读）

## 执行流程

本阶段由 pipeline 主编排先后渲染**两份** SUBAGENT_PROMPT，分别拉起两个 subagent：

1. **第一段**：按 `../references/subagent-prompt-template.md` §2.3 **Tasks-Generate 模板**
   渲染，调用 `speckit-tasks` 生成 `tasks.md`。
2. **第二段**：按 §2.4 **Tasks-Analyze 模板**渲染，调用 `speckit-analyze` 验证产物合规
   并产出 `tasks-report.md`。

拆成两段的理由：speckit-analyze 是独立skill，需要在 tasks.md 落盘后重新拉起 agent 进程。

渲染填充：`${ID}` / `${VERSION}` / `${WORK_DIR}` / `attempts` / `round`；
若 `attempts > 1`，附上 `iteration-patches/attempt-${attempts}.md` 摘要。

> 两段的 shell 命令、tasks-report.md 统一模板、Verdict 判定规则（四态归因）
> 均在 subagent-prompt-template.md §2.3 + §2.4 中定义，本子 skill 不重复。

### 内容门禁（主编排内联）

在 `tasks-analyze` 的合规判断中，主编排必须对 tasks.md 增加以下内容检查：

| 检查项 | 通过标准 |
|--------|----------|
| 格式与可执行性 | 每项任务使用 `- [ ] TNNN ...` 格式，包含具体动作和文件路径 |
| 用户价值切片 | 用户故事任务按优先级组织；每个故事均包含独立测试标准和端到端可验证的交付切片 |
| TDD 与验收 | TDD 被要求时，每个功能切片有对应测试任务或验收测试；任务描述可直接执行 |
| 依赖关系 | tasks.md 含依赖顺序和并行机会；阻塞任务在依赖任务之前 |
| 最小主流程 | 首个可交付用户故事覆盖 Walking Skeleton 或等效的核心 happy path |
| 完整性 | plan.md 中的接口、数据模型、迁移/配置和文档更新需求均映射到任务，或明确标注不适用 |

不要求任务标题使用固定术语；由 `tasks-analyze` 按语义判断。任一项不满足时，
`compliance.verdict=needs_fix`，报告须列出缺失项，主编排按同 attempt round 重试整组
tasks-generate + tasks-analyze。

### 消费 subagent 回传 JSON

第一段失败（`status=fail`）按 `references/error-handling.md` 处理，通常升级为回退重入。

第一段成功后进入第二段。按第二段 `compliance.verdict` 分支：

| Verdict | 主会话动作 |
|---------|-----------|
| `pass` | 进入"更新状态"推进 phase |
| `needs_fix` | 同 attempt 内 round +1，重新跑第一段 + 第二段（最多 3 个 round） |
| `plan_insufficient` | 整理 findings 为 `attempt-${N+1}.md`，`target_phase_to_re_enter=plan`，走回退重入 |
| `spec_insufficient` | 同上，`target_phase_to_re_enter=specify` |

### 更新状态

`compliance.verdict=pass` 时：

- 更新 `meta.yaml.phase` 为 `tasks-generated`
- `meta.yaml.history` 追加成功记录，清空 `meta.yaml.last_failure`

## 可重入约定

通用规则见 `../references/shared-reentry-conventions.md`。本阶段差异：

- **round 重试触发条件**：`compliance.verdict=needs_fix`，同 attempt 内最多 3 个 round
- **两段 subagent 均追加重入增量段**（subagent-prompt-template.md §2.3 + §2.4）
- **报告文件不做快照**：`tasks-report.md` 每轮覆盖重写，历史可经 process.log 回溯

幂等性、产物保留、process.log 仅追加等通用约束见 `../references/reentry-protocol.md`。

## 信息不足回退（回退重入入口）

`verdict=plan_insufficient` 或 `spec_insufficient` 时，`attempt-${N+1}.md` 必须明确：

- `failed_phase: tasks`
- `root_cause`：可直接复用 `tasks-report.md` 中对应 Finding 的根因字段
- `target_phase_to_re_enter: plan` 或 `specify`（按归因最深处确定）

## 参考资料

- SUBAGENT_PROMPT（tasks-generate + tasks-analyze 模板 + 报告模板）：`../references/subagent-prompt-template.md` §2.3 + §2.4
- 通用回传 JSON Schema：`../references/subagent-prompt-template.md` §1.3
- 上下文与元数据模板：`../references/context-and-meta-template.md`
- 可重入协议：`../references/reentry-protocol.md`
- 错误处理：`../references/error-handling.md`

## 产出

- `${WORKDIR}/tasks.md` — 全量任务清单
- `${WORKDIR}/tasks-report.md` — speckit-analyze 合规报告
- `${WORKDIR}/process.log` — stream-json 流式日志（追加，含两段 banner）
- `${WORKDIR}/meta.yaml` — 更新后的 attempts / round / history
- 需求 phase 为 `tasks-generated`
