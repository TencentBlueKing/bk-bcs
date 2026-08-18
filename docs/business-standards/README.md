# 业务规范（Business Standards）

> 本目录存放**项目自定义**的业务领域规范（领域语言、验收口径、特殊流程等）。
> harness-generating 只在目录不存在时创建本骨架；**harness-gardening 永不覆写本目录**。

Agent **按 tags / scenarios 选择性加载**，禁止默认把本目录全部灌入上下文。

## Frontmatter 元数据

业务规范 Markdown 建议在文件头使用 YAML frontmatter：

```yaml
---
tags: ["billing", "cluster"]          # 字符串数组：领域/模块标签
scenarios: ["create-cluster"]         # 字符串数组：适用场景
---
```

- `tags`：用于按模块过滤（例如只加载与当前改动目录相关的条目）
- `scenarios`：用于按任务类型过滤（例如「扩缩容」「发布」）

## 规范索引

| 文档 | tags | scenarios | 说明 |
|------|------|-----------|------|
| （示例）`example-domain.md` | `["example"]` | `["onboarding"]` | 示例行，新增规范后替换 |

当前尚无业务规范文件。新增时请更新本表。

## 与 `docs/standards/` 的区别

| 目录 | 权威来源 | 园艺行为 |
|------|---------|---------|
| `docs/standards/` | harness 预设库，单向同步 | 可覆写以对齐预设 |
| `docs/business-standards/` | 本仓库业务 Owner | **永不覆写** |
