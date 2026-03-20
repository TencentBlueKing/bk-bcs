---
name: chat-x-custom-component
description: 帮助开发者创建 @blueking/chat-x 自定义 message 组件。当用户需要开发自定义消息类型时使用。
---

# Chat-X 自定义 Message 组件开发

## 概述

自定义 message 组件允许你扩展内置消息类型、渲染任意 UI 组件。

## 开发流程

```
1. 声明类型扩展 → 2. 创建组件 → 3. 集成到 MessageContainer → 4. 测试
```

## 核心机制

### 类型扩展

```typescript
import { type BaseMessage } from '@blueking/chat-x';

declare global {
  interface AIBluekingMessageMap {
    custom: BaseMessage<'custom', { content: string }>;
  }
}
```

### MessageSlot 机制

```typescript
import { useMessageSlotId } from '@blueking/chat-x';
const { messageSlotId } = useMessageSlotId();
// 用于 Teleport: <Teleport :to="messageSlotId">
```

## 常见场景

- ECharts 图表消息
- bkui-vue 表格消息
- 动态表单消息
- 卡片列表消息

---

## 📦 可用资源

- `skill://chat-x-custom-component/references/api-reference.md`
- `skill://chat-x-custom-component/references/best-practices.md`
- `skill://chat-x-custom-component/references/full-example.md`
- `skill://chat-x-custom-component/references/integration-guide.md`
- `skill://chat-x-custom-component/references/type-extension.md`


---
## 📦 可用资源

- `skill://chat-x-custom-component/references/QA.md`
- `skill://chat-x-custom-component/references/api-reference.md`
- `skill://chat-x-custom-component/references/best-practices.md`
- `skill://chat-x-custom-component/references/component-templates.md`
- `skill://chat-x-custom-component/references/full-example.md`
- `skill://chat-x-custom-component/references/integration-guide.md`
- `skill://chat-x-custom-component/references/type-extension.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
