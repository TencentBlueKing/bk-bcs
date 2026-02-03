---
id: perf-virtual-list
name: 长列表虚拟滚动优化方案
category: performance
description: 处理超过 1000 条数据的大型列表渲染时的性能优化方案，包含定高和不定高两种策略。
tags: [performance, vue3, virtual-scroll, list]
updated_at: 2026-01-09
---

# 长列表虚拟滚动优化方案

当列表数据量巨大（如日志列表、审计记录，n > 1000）时，直接渲染会导致 DOM 节点过多，页面卡顿。

## 核心原理
只渲染当前**可视区域 (Viewport)** 内的元素，加上缓冲区 (Buffer) 的元素。随着滚动条滚动，动态替换 DOM 内容。

## 推荐方案

### 1. 定高列表 (Item Height Fixed)

如果每一行高度固定（例如 40px），推荐使用轻量实现。

**使用方式：**
```html
<VirtualList :items="logList" :item-height="40" :container-height="400">
  <template #default="{ item }">
    <div class="log-item">{{ item.message }}</div>
  </template>
</VirtualList>
```

> 📦 获取完整组件实现：`skill://virtual-list/assets/VirtualList.vue`

### 2. 不定高列表 (Dynamic Height)
如果列表项高度不固定（如包含展开/收起、不同长度文本），计算逻辑会变得复杂。
**推荐库**: `vue-virtual-scroller` 的 `DynamicScroller` 组件。

```html
<template>
  <DynamicScroller
    :items="items"
    :min-item-size="54"
    class="scroller"
  >
    <template #default="{ item, index, active }">
      <DynamicScrollerItem
        :item="item"
        :active="active"
        :size-dependencies="[
          item.message,
        ]"
        :data-index="index"
      >
        <div class="message">{{ item.message }}</div>
      </DynamicScrollerItem>
    </template>
  </DynamicScroller>
</template>
```

## 注意事项
1. **滚动白屏**: 滚动过快时可能出现瞬间白屏，适当增加 `buffer` 缓冲区大小。
2. **搜索/筛选**: 虚拟列表与搜索过滤不冲突，只需对 `props.items` 进行 computed 过滤即可。


---
## 📦 可用资源

- `skill://virtual-list/assets/VirtualList.vue`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
