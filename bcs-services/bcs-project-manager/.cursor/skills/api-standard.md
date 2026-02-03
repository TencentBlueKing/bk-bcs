---
id: eng-api-standard
name: 统一网络请求封装 (Axios)
category: engineering
description: 蓝鲸双协议兼容的 Axios 封装，自动处理旧版和新版 HTTP 协议
tags: [axios, http, api, request, blueking]
updated_at: 2026-01-16
---

# 统一网络请求封装 (Axios)

蓝鲸体系处于 HTTP 协议升级过渡期，本封装自动兼容 **旧版协议** 和 **新版协议**。

## ⚠️ 核心规则

1. **统一导出**: 所有请求必须通过 `src/api/http.ts` 的实例发起
2. **自动剥壳**: 响应拦截器自动提取 `data` 字段
3. **自动登录**: 401 自动跳转登录页
4. **统一错误**: 错误信息统一通过 Message 提示

## 快速开始

```typescript
// src/api/http.ts
import axios from 'axios';
import { Message } from 'bkui-vue';

const http = axios.create({ baseURL: '/api', timeout: 60000 });

http.interceptors.response.use(
  (res) => {
    const { data } = res;
    // 旧版协议（有 code 字段）
    if (data.code !== undefined) {
      if (data.code !== 0) {
        Message({ theme: 'error', message: data.message });
        return Promise.reject(new Error(data.message));
      }
      return data.data;
    }
    return data.data ?? data;
  },
  (error) => {
    if (error.response?.status === 401) window.location.href = '/login';
    Message({ theme: 'error', message: error.message });
    return Promise.reject(error);
  }
);
export default http;
```

## 常见错误

| 错误 | 解决 |
|------|------|
| 401 循环跳转 | 登录页排除拦截器 |
| 数据双层嵌套 | 删除多余的 `.data` |

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 完整实现 | `skill://api-standard/references/full-implementation.md` |
| 协议迁移 | `skill://api-standard/references/protocol-migration.md` |


---
## 📦 可用资源

- `skill://api-standard/references/full-implementation.md`
- `skill://api-standard/references/protocol-migration.md`
- `skill://api-standard/assets/http.ts`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
