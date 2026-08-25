# State Mutation Guide

> pipeline 状态管理规则与调用方语义指令参考。
>
> **核心原则**：meta.yaml 由 pipeline 独占写入，外部只读。
> 调用方通过语义指令表达意图，pipeline 自行执行状态变更。
> 这样做是因为 meta.yaml 是 pipeline 的内部状态机——让单一写者管理状态，
> 避免外部修改导致的状态不一致和 mutation_invalid 校验负担。

## 目录

- [State Mutation Guide](#state-mutation-guide)
  - [目录](#目录)
  - [1. meta.yaml 写入权限](#1-metayaml-写入权限)
  - [2. 语义指令集](#2-语义指令集)
  - [3. 各指令的 pipeline 内部动作](#3-各指令的-pipeline-内部动作)
  - [4. 卡点判定优先级](#4-卡点判定优先级)
  - [5. 内容文件修改权限](#5-内容文件修改权限)
  - [6. phase 状态链（pipeline 内部规则）](#6-phase-状态链pipeline-内部规则)
  - [7. pipeline 启动校验](#7-pipeline-启动校验)
  - [8. runner 语义指令速查表](#8-runner-语义指令速查表)
  - [9. 用户独立模式速查表](#9-用户独立模式速查表)

---

## 1. meta.yaml 写入权限

| 写者 | 权限 | 原因 |
|------|------|------|
| pipeline | ✅ 所有字段（独占写入） | pipeline 是状态机的唯一所有者，保证状态推进的原子性和一致性 |
| runner / 用户 | ❌ 只读 | 通过语义指令间接触发变更，避免双写者冲突和非法状态跳跃 |

`tech_type`：由 specify 阶段 tech-lead subagent 完成后写入，**此后 pipeline 不得覆盖**。
外部调用方只读。若需修改技术类型，须重新从 initialized 执行（此时 tech_type 留空，
specify 重新判断后写入）。

## 2. 语义指令集

调用方（runner 或用户）通过以下 6 个指令与 pipeline 交互。
每个指令对应一种明确的用户意图——pipeline 收到后自行决定如何修改 meta.yaml。

| 指令 | 语义 | 适用卡点 | 调用方需提前做的事 |
|------|------|---------|-----------------|
| `execute` | 首次调度或正常继续推进 | 无卡点 / 首次 | 注入 req.md（仅首次） |
| `approve` | 审查通过，放行进入 implement | confirm | 审查 spec.md / plan.md / tasks.md |
| `reject` | 审查不通过，回退重做 | confirm | 写好 `iteration-patches/attempt-${N}.md` |
| `answer` | 已回答问题，继续推进 | blocked | 修改 questions.md（[open] → [answered]） |
| `retry` | 失败已修复，重试 | fail | 可选：修改 req.md / spec.md / context.md / 写 attempt-${N}.md |
| `abort` | 放弃当前需求 | 任意 | 无 |

**指令在 Task prompt 中的传递方式**：

```
任务：执行 tapd-story-pipeline
指令：${ACTION}
工作目录：${WORKDIR}
工作空间 ID：${WORKSPACE_ID}          # 仅 execute 时需要
agent 工具：${AGENT_TOOL}             # 仅 execute 时需要
迭代分支：${ITER_BRANCH}              # 仅 execute + runner 调度时需要

请按 skills/tapd-story-pipeline/SKILL.md 执行。
退出后输出退出报告。不要询问后续操作。
```

## 3. 各指令的 pipeline 内部动作

pipeline 收到指令后自行执行以下 meta.yaml 变更——调用方无需关心具体字段操作：

| 指令 | meta.yaml 变更 | 后续动作 |
|------|---------------|---------|
| `execute` | 首次：初始化全部字段；非首次：无变更 | 从当前 phase 继续推进 |
| `approve` | `phase` → `confirmed`；`pending_review` → null | 继续推进到 implement |
| `reject` | 若 `attempts < max_attempts`：`phase` → target（从 attempt-md 读取）；`attempts` +1；`pending_review` → null | 从 target phase 重跑；达到上限时写 fail 并退出 |
| `answer` | 校验 questions.md 无 [open]；`round` +1 | 继续当前阶段 |
| `retry` | 若 `attempts < max_attempts`：`last_failure` → null；`attempts` +1；若有 attempt-md 则按 target 回退 phase | 从目标 phase 重跑；达到上限时写 fail 并退出 |
| `abort` | 写入 `last_failure.type=user_aborted` | 退出 |

## 4. 卡点判定优先级

pipeline 退出后，调用方按以下顺序读取 meta.yaml（只读）确定卡点类型：

1. `phase == committed` → pipeline 完工（完工前主编排已校验 `commit.md` 存在、成本已对账补齐）
2. `last_failure` 非空 → fail 卡点（含 `mutation_invalid` / `user_aborted` 子类）
3. `pending_review` 非空 → confirm 卡点
4. `questions.md` 有 `[open]` 条目 → blocked 卡点
5. 否则 → pipeline 退出异常（需查 process.log）

### Attempt 上限

- `meta.yaml.max_attempts` 当前固定为 `3`，`attempts` 从 `1` 起计。
- 需要回退重入或执行 `reject` / `retry` 时，主编排必须先检查
  `attempts < max_attempts`。
- 若已达到上限，禁止修改 `phase`、`attempts` 或 `pending_review`，并写入
  `last_failure.type=semantic`、`last_failure.message="attempts exhausted (3/3): {最近失败摘要}"` 后退出。
- 此时由调用方决定人工修订内容文件后重新初始化，或执行 `abort`；不得自动再次调度。

## 5. 内容文件修改权限

meta.yaml 由 pipeline 独占，但以下**内容文件**仍由外部修改——
它们是调用方的"输入"，不是 pipeline 的"内部状态"：

| 文件 | 外部允许修改 | 修改时机 | 说明 |
|------|------------|---------|------|
| `req.md` | ✅ | retry 前 | 需求内容是用户输入 |
| `questions.md`（答复字段） | ✅ | answer 前 | 仅修改 [open] → [answered]，填入回答 |
| `spec.md` / `plan.md` / `tasks.md` | ✅ | retry 前 | 修正产物后 pipeline 基于当前 phase 重跑 |
| `context.md` | ✅ | retry 前 | 补充白名单 |
| `iteration-patches/attempt-${N}.md` | ✅（新增） | reject / retry 前 | 改进方案，pipeline 在重入时读取确定回退目标 |
| `process.log` | ❌ | — | 审计日志，只追加 |
| `commit.md` | ❌ | — | pipeline 产出；commit 阶段有产出门禁校验，主编排完工前兜底校验其存在 |

## 6. phase 状态链（pipeline 内部规则）

```
initialized → tech-clarified → researched → tasks-generated → confirmed → implemented → validated → committed
```

pipeline 自行遵守以下规则（调用方无需关心）：
- 只能沿链**正向推进**或**回退到链上某节点**
- 不能跨越前进（如 initialized → confirmed）
- 回退时校验前置产物存在

## 7. pipeline 启动校验

pipeline 收到指令后做轻量校验，确保输入合法：

| 校验项 | 适用指令 | 失败处理 |
|--------|---------|---------|
| 产物一致性：当前 phase 的前置产物存在 | 所有 | `last_failure.type=mutation_invalid` 退出 |
| questions.md 格式：条目状态为合法四态值 | `answer` | 同上 |
| attempt-md 存在且格式正确 | `reject` / `retry`（有 attempt-md 时） | 同上 |

## 8. runner 语义指令速查表

```
1. 调度 pipeline：action=execute（首次）
2. pipeline 退出后读 meta.yaml（只读）：
   phase == committed → 该需求完成，下一个
3. last_failure 非空：
     type=system 且 attempts < 2 → action=retry（自动重试）
     其他 → 与用户对话：
       用户选择修复 → 用户改内容文件/写 attempt-md → action=retry
       用户放弃 → action=abort
4. pending_review 非空：
     展示 artifacts → 用户决策：
       通过 → action=approve
       回退 → 用户写 attempt-md → action=reject
       放弃 → action=abort
5. questions.md 有 [open]：
     与用户对话 → 写回 [answered] → action=answer
6. 以上皆无 → pipeline 退出异常，action=execute 重试或放弃
```

## 9. 用户独立模式速查表

用户的自然语言直接映射为语义指令：

| 用户说 | 映射到 action |
|--------|--------------|
| "帮我实现需求 #ID" | `execute` |
| "通过" / "确认" / "审查通过" | `approve` |
| "回退到 specify" / "重做规范" | `reject`（先写 attempt-md） |
| "继续" / "我已经回答了" | `answer` |
| "重试" / "我改好了" | `retry` |
| "放弃" / "不做了" | `abort` |

```
操作步骤：
1. 说"帮我实现需求 #ID" → pipeline 以 action=execute 启动
2. 看退出报告（phase / 卡点类型 / 下一步指令）
3. 根据卡点操作内容文件 → 说对应的自然语言触发下一轮
4. 无需手动编辑 meta.yaml
```
