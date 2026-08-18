# Pipeline 用户操作手册

> 本文档总结 `tapd-story-pipeline` 在两种调度模式下的操作规范。
> 面向用户说明如何使用 `tapd-story-pipeline`。

---

## 一、Runner 调度 Pipeline

### 1.1 首次调度（新需求）

| 步骤 | runner 动作 | 涉及文件 |
|------|------------|---------|
| 1 | 设置 `iteration_state.selected_story = ID` | `iteration-state.json` |
| 2 | 创建工作目录 `specs/${VERSION}/${ID}/` | 目录创建 |
| 3 | 注入 req.md（从 `.tmp/${ID}.md` 复制或 TAPD MCP 拉取） | `req.md` |
| 4 | 发起 Task，action=execute | 无文件修改 |

### 1.2 重新调度（pipeline 退出后再次调起）

Runner 读取 meta.yaml（**只读**），确定卡点类型和下一个 action：

| 卡点类型 | runner 动作 | action | 涉及文件 |
|---------|------------|--------|---------|
| **committed** | sync 到 iteration-state.json | — | `iteration-state.json` |
| **fail (system, attempts<2)** | 自动重试 | `retry` | 无 |
| **fail (semantic)** | 与用户对话 → 用户修改内容文件 | `retry` | 用户改 `req.md`/`spec.md` 等 + 写 `attempt-${N}.md` |
| **confirm** | 展示 artifacts → 用户审查 | `approve` 或 `reject` | reject 时用户写 `attempt-${N}.md` |
| **blocked** | 与用户对话 → 写答复 | `answer` | `questions.md`（[open] → [answered]） |
| **异常退出** | 通知用户 | `execute`（重试）| 无 |
| **放弃** | 不再调度 | `abort` | 无 |

**runner 不修改 meta.yaml** —— 所有状态变更由 pipeline 收到 action 后自行执行。

---

## 二、人工（用户独立模式）调度与重入 Pipeline

### 2.1 首次调度

用户说："帮我实现需求 #1234567890，workspace 20000001"
→ pipeline 以 `action=execute` 启动，自动创建目录、拉取需求、初始化 meta.yaml。

### 2.2 重入（pipeline 退出后继续）

| 卡点类型 | 用户操作 | 用户说 | 需修改的文件 |
|---------|---------|--------|------------|
| **confirm** | 审查 spec/plan/tasks | "通过" 或 "回退" | reject 时写 `attempt-${N}.md` |
| **blocked** | 回答 questions.md 问题 | "继续" | `questions.md`（[open] → [answered]） |
| **fail** | 修改内容文件 | "重试" | `req.md` / `spec.md` / `context.md` + 可选 `attempt-${N}.md` |
| **放弃** | — | "放弃" | 无 |

**用户无需手动编辑 meta.yaml** —— 自然语言自动映射为语义指令。

### 2.3 semantic 失败的修复操作

| 根因位置 | 需修改的文件 | 用户说 |
|---------|------------|--------|
| 需求理解错误 | `req.md` + `attempt-${N}.md` | "重试" |
| 白名单不全 | `context.md` | "重试" |
| 规范有误 | `spec.md` + `attempt-${N}.md` | "重试" |
| 计划有误 | `plan.md` + `attempt-${N}.md` | "重试" |
| 任务有误 | `tasks.md` + `attempt-${N}.md` | "重试" |

pipeline 收到 `retry` 指令后，自行读取 `attempt-${N}.md` 确定回退目标，执行 meta.yaml 状态变更。

### 2.4 操作安全建议

1. **不要手动编辑 meta.yaml** —— pipeline 会自行管理
2. 修改内容文件（req.md / spec.md 等）后通过语义指令触发 pipeline
3. 回退重入时务必写好 `iteration-patches/attempt-${N}.md`，pipeline 依赖它确定回退目标
4. 确保 `questions.md` 的答复格式正确（[open] → [answered] + 填入回答内容）

## hook 成本采集安装说明（目标项目首次启用）

`tapd-story-pipeline` 通过宿主 IDE PostToolUse hook 自动采集 subagent 执行成本。
首次在目标项目使用前，需完成以下两步安装：

### 1. 拷贝 hook 脚本

本skill默认安装后，hook脚本位于：

```
${CLAUDE_PROJECT_DIR}/.agents/skills/tapd-story-pipeline/scripts/subagent_usage.py
```

> 软链事实：目标项目的 `.codebuddy / .cursor` 通常是 `.agents` 的软链，
> 拷贝到 `.agents/skills/...` 后三 IDE 自动同步。

### 2. 合并 PostToolUse hook 配置

把本仓库 `skills/tapd-story-pipeline/scripts/settings.json` 中的
`hooks.PostToolUse` 段合并到目标项目的 `.claude/settings.json`：

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Task",
        "hooks": [
          {
            "type": "command",
            "command": "python3 ${CLAUDE_PROJECT_DIR}/.agents/skills/tapd-story-pipeline/scripts/log_usage.py"
          }
        ]
      }
    ]
  }
}
```

> `${CLAUDE_PROJECT_DIR}` 由宿主 IDE 在 hook 触发时自动注入为目标项目根。

### 3. 验证安装

跑一次 pipeline 单需求；confirm 卡点退出后查看：

```bash
ls ${WORK_DIR}/cost-events.jsonl  # 应存在，每个 subagent 阶段一行
```

若 `cost-events.jsonl` 不存在，pipeline 仍可正常推进，但 `meta.yaml.stats.cost`
所有 `total_*` 字段会保持为 0，`process.log` 中会有 warning 提示。

### 4. 已知约束

- hook 采集的 `usage.credit` 是积分单位，承载 `meta.yaml.stats.cost.total_cost_usd` 字段值（不再是美元）
- duration 来自 `tool_call_brief: "Execution Summary: N tool uses, cost: <s>s"` 解析，秒级精度
- hook 任何异常都静默并返回 `{"continue": true}`，绝不阻塞主流程
