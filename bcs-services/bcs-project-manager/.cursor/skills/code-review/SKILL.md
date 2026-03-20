---
id: qual-code-review
name: 代码评审专家
category: quality
description: 基于 Google Code Review 指南的代码评审技能
tags: [code-review, quality, git, pr, mr]
updated_at: 2026-01-20
allowed-tools: [Read, Grep, Glob, Shell]
---

# 代码评审专家

## ⚠️ 核心规则

1. **追求持续改进，而非完美** - 倾向于批准能提升代码健康状态的变更
2. **对事不对人** - 有建设性的反馈，保持礼貌尊重
3. **解释为什么** - 帮助开发者理解原因

## 快速开始

```bash
/code-review              # 智能评审（自动检测变更范围）
/code-review staged       # 评审暂存区（提交前检查）
/code-review last-commit  # 评审最近一次提交
```

> 默认按优先级检测：暂存区 → 工作区 → 最近提交

## 问题分级

| 前缀 | 含义 | 处理 |
|------|------|------|
| `[必须]` | 严重问题 | 阻止合入 |
| `[建议]` | 改进建议 | 讨论后决定 |
| `[Nit]` | 小问题 | 可忽略 |

## 检查维度

| 维度 | 核心检查项 |
|------|-----------|
| 设计 | 代码归属、系统集成、无过度工程 |
| 功能 | 行为符合预期、边缘情况已处理 |
| 复杂度 | 代码可简化、易于理解 |
| 测试 | 有自动化测试、测试设计良好 |
| 安全 | 无 XSS、输入校验、敏感数据安全 |
| 性能 | 无内存泄漏、大列表虚拟滚动 |

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 完整检查清单 | `skill://code-review/references/checklist.md` |
| Git 场景指南 | `skill://code-review/references/git-scenarios.md` |
| 常见错误规则 | `skill://code-review/references/scoring-standard.md` |


---
## 📦 可用资源

- `skill://code-review/references/checklist.md`
- `skill://code-review/references/git-scenarios.md`
- `skill://code-review/references/report-examples.md`
- `skill://code-review/references/report-format.md`
- `skill://code-review/references/scoring-standard.md`
- `skill://code-review/references/writing-guidelines.md`
- `skill://code-review/assets/pre-commit-review.sh`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
