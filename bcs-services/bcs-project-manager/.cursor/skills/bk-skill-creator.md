---
id: engineering/bk-skill-creator
name: Skill 创建指南
category: engineering
description: 指导如何创建符合渐进式披露架构的 skill 文档
tags: [skill, knowledge, template, guide]
updated_at: 2026-01-23
---

# Skill 创建指南

## ⚠️ 核心规则

1. **SKILL.md ≤ 2KB**: 超出内容必须移到 `references/`
2. **必须包含 Front Matter**: id, name, category, description, tags, updated_at
3. **按需加载引导**: 详细内容通过 `skill://` URI 引用
4. **存放位置**: 所有 skill 必须放在 `bkui-knowledge/knowledge/skills/` 目录下

## 快速开始

```bash
cp -r knowledge/skills/.template knowledge/skills/your-skill-id
vim knowledge/skills/your-skill-id/SKILL.md
bash scripts/validate-skill.sh your-skill-id
```

详细步骤: `skill://bk-skill-creator/references/quick-start.md`

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 快速开始 | `skill://bk-skill-creator/references/quick-start.md` |
| 目录结构 | `skill://bk-skill-creator/references/structure-guide.md` |
| 常见错误 | `skill://bk-skill-creator/references/common-mistakes.md` |
| 检查清单 | `skill://bk-skill-creator/references/skill-checklist.md` |
| 写作技巧 | `skill://bk-skill-creator/references/writing-tips.md` |


---
## 📦 可用资源

- `skill://bk-skill-creator/references/common-mistakes.md`
- `skill://bk-skill-creator/references/quick-start.md`
- `skill://bk-skill-creator/references/skill-checklist.md`
- `skill://bk-skill-creator/references/structure-guide.md`
- `skill://bk-skill-creator/references/writing-tips.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
