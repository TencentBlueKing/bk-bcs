---
id: qual-unit-test
name: Vue 3 组件单元测试指南 (Vitest)
category: quality
description: 使用 Vitest 和 Vue Test Utils 的标准模版与断言技巧
tags: [test, vitest, vue-test-utils, quality]
updated_at: 2026-01-09
---

# Vue 3 组件单元测试指南

采用 **Vitest** 作为测试运行器，配合 **@vue/test-utils** 进行组件测试。

## 环境准备

```json
"devDependencies": {
  "vitest": "^1.0.0",
  "@vue/test-utils": "^2.4.0",
  "jsdom": "^23.0.0"
}
```

## 测试模版

```typescript
import { mount } from '@vue/test-utils';
import { describe, it, expect, vi } from 'vitest';
import MyComponent from './MyComponent.vue';

describe('MyComponent.vue', () => {
  it('renders properly', () => {
    const wrapper = mount(MyComponent, {
      props: { title: 'Hello' }
    });
    expect(wrapper.text()).toContain('Hello');
  });

  it('emits event on click', async () => {
    const wrapper = mount(MyComponent);
    await wrapper.find('button').trigger('click');
    expect(wrapper.emitted()).toHaveProperty('submit');
  });
});
```

## 常用技巧

```typescript
// Mock 第三方组件
const wrapper = mount(MyComponent, {
  global: { stubs: { 'bk-table': true } }
});

// Mock API（文件顶部）
vi.mock('@/api/user', () => ({
  getUserInfo: vi.fn(() => Promise.resolve({ id: 1 }))
}));
```

## 运行测试

```bash
npm run test:unit
```

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 完整测试示例 | `skill://unit-testing/assets/component.spec.ts` |


---
## 📦 可用资源

- `skill://unit-testing/assets/component.spec.ts`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
