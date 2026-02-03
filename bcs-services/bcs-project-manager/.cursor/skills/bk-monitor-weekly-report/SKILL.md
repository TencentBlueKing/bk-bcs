---
name: bk-monitor-weekly-report
description: 通过 GitHub MCP 获取用户 PR 并生成周报。当用户提到生成周报、写周报时使用。
---

# 周报生成器

## 工作流程

### Step 1: 获取用户信息

调用 `toolbase_github_get_my_user`，记录 `login` 字段。

### Step 2: 搜索 PR

对以下仓库搜索用户最近 7 天的 PR：

- TencentBlueKing/bk-monitor
- TencentBlueKing/bk-monitor-grafana-plugins
- TencentBlueKing/bkui-vue2
- TencentBlueKing/bk-weweb
- TencentBlueKing/bk-monitor-grafana

调用 `toolbase_github_search_issues`。

### Step 3: PR 分类

- 需求：feat、feature、新增、添加
- Issues：fix、bug、修复、优化
- 其他：不匹配以上规则

优先级：Labels > 标题关键词

### Step 4: 生成周报

按 weekly-template.md 输出。

---

## 📦 可用资源

- `skill://bk-monitor-weekly-report/references/classification-rules.md`
- `skill://bk-monitor-weekly-report/references/weekly-template.md`


---
## 📦 可用资源

- `skill://bk-monitor-weekly-report/references/classification-rules.md`
- `skill://bk-monitor-weekly-report/references/weekly-template.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
