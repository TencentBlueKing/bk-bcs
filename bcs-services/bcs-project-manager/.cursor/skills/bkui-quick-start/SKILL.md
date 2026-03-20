---
id: bkui-quick-start
name: BKUI 快速入门
category: engineering
description: 蓝鲸前端知识库入口指南，包含规范、索引和工作流程
tags: [bkui, 规范, 索引, 入门, vue3]
updated_at: 2026-01-23
---

# BKUI 快速入门

> 蓝鲸前端知识库入口指南。

## 强制规范

- **组件库**: bkui-vue (前缀 `bk-`)
- **语法**: Vue 3 `<script setup lang="ts">`
- **样式**: MagicBox 原子类 (mt10, mb20)
- **布局**: 必须使用 `bk-navigation`

## 常见错误 (必须避免)

| 组件 | 错误写法 | 正确写法 |
|------|----------|----------|
| bk-navigation | `:default-open-keys` | `default-open` |
| bk-menu | `:default-open-keys` | `:opened-keys` |
| bk-dialog | `v-model` | `v-model:isShow` |

## 高优先级组件

- `bk-navigation` - 布局组件，易出错
- `bk-menu` - 与 navigation 配合
- `bk-table` - 列表页核心组件
- `bk-form` - 表单验证
- `bk-dialog` - v-model:isShow

## 工作流程

1. **分析需求** → 确定需要哪些资源
2. **布局组件** → `get_component_api({ componentName: 'navigation' })`
3. **模板代码** → `get_skill({ skillId: 'bkui-builder' })`

## 触发条件

遇到 bk- 前缀组件、bkui-vue、蓝鲸前端、设计稿还原时使用。

---

## 按需加载资源

- `skill://bkui-quick-start/references/skills-index.md` - 完整 Skills 索引
- `skill://bkui-quick-start/references/components-list.md` - 组件完整列表


---
## 📦 可用资源

- `skill://bkui-quick-start/references/components-list.md`
- `skill://bkui-quick-start/references/skills-index.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
