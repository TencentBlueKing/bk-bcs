---
name: chat-x-contribute
description: 帮助开发者上手 @blueking/chat-x 组件库开发。当用户需要了解 chat-x、开发新组件时使用。
---

# Chat-X 组件库开发指南

## 项目概述

`@blueking/chat-x` 是蓝鲸 AI Chat 组件库。

**技术栈**：Vue 3 + TypeScript + bkui-vue + Vite

**核心特性**：流式渲染、Markdown、代码高亮、LaTeX、Mermaid

## 开发环境

```bash
pnpm install                           # 安装依赖
pnpm --filter @blueking/chat-x dev     # 启动开发
pnpm --filter @blueking/chat-x build   # 构建
```

## 项目结构

```
packages/chat-x/src/
├── components/    # 组件
├── composables/   # 组合式函数
├── types/         # 类型定义
└── index.ts       # 入口
```

## 开发新组件

1. 创建文件：`src/components/new-component/`
2. 导出组件：`src/components/index.ts`
3. 添加类型：`src/types/`
4. 编写文档：`wikis/components/`

## 最佳实践

- 使用 `<script setup>` 语法
- Props 使用 TypeScript 接口
- 文件名使用 kebab-case
- HTML 内容使用 DOMPurify 过滤

---

## 📦 可用资源

- `skill://chat-x-contribute/references/component-api.md`
- `skill://chat-x-contribute/references/development-workflow.md`
- `skill://chat-x-contribute/references/QA.md`
- `skill://chat-x-contribute/references/style-guide.md`
- `skill://chat-x-contribute/references/type-definitions.md`


---
## 📦 可用资源

- `skill://chat-x-contribute/references/QA.md`
- `skill://chat-x-contribute/references/component-api.md`
- `skill://chat-x-contribute/references/development-workflow.md`
- `skill://chat-x-contribute/references/style-guide.md`
- `skill://chat-x-contribute/references/type-definitions.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
