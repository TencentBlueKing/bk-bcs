---
id: eng-bkui-cheatsheet
name: BKUI 组件属性速查表
category: engineering
description: 高频易错属性映射和避坑指南
tags: [bkui, cheatsheet, props, pitfalls]
updated_at: 2026-01-16
---

# BKUI 组件属性速查表

还原设计稿时，严格遵守以下属性映射，**严禁使用 ElementUI/AntD 的属性名**。

## ⚠️ 高频错误速查

| 组件 | 错误 | 正确 |
|------|------|------|
| Button | `type="primary"` | `theme="primary"` |
| Input | `prefix-icon="xx"` | `<template #prefix>` |
| Icon | `<i class="bk-icon...">` | `import { Plus } from 'bkui-vue/lib/icon'` |
| Dialog | `v-model="show"` | `v-model:isShow="show"` |
| Table | 忘记 `remote-pagination` | 远程分页必须加 |
| DatePicker | `shortcuts.value: []` | `shortcuts.value: () => []` |

## 基础组件

```vue
<!-- Button -->
<bk-button theme="primary">主要</bk-button>
<bk-button text theme="primary">编辑</bk-button>

<!-- Icon -->
<script setup>
import { Plus, Search } from 'bkui-vue/lib/icon';
</script>

<!-- Input -->
<bk-input v-model="value">
  <template #prefix><Search /></template>
</bk-input>
```

## 表格开发

```vue
<bk-table :data="list" :pagination="pagination" remote-pagination>
  <bk-table-column label="名称" prop="name" />
  <bk-table-column label="操作">
    <template #default="{ row }">
      <bk-button text theme="primary">编辑</bk-button>
    </template>
  </bk-table-column>
</bk-table>
```

> 分页对象: `{ current, limit, count }`

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 复杂组件 | `skill://bkui-cheatsheet/references/complex-components.md` |


---
## 📦 可用资源

- `skill://bkui-cheatsheet/references/complex-components.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
