---
name: bk-monitor-tapd-dev
description: 根据 TAPD 单据分析开发需求并生成方案。当用户需要开发 TAPD 需求/缺陷/任务时使用。
---

# TAPD 单据开发助手

## 工作流程

### Step 1: 获取 TAPD ID 并验证分支（阻塞）

1. 获取 ID：用户提供或从分支名提取（`git branch --show-current`，匹配 `#(\d+)`）
2. 验证分支是否包含该 ID
   - 匹配 → 继续
   - 不匹配 → 询问用户选择（创建分支/继续开发）

### Step 2-3: 获取单据

1. 解析 workspace_id = id.substring(2, 10)
2. 调用 MCP（stories_get、bugs_get、tasks_get）获取单据详情和评论

### Step 4-6: 分析与开发

1. 分析需求（标题、描述、评论、截止日期）
2. 按 dev-plan-template.md 生成方案
3. 等待用户确认后开始开发

## 注意事项

- 分支验证是第一优先级
- 方案需用户确认才能开始开发
- 评论与描述冲突时以最新评论为准

---

## 📦 可用资源

- `skill://bk-monitor-tapd-dev/references/branch-workflow.md`
- `skill://bk-monitor-tapd-dev/references/dev-plan-template.md`
- `skill://bk-monitor-tapd-dev/references/mcp-calls.md`
- `skill://bk-monitor-tapd-dev/references/tapd-mcp-guide.md`


---
## 📦 可用资源

- `skill://bk-monitor-tapd-dev/references/branch-workflow.md`
- `skill://bk-monitor-tapd-dev/references/dev-plan-template.md`
- `skill://bk-monitor-tapd-dev/references/mcp-calls.md`
- `skill://bk-monitor-tapd-dev/references/tapd-mcp-guide.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
