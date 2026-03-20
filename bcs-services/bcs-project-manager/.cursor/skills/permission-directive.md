---
id: eng-permission-auth
name: 前端权限控制方案 (IAM)
category: engineering
description: 基于蓝鲸 IAM 的前端鉴权方案，包含 v-authority 指令实现、权限组件封装及无权限交互规范。
tags: [iam, permission, authority, directive, vue3]
updated_at: 2026-01-09
---

# 前端权限控制方案 (IAM)

在蓝鲸体系中，权限控制不仅仅是“显示/隐藏”，更重要的是**“发现与申请”**。我们推荐使用**“置灰 + 提示申请”**的交互模式。

## 1. 核心指令 `v-authority`

这个指令会自动处理点击拦截、样式置灰和申请弹窗的唤起。

**使用方式：**
```html
<bk-button v-authority="{ permission: hasAuth, actionId: 'host_edit' }">
  编辑
</bk-button>
```

**指令功能：**
- 有权限 → 正常交互
- 无权限 → 置灰 + 点击触发申请弹窗 + Tooltip 提示

> 📦 获取完整指令实现：`skill://permission-directive/assets/authority-directive.ts`

## 2. 鉴权组件 `AuthButton`

对于需要更多自定义的场景，可封装鉴权按钮组件。

## 3. 路由级鉴权

在 `vue-router` 的 `beforeEach` 中处理页面级权限。

```typescript
router.beforeEach(async (to, from, next) => {
  const meta = to.meta as any;
  if (meta.auth) {
    const hasAuth = await checkPageAuth(meta.authAction);
    if (!hasAuth) {
      next({ name: '403', query: { action: meta.authAction } });
      return;
    }
  }
  next();
});
```


---
## 📦 可用资源

- `skill://permission-directive/assets/authority-directive.ts`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
