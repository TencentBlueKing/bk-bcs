---
name: chat-x-unit-test
description: 帮助开发者为 Vue 3 组件编写单元测试。当用户需要添加单元测试时使用。
---

# Vue 组件单元测试指南

## 测试规范

测试文件与组件同目录，命名为 `组件名.spec.ts`。

## 运行测试

```bash
pnpm --filter @blueking/chat-x test              # 所有测试
npx vitest run src/components/your-component     # 指定目录
```

## 基础模板

```typescript
import { defineComponent, h } from 'vue';
import { type VueWrapper, mount } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import YourComponent from './your-component.vue';

vi.mock('vue-tippy', () => ({ /* ... */ }));

describe('YourComponent', () => {
  let wrapper: VueWrapper;

  beforeEach(() => { vi.clearAllMocks(); });
  afterEach(() => { wrapper?.unmount(); });

  it('应该正确渲染', () => {
    wrapper = mount(YourComponent);
    expect(wrapper.find('.your-component').exists()).toBe(true);
  });
});
```

## 测试分类

- 渲染测试：组件正确渲染
- Props 测试：属性传递和响应
- 事件测试：emit 事件
- Slot 测试：插槽内容

## 同步更新原则

修改组件代码时必须同步更新测试用例。

---

## 📦 可用资源

- `skill://chat-x-unit-test/references/mock-patterns.md`
- `skill://chat-x-unit-test/references/test-qa.md`
- `skill://chat-x-unit-test/references/test-strategies.md`


---
## 📦 可用资源

- `skill://chat-x-unit-test/references/mock-patterns.md`
- `skill://chat-x-unit-test/references/test-qa.md`
- `skill://chat-x-unit-test/references/test-strategies.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
