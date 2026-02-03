---
name: bk-monitor-tapd-summary
description: 获取 TAPD 待办单据并生成归纳总结。当用户提到获取 TAPD 单据、查看待办时使用。
---

# TAPD 待办单据汇总器

## 工作流程

### Step 1: 获取待办

对 workspace `10158081`（蓝鲸监控）和 `70093903`（蓝鲸开发工具）分别调用：

- stories_get（需求）
- bugs_get（缺陷）
- tasks_get（任务）

只获取未开发完成状态的单据。

### Step 2: 数据处理

1. 合并所有单据，标记类型和项目
2. 按紧急程度分组（已逾期 > 本周内 > 本月内 > 远期）
3. 组内按优先级排序
4. 最多处理 60 条

### Step 3: 生成链接

详见 summary-template.md

### Step 4: 输出汇总

按 summary-template.md 模板输出。

## 注意事项

- 只展示未开发完成的单据
- 去重：同一单据只显示一次
- 缺陷无 due 字段时归类到「远期」

---

## 📦 可用资源

- `skill://bk-monitor-tapd-summary/references/mcp-calls.md`
- `skill://bk-monitor-tapd-summary/references/status-mapping.md`
- `skill://bk-monitor-tapd-summary/references/summary-template.md`


---
## 📦 可用资源

- `skill://bk-monitor-tapd-summary/references/mcp-calls.md`
- `skill://bk-monitor-tapd-summary/references/status-mapping.md`
- `skill://bk-monitor-tapd-summary/references/summary-template.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
