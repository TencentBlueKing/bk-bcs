---
id: sec-js-check
name: JavaScript 安全审查
category: security
description: 检查 XSS、CSRF、原型污染等安全问题，支持 React/Vue/Angular
tags: [security, javascript, xss, csrf, frontend]
updated_at: 2026-01-20
allowed-tools: [Read, Grep, Glob, Shell]
---

# JavaScript 安全审查

## ⚠️ 核心规则

1. **永不信任用户输入** - 所有用户数据需校验和转义
2. **默认安全** - 使用安全 API（textContent 而非 innerHTML）
3. **纵深防御** - 多层安全措施，不依赖单一防护

## 快速开始

```bash
/js-security-check                    # 智能扫描 src 目录
/js-security-check file src/xxx.vue   # 扫描指定文件
/js-security-check report             # 生成详细报告
```

## 问题分级

| 前缀 | 含义 | 处理方式 |
|------|------|----------|
| `🔴 严重` | 可被直接利用 | 阻止发布 |
| `🟡 中等` | 需特定条件 | 尽快修复 |
| `⚪ 建议` | 最佳实践 | 可选优化 |

## 检查维度

| 维度 | 检查项 |
|------|--------|
| DOM 安全 | innerHTML、document.write、insertAdjacentHTML |
| URL 安全 | Open Redirect、javascript: scheme |
| 代码执行 | eval、new Function、setTimeout(string) |
| 原型污染 | __proto__、constructor、Object.assign |
| 存储安全 | localStorage 敏感信息、Cookie 属性 |
| 框架安全 | v-html、dangerouslySetInnerHTML |

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 完整检查清单 | `skill://js-security-check/references/checklist.md` |
| 修复示例 | `skill://js-security-check/references/fix-examples.md` |
| 评分标准 | `skill://js-security-check/references/scoring-standard.md` |


---
## 📦 可用资源

- `skill://js-security-check/references/checklist.md`
- `skill://js-security-check/references/fix-examples.md`
- `skill://js-security-check/references/report-format.md`
- `skill://js-security-check/references/scoring-standard.md`
- `skill://js-security-check/references/security-toolkit.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
