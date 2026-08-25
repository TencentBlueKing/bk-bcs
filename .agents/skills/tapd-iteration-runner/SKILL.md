---
name: tapd-iteration-runner
slug: tapd-iteration-runner
version: 2.0.0
description: |
  TAPD 迭代调度器——批量开发一个迭代中的全部需求。自动完成环境初始化、
  需求依赖分析、逐需求调度 pipeline 实现、迭代汇总报告四段编排。
  Use this skill whenever the user mentions 迭代执行, 开发迭代, 批量需求实现,
  TDD 迭代, dry run, batch requirement implementation, TAPD iteration development,
  speckit-based iteration, 迭代开发, or any workflow involving batch TAPD story processing
  with iteration-level coordination — even if the user only says "帮我开发这个迭代"
  or "跑一下迭代".
metadata:
  requires:
    mcps: ["tapd", "git"]
    os: ["linux", "macos", "windows"]
    skills: ["tapd-story-pipeline"]
---

# TAPD 迭代调度器

## 1. 定位

本 skill 负责迭代级编排：建分支、拉需求、依赖排序、依次调度每个需求的 pipeline、迭代收尾。

**runner 不知道的事**：

- 需求内部的子 skill 步骤是什么
- speckit 命令长什么样
- context.md / req.md / spec.md 的格式
- subagent prompt 怎么渲染
- 单需求层面的重入怎么做

这些全部由 `tapd-story-pipeline` 承担。runner 在自己的主上下文中**内联加载并执行**
`tapd-story-pipeline/SKILL.md`，仅根据 `meta.yaml` 与 `questions.md` 做决策。

## 2. 基础配置

项目根目录下的 `project.json` 提供默认配置：

```json
{
  "workspace_id": "对应的 TAPD 工作空间 ID",
  "owner": "开发者 ID"
}
```

**取值优先级**（适用于 `workspace_id` 和 `owner`）：

1. 用户在消息中显式指定 → 最高优先
2. `project.json` 中的配置 → 次优先
3. 通过 TAPD MCP 自动获取（仅 `owner`）→ 再次
4. 交互式询问用户 → 最低优先（兜底）

`iteration_id` 无默认值，必须由用户提供或通过交互获取。

## 3. runner → pipeline 参数传递

用户在发起迭代时可指定 agent 工具名称（如 `agent` / `claude` / `cursor`），存入
`iteration-state.json.agent_tool`。未指定时默认为 `agent`。

runner 内联执行 pipeline 时把 `agent_tool` 作为入参传入；pipeline 在自己的 `meta.yaml` 中
也写一份，作为后续单次调起的入参回退。（`agent_tool` 决定 pipeline 子 skill 用哪种 subagent 工具。）

`workspace_id` 同理：runner 在 loop 内联执行时把 `iteration-state.json.workspace_id`
作为入参传给 pipeline；pipeline 首次执行时把它落到 `meta.yaml.workspace_id`，
后续 pipeline 子 skill（如 commit 调 `stories_update`、req.md 兜底注入调 `stories_get`）
统一从 `meta.yaml` 取，避免依赖运行时入参或 `project.json` 兜底。

## 4. 子 Skill 编排

| 顺序 | 子 skill | 职责 | 停断点 |
|------|----------|------|--------|
| 1 | `tapd-iteration-init` | 环境检查、分支创建、状态初始化、恢复检测 | 否 |
| 2 | `tapd-iteration-analysis` | 需求依赖分析、DAG 构建、拓扑排序 | 否 |
| 3 | （内联）loop | 对 sequence 中每个需求**内联执行** `tapd-story-pipeline` | 是 |
| 4 | `tapd-iteration-report` | 迭代汇总：changelog + 效能报告 + 推送 | 否 |

## 5. iteration-state.json

迭代级状态文件，保存在 `specs/${VERSION}/iteration-state.json`。

**完整字段定义**见 `references/iteration-state-schema.md`。

**决策常用字段速查**：

| 你需要… | 读取字段 |
|---------|---------|
| 判断迭代所处阶段 | `status` |
| 遍历待实现需求 | `sequence` |
| 判断某需求是否完成 | `stories.*.phase`（派生自 meta.yaml） |
| 传递参数给 pipeline | `iter_branch`, `agent_tool`, `workspace_id` |
| 汇总效能数据 | `stories.*.stats`（派生自 meta.yaml） |

**写入权限**：

- runner **完全拥有** `iteration-state.json`，因为迭代级状态只有一个写者，避免 runner 与 pipeline 并发修改导致数据冲突
- `stories.*.phase` 和 `stats` 是 `meta.yaml` 的**派生缓存**——runner 在 loop 中从各需求的 `meta.yaml` 同步过来，用于在 `iteration-state.json` 单点呈现迭代级全局视图
- runner **不直接修改** pipeline 的 `meta.yaml`——pipeline 对自己的 meta.yaml 有完整的校验逻辑（`state-mutation-guide.md` §4），外部修改必须遵循 §7 速查表规则，确保 phase 与产物的对应关系始终合法

## 6. 状态流转

```
initialized → analyzed → implementing → completed → reported
                        └→ bugfix ──────┘
```

| 状态 | 说明 |
|------|------|
| `initialized` | 环境初始化已完成 |
| `analyzed` | 需求依赖分析完成 |
| `implementing` | 需求实现中（loop 进行中）|
| `bugfix` | 已发布版本出现 bug，回到迭代修复 |
| `completed` | 所有需求 phase=committed |
| `reported` | 迭代汇总已完成（终态）|

## 7. loop 调度逻辑

runner 主 SKILL 的核心。对 `sequence` 中每个 ID，用语义指令调度 pipeline。
runner 不修改 `meta.yaml`——所有状态变更由 pipeline 执行。

以下为逻辑伪代码（非可执行脚本），描述 runner 的语义调度循环：

```pseudocode
for ID in sequence:
    iteration_state.selected_story = ID
    workdir = f"specs/{VERSION}/{ID}"

    # 记录壁钟开始时间（仅首次；持久化以便恢复后仍可计算耗时）
    if not iteration_state.stories[ID].started_at:
        iteration_state.stories[ID].started_at = now_iso8601()
        save_iteration_state()

    # 注入 req.md（首次）
    if not exists(f"{workdir}/req.md"):
        ensure_dir(workdir)
        if exists(f"specs/{VERSION}/.tmp/{ID}.md"):
            copy(f"specs/{VERSION}/.tmp/{ID}.md", f"{workdir}/req.md")
        else:
            desc = tapd_mcp.stories_get(workspace_id, id=ID).description
            write_file(f"{workdir}/req.md", desc)

    action = "execute"  # 首次用 execute

    while True:
        # 内联执行 pipeline（runner 不修改 meta.yaml）：
        # 按 action 与入参加载并执行 skills/tapd-story-pipeline/SKILL.md，
        # pipeline 在当前上下文中推进到下一个卡点后退出
        # （其子 skill specify/plan/... 各自拉 speckit subagent）。
        run_inline(skill="skills/tapd-story-pipeline/SKILL.md",
                   action=action, story_id=ID, workdir=workdir,
                   workspace_id=iteration_state.workspace_id,
                   agent_tool=iteration_state.agent_tool,
                   iter_branch=iteration_state.iter_branch)

        # pipeline 退出后读 meta.yaml（只读）判断结果
        meta = read_yaml(f"{workdir}/meta.yaml")

        if meta.phase == "committed":
            # 记录壁钟耗时 = committed 时刻 − started_at
            iteration_state.stories[ID].stats.duration_sec = \
                now_epoch() - parse_iso8601(iteration_state.stories[ID].started_at)
            sync_to_iteration_state(ID, phase="committed", stats=meta.stats)
            save_iteration_state()
            break

        # 根据卡点确定下一个 action（参见 state-mutation-guide.md §8）
        if meta.last_failure:
            if meta.last_failure.type == "system" and meta.attempts < 2:
                action = "retry"  # 自动重试，无需与用户对话
            else:
                user_input = chat_with_user(meta.last_failure)
                if user_input == "重试":
                    # 用户可能已修改 req.md / spec.md 等内容文件，并写好 attempt-md
                    action = "retry"
                else:
                    action = "abort"

        elif meta.pending_review:
            user_input = chat_with_user(meta.pending_review)
            if user_input == "通过":
                action = "approve"
            elif user_input == "回退":
                # 用户已写好 iteration-patches/attempt-${N}.md
                action = "reject"
            else:
                action = "abort"

        elif has_open_questions(f"{workdir}/questions.md"):
            user_answers = chat_with_user(read_open_questions(workdir))
            write_answers_to_questions_md(workdir, user_answers)
            action = "answer"

        else:
            # pipeline 退出异常
            notify_user(f"pipeline 退出异常，请检查 {workdir}/process.log")
            if chat_with_user("重试？") == "重试":
                action = "execute"
            else:
                break
```

**关键原则**：runner 只做三件事——
1. 读 meta.yaml（只读）判断卡点
2. 与用户对话
3. 写用户输入文件（questions.md 答复、attempt-md）并确定下一个 action

## 8. 与 pipeline 的通信契约

### 8.1 内联调用约定

runner 在 loop 中**内联执行** pipeline，传入以下入参：

| 入参 | 来源 |
|------|------|
| `ACTION` | loop 决策结果（`execute` / `approve` / `reject` / `answer` / `retry` / `abort`）|
| `WORKDIR` | `specs/${VERSION}/${ID}` |
| `WORKSPACE_ID` | `iteration-state.json.workspace_id` |
| `AGENT_TOOL` | `iteration-state.json.agent_tool` |
| `ITER_BRANCH` | `iteration-state.json.iter_branch` |

runner 按上述入参与 `ACTION` 加载并执行 `skills/tapd-story-pipeline/SKILL.md`。
pipeline 推进到下一个卡点后退出（输出其 §8.1 退出报告），控制权回到 loop。

**ACTION 取值**：`execute` / `approve` / `reject` / `answer` / `retry` / `abort`
各指令语义见 `../tapd-story-pipeline/references/state-mutation-guide.md` §2。

### 8.2 退出后读取的文件

runner 只读以下文件（**不写 meta.yaml**）：

- `${WORKDIR}/meta.yaml` — 只读，判断卡点类型
- `${WORKDIR}/questions.md` — 只读判断 + 写答复（外部可写的内容文件）
- `${WORKDIR}/commit.md` — 仅在 `tapd-iteration-report` 阶段批量读取

### 8.3 退出报告用途

pipeline 退出时输出的文字报告面向用户，runner 的决策**不依赖**退出报告——所有决策数据从 `meta.yaml`（只读）和 `questions.md` 读取。

### 8.4 契约版本

本契约基于 `tapd-story-pipeline v1.1.0` 语义指令模式 + runner v2.0.0 内联执行模式。runner 的 loop 逻辑与 `state-mutation-guide.md` §8 runner 速查表保持同步。

## 9. 恢复机制

当 `tapd-iteration-init` 检测到已有未完成的 `iteration-state.json` 时，按
`references/recovery.md` 流程恢复。恢复时：

1. 读 `iteration-state.json`
2. 对 `sequence` 中所有非 `committed` 需求，加载其 `meta.yaml`
3. 同步 `meta.yaml.phase` / `meta.yaml.stats` 到 `iteration-state.json.stories.*.phase` / `stats`
   （派生缓存的一次性回填）
4. 遗留的卡点（`last_failure` / `pending_review` / `questions.md [open]`）由 loop 在下一轮调度时自然处理

## 10. 迭代汇总触发

每次 pipeline 退出且推进了 phase 后，runner 检查所有需求 phase 是否全部为 `committed`：

- 是 → `iteration-state.json.status = completed`，加载 `tapd-iteration-report/SKILL.md` 执行汇总；汇总完成后 `status = reported`（终态）
- 否 → 继续 loop

## 11. 平台约定

| OS | Shell | 备注 |
|----|-------|------|
| Linux / macOS | Bash | 默认 |
| Windows | PowerShell | 首次运行设置 `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8` |

`${AGENT_TOOL}` 从 `iteration-state.json.agent_tool` 字段读取。

## 12. 参考文件

| 文件 | 用途 | 加载时机 |
|------|------|---------|
| `references/iteration-state-schema.md` | iteration-state.json 完整字段定义与写入权限 | 需要查看字段结构时 |
| `references/recovery.md` | 迭代级断点恢复映射表 | init 阶段检测到已有状态时 |
| `references/error-handling.md` | 迭代层错误处理规则 | 遇到 TAPD MCP / git / pipeline 异常时 |
| `references/dependency-types.md` | 技术依赖与业务依赖判定条件 | analysis 阶段识别需求依赖时 |
| `references/report-scripts.md` | changelog 和 summary 生成脚本 | report 阶段需要实现细节时 |
| `../tapd-story-pipeline/references/state-mutation-guide.md` | **必读**：runner 决策依据（§7 速查表） | loop 决策时 |

## 13. Example: 3 需求迭代 Happy Path

用户输入："帮我开发 TAPD 迭代 #1001234 的需求，workspace_id 是 20000001"

**阶段 1 — init：**
- 读取 project.json 获取 workspace_id=20000001
- 调 TAPD MCP `iterations_get` → 迭代名称 `iteration-v0.9.x`
- 创建分支 `v0.9.x`，创建 `specs/v0.9.x/`，写入 `iteration-state.json`（status=initialized）

**阶段 2 — analysis：**
- 调 TAPD MCP `stories_get` → 3 个子需求 A、B、C
- A 无依赖，B 依赖 A（共享模块），C 依赖 B（接口依赖）
- sequence = [A, B, C]，status → analyzed

**阶段 3 — loop：**
```
需求 A：内联执行 pipeline → phase 推进到 committed → sync → 下一个
需求 B：内联执行 pipeline → phase=tasks-generated（confirm 卡点）
  → runner 展示 artifacts → 用户"通过" → phase=confirmed → 再次内联执行 → committed
需求 C：内联执行 pipeline → committed
全部 committed → status=completed
```

**阶段 4 — report：**
- 生成 changelog.md（v0.9.0）
- 生成 summary.md（效能统计）
- patch 自增 → 1
- git commit + push → status=reported（终态）
