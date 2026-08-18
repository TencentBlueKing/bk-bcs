---
name: tapd-story-implement
slug: tapd-story-implement
version: 3.0.0
description: |
  迭代执行流水线代码实现阶段。基于 tasks.md 调用 /speckit.implement 以 TDD 模式完成全部任务。
  本子 skill 不创建/切换分支，重试策略为原地修复；仅在跨 attempt 回退时做 git 路径级 checkout。
  命令详见 ../references/subagent-prompt-template.md §2.5。
---

# 需求实现执行

## 前置条件

- 当前需求 phase 为 `confirmed`
- 当前已在迭代分支上工作，**本子 skill 不创建分支、不切分支**
- `${WORKDIR}/` 下 `spec.md`、`plan.md`、`tasks.md` 已存在
- pipeline 主编排已按 `references/context-and-meta-template.md` 重写当前阶段的 `context.md`
  （白名单含：spec.md / plan.md / tasks.md + 编码与安全规范文档 + `Code scope` 白名单）
- 进入本子 skill 前 pipeline 主编排执行 `git rev-parse HEAD`，写入
  `meta.yaml.implement_baseline_commit`（仅用于回退重入的代码回滚）

## 输入

- `${WORKDIR}/meta.yaml`（读取 `attempts`）
- 当前需求 ID（由 pipeline 主编排传入）
- `${WORKDIR}/spec.md` / `plan.md` / `tasks.md`
- `${WORKDIR}/context.md`
- `${WORKDIR}/iteration-patches/attempt-${N}.md`（若为回退重入，必读）

## 执行流程

### 路由规则

pipeline 主编排在调起 implement subagent 前，读取 `meta.yaml.tech_type` 确定执行 agent：

| tech_type | 执行 agent | 说明 |
|-----------|-----------|------|
| `backend` | `backend-developer` | 单轮，完成后 phase → `implemented` |
| `frontend` | `frontend-developer` | 单轮，完成后 phase → `implemented` |
| `fullstack` | `backend-developer` → `frontend-developer`（串行）| 两轮，见下方"fullstack 串行实现"节 |

### fullstack 串行实现

fullstack 需求的 implement 阶段分两轮，均在 `confirmed → implemented` 同一 phase 转换内完成：

**Round 1（后端）**：
- 调起 `backend-developer`，完成后端部分的 TDD 实现
- 完成后更新 `meta.yaml.implement_rounds.completed = 1`、`current = frontend`

**Round 2（前端）**：
- 调起 `frontend-developer`
- **以 plan.md 中的 API 契约为权威输入**，可参考 Round 1 实现作为上下文（确认接口具体细节），不以 Round 1 实际代码产出覆盖 plan.md 契约
- 完成后更新 `meta.yaml.implement_rounds.completed = 2`，phase → `implemented`

**重入时**（同 attempt 内 round 重试或回退重入）：
- 读取 `implement_rounds.completed` 确定从哪一轮恢复
- `completed = 0`：从后端重新开始
- `completed = 1`：跳过后端，直接进入前端 Round 2
- `completed = 2`：两轮均已完成，不应再次进入 implement

pipeline 主编排按 `../references/subagent-prompt-template.md` §2.5 **Implement 模板**渲染
SUBAGENT_PROMPT（含 speckit-implement skill + 自检 + 测试运行），
通过 `Task(subagent_name=<IMPL_AGENT>, ...)` 拉起 subagent，`<IMPL_AGENT>` 由路由逻辑确定（见上方路由规则）。

渲染填充：`${ID}` / `${VERSION}` / `${WORK_DIR}` / `attempts` / `round`；
若 `attempts > 1`，附上 `iteration-patches/attempt-${attempts}.md` 摘要。

> subagent 内的 speckit skill、自检步骤（Code scope 校验 + 测试链）、
> 回传 JSON 的 metrics.tests 填写要求均在 subagent-prompt-template.md §2.5 中定义。

### 消费 subagent 回传 JSON

| status | 主会话动作 |
|--------|-----------|
| `ok` | 推进 phase → `implemented` |
| `blocked_on_questions` | 不预期；按 `error-handling.md` 处理，升级为回退重入 |
| `fail` | 决策三选一：**重试** / **回退重入** / **终止** |

### 重试（同 attempt 内 round +1，原地修复）

采取**原地修复策略**：

- **不回滚代码**、**不动 `implement_baseline_commit`**
- 已通过的任务产出继续保留，作为重试起点
- pipeline 主编排将失败上下文写入 `meta.yaml.implement_attempts[-1].rounds[-1]`，
  追加到下一轮 SUBAGENT_PROMPT 的重入增量段（subagent-prompt-template.md §3.2）
- round +1（attempts 不变），最多 3 个 round

> 连续 3 round 仍失败 → 强制升级为回退重入。
> 严禁在 round 重试中做任何 git 回滚。

### 回退重入（attempt +1）

两步流程：

1. **pipeline 退出前**：写 `meta.yaml.last_failure`（type=semantic），附建议的
   `target_phase_to_re_enter`，退出等待调用方决策。
2. **调用方修复后再次启动**：pipeline 主编排对 `Code scope` 白名单路径执行**路径级**
   回滚到 `implement_baseline_commit`，严格排除 `${WORKDIR}/` 目录。

```bash
# 回滚示例（pipeline 主编排执行，路径来自 Code scope 白名单）
BASELINE=$(yq '.implement_baseline_commit' "${WORKDIR}/meta.yaml")
git checkout "${BASELINE}" -- internal/order/ api/order/
git clean -fd -- internal/order/ api/order/
```

**严禁**：`git reset --hard` / `git stash` / `git checkout <branch>` / `git clean -fdx`（无路径限定）。

### 终止

调用方接到 fail 卡点后不再调起 pipeline 即可。

### 更新状态

`status=ok` 时：phase → `implemented`；history 追加成功记录（含 metrics.tests）。

## 可重入约定

| 重入类型 | pipeline 主编排动作 | meta.yaml 变化 |
|---------|--------------------|----------------|
| 同 attempt 内 round 重试 | 不动代码；失败上下文追加到 SUBAGENT_PROMPT | round +1，attempts 不变 |
| 回退重入到上游阶段 | `git checkout <baseline> -- <Code scope>`（排除 WORKDIR） | 仅当 `attempts < max_attempts` 时 attempts +1、phase 切换、baseline 重新取一次 |

子 skill 在执行前必须：

- 读取最新 `attempt-*.md`（若有），追加重入增量段（subagent-prompt-template.md §3）；
- 若同 attempt round 重试，追加 `implement_attempts[-1].rounds[-1].issue` 作为"已知失败原因"；
- 校验 `Code scope` 白名单未被意外缩窄。

幂等性、产物保留、process.log 仅追加等通用约束见 `../references/reentry-protocol.md`。

## 信息不足回退（回退重入入口）

| 根因 | target_phase_to_re_enter |
|------|-------------------------|
| 接口契约不明、外部依赖未识别 | `specify` |
| 计划缺少模块/数据模型 | `plan` |
| 任务遗漏关键场景 | `tasks` |

## 参考资料

- SUBAGENT_PROMPT（implement 模板 + 自检 + 测试）：`../references/subagent-prompt-template.md` §2.5
- 通用回传 JSON Schema：`../references/subagent-prompt-template.md` §1.3
- 重入增量段：`../references/subagent-prompt-template.md` §3
- 上下文与元数据模板：`../references/context-and-meta-template.md`
- 可重入协议（含 git 快照规则）：`../references/reentry-protocol.md`
- 错误处理：`../references/error-handling.md`

## 产出

- 代码实现已完成并通过单元/集成/端到端测试
- `${WORKDIR}/process.log` — 追加相关状态记录
- `${WORKDIR}/meta.yaml` — 更新后的 attempts / round / history（含 tests 子结构）
- 需求 phase 为 `implemented`
