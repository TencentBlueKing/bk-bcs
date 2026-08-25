# 前端开发规范

> 通用前端开发规范文档，适用于基于 Vue 3 + TypeScript + Vite 技术栈的中后台项目。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应章节；禁止默认全文灌入。完整条款见正文；落地优先级见文末「规范落地优先级」。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 写死业务目录；裸用 axios；手改 Generated SDK | 二、六 |
| 必须 | 请求走统一封装；状态用 Pinia；Store 解构用 `storeToRefs` | 四、五、七 |
| 验证 | 跑 `package.json` 已有 lint/build/typecheck（无则不得虚构通过） | 九、文末 P0 |

---

## 一、技术栈要求

| 技术 | 版本要求 | 用途 |
|------|---------|------|
| Vue | 3.x（以 package.json 为准） | 前端框架 |
| TypeScript | 以 package.json 为准 | 类型安全 |
| Vite | 以 package.json 为准（本预设要求存在 vite.config.*） | 构建工具 |
| Pinia | 以 package.json 为准 | 状态管理 |
| vue-router | 以 package.json 为准 | 路由管理 |
| HTTP 客户端 | 以项目现有封装为准 | 网络请求 |

---

## 二、项目结构

> **禁止写死业务目录。** 以下仅为常见 Vite + Vue 3 布局的**示例**，实际路径以探测到的前端根（含 `package.json` 且存在 `vite.config.*` 的目录）为准。

- 前端根：探测得到的目录（可能是 `src/dashboard-front`、`apps/ui` 等，**不一定是** `web/src`）。
- 请求封装 / 路由 / 页面 / store：沿用该目录现有命名（`http`/`api`、`views`/`pages`、`store`/`stores` 等均可）。
- **不要**假设存在 `services/generated/` 或强制「所有请求必须经 `src/api/http.ts`」——以仓库真实入口为准。


## 三、编码规范

### 3.1 组件编写规范

```vue
<script setup lang="ts">
// 1. 导入区
import { ref, computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/stores/user';
import type { ResourceItem } from '@/types/resource';

// 2. Props & Emits
interface Props {
  id: string;
  title?: string;
}
const props = withDefaults(defineProps<Props>(), {
  title: '默认标题',
});
const emit = defineEmits<{
  (e: 'update', value: string): void;
  (e: 'delete', id: string): void;
}>();

// 3. Store
const userStore = useUserStore();
const { userInfo } = storeToRefs(userStore);

// 4. 响应式状态
const loading = ref(false);
const list = ref<ResourceItem[]>([]);

// 5. 计算属性
const isEmpty = computed(() => list.value.length === 0);

// 6. 方法
const fetchData = async () => {
  loading.value = true;
  try {
    list.value = await resourceService.list(props.id);
  } finally {
    loading.value = false;
  }
};

// 7. 生命周期
onMounted(fetchData);
</script>

<template>
  <!-- 模板内容 -->
</template>

<style scoped>
/* 样式 */
</style>
```

**要点：**

- 统一使用 `<script setup lang="ts">` 语法
- Props 使用 TypeScript 接口定义 + `withDefaults`
- Emits 使用泛型函数签名定义
- 代码区块按"导入 → Props/Emits → Store → 状态 → 计算属性 → 方法 → 生命周期"排列

### 3.2 TypeScript 严格规范

| 规则 | 说明 |
|------|------|
| 启用 `strictTemplates` | `vue-tsc` 严格模板检查，捕获未定义组件和错误属性 |
| 避免无必要的 `any` | 以项目 ESLint 规则为准；不得写死比仓库更严的禁令 |
| Props 类型化 | 所有 Props 使用 TypeScript 接口定义 |
| Ref 泛型 | `ref<Type>()` 必须指定泛型参数 |
| 路由参数类型化 | 使用 `typed-router` 或显式断言 |

### 3.3 命名约定

| 类型 | 规则 | 示例 |
|------|------|------|
| 组件文件 | PascalCase.vue | `UserList.vue` |
| 组合式函数 | use + PascalCase | `usePolling.ts` |
| Store 文件 | 小写名词 | `user.ts` |
| Store 导出 | use + PascalCase + Store | `useUserStore` |
| Service 文件 | 小写模块名 | `resource.ts` |
| 类型文件 | 小写模块名 | `resource.ts` |
| 工具函数 | camelCase | `formatDate.ts` |
| 常量 | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |
| CSS class | kebab-case | `.page-header` |

---

## 四、状态管理规范（Pinia）

### 4.1 推荐写法：Setup Store

```typescript
// src/stores/user.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

export const useUserStore = defineStore('user', () => {
  // State
  const userInfo = ref<UserInfo | null>(null);
  const loading = ref(false);

  // Getter
  const username = computed(() => userInfo.value?.name ?? '');
  const isLoggedIn = computed(() => !!userInfo.value);

  // Action
  const fetchUserInfo = async () => {
    if (userInfo.value) return userInfo.value;
    loading.value = true;
    try {
      userInfo.value = await http.get('/user/info');
    } finally {
      loading.value = false;
    }
    return userInfo.value;
  };

  const logout = () => {
    userInfo.value = null;
  };

  return { userInfo, loading, username, isLoggedIn, fetchUserInfo, logout };
});
```

### 4.2 使用规范

| 规则 | 说明 |
|------|------|
| 使用 `storeToRefs` | 解构 state/getter 时保持响应性 |
| Action 直接调用 | `store.action()` 不需要 `storeToRefs` |
| 单例模式 | 不要重复 `new`，直接 `useXxxStore()` |
| 异步初始化 | 在 `App.vue` 或路由守卫中完成 |

```typescript
// ✅ 正确
const userStore = useUserStore();
const { userInfo, loading } = storeToRefs(userStore);
userStore.fetchUserInfo();

// ❌ 错误：解构丢失响应性
const { userInfo } = useUserStore(); // 不是响应式！
```

---

## 五、网络请求规范

### 5.1 统一 Axios 封装

```typescript
// src/api/http.ts
import axios from 'axios';

const http = axios.create({
  baseURL: '/api',
  timeout: 60000,
});

// 响应拦截器：自动剥壳
http.interceptors.response.use(
  (res) => {
    const { data } = res;
    // 兼容旧版协议（含 code 字段）
    if (data.code !== undefined) {
      if (data.code !== 0) {
        showError(data.message);
        return Promise.reject(new Error(data.message));
      }
      return data.data;
    }
    // 新版协议（直接返回 data）
    return data.data ?? data;
  },
  (error) => {
    if (error.response?.status === 401) {
      window.location.href = '/login';
    }
    showError(error.message);
    return Promise.reject(error);
  },
);

export default http;
```

### 5.2 核心规则

| 规则 | 说明 |
|------|------|
| 统一入口 | 所有请求应通过项目统一的 HTTP 封装发起（路径以仓库为准） |
| 自动剥壳 | 拦截器自动提取 `data` 字段，业务代码直接拿数据 |
| 401 自动跳转 | 未登录自动跳转登录页 |
| 统一错误提示 | 错误信息统一通过 Toast/Message 提示 |
| 禁止裸用 `axios` | 禁止在业务代码中直接 `import axios from 'axios'` |

---

## 六、三层代码架构

前端接入后端接口时，必须遵循**三层分离**：

```
┌────────────────────────────────────────────────┐
│  Pages 层（页面组件）                             │
│  - 组织 UI 和交互逻辑                            │
│  - 调用 Services 层获取数据                       │
│  - 不直接调用 generated 代码                      │
├────────────────────────────────────────────────┤
│  Services 层（桥接层）                            │
│  - services/{module}.ts                         │
│  - 封装 generated 调用                           │
│  - 类型转换（snake_case → camelCase）            │
│  - 错误处理、业务工具函数                         │
├────────────────────────────────────────────────┤
│  Generated 层（若项目有 codegen）                  │
│  - 目录名以仓库为准（不一定是 services/generated） │
│  - 由 Proto/OpenAPI 等自动生成时禁止手改           │
└────────────────────────────────────────────────┘
```

> 若项目**没有** codegen / generated 层，忽略本图 Generated 行，按现有 `http`/`api` 组织即可。

### 各层职责

| 层级 | 职责 | 修改规则 |
|------|------|---------|
| Generated 层（可选） | 自动生成的 SDK | 有则禁止手改，仅重新生成 |
| Services / API 桥接 | 封装请求、类型适配 | 可修改 |
| Pages / Views | UI | 自由修改 |

### 旧实现迁移规则

- 旧接口函数标记 `@deprecated`，不直接删除
- 新实现放在同一文件或新文件中
- 迁移完成后全局搜索确认无引用再清理

---

## 七、Vue 3 最佳实践

### 7.1 响应式规则

| 规则 | 说明 |
|------|------|
| `ref` vs `reactive` | 优先使用 `ref`，避免 `reactive` 丢失响应性 |
| `storeToRefs` | 解构 Store 时必须使用 |
| `watch` 深度监听 | 数组用 `{ deep: true }`，Vue 3.5+ 支持 `watch.length` |
| `computed` | 只读派生数据优先用 `computed` |

### 7.2 组件通信

| 场景 | 方案 |
|------|------|
| 父 → 子 | Props |
| 子 → 父 | Emits |
| 跨层级 | Provide/Inject 或 Pinia |
| 兄弟组件 | 共同父组件提升状态，或 Pinia |

### 7.3 性能优化

| 技术 | 适用场景 |
|------|---------|
| `v-memo` | 大列表中不频繁变化的项 |
| `shallowRef` | 大对象只需整体替换 |
| 虚拟滚动 | 列表 > 100 条时 |
| 路由懒加载 | 所有页面组件 |
| 组件异步加载 | 弹窗、抽屉等低频组件 |

---

## 八、UI 组件使用规范

### 8.1 基本原则

| 原则 | 说明 |
|------|------|
| 组件优先 | 优先使用 UI 库组件，禁止使用原生 `<table>`、`<button>` 等 |
| 布局组件化 | 页面骨架必须使用 UI 库提供的布局组件 |
| 样式统一 | 使用 UI 库提供的原子类或 Design Token |
| Icon 规范 | 从 UI 库包导入，禁止使用 CDN 图标 |

### 8.2 页面布局模式

| 页面类型 | 布局方案 |
|---------|---------|
| 列表页 | 搜索栏 + 操作区 + 表格 + 分页 |
| 详情页 | 面包屑 + 信息卡片 + Tab 面板 |
| 表单页 | 分步表单 / 单页表单 + 提交按钮 |
| 仪表盘 | 统计卡片 + 图表网格 |

### 8.3 常见错误

| 错误 | 正确做法 |
|------|---------|
| 手写 div 布局 | 使用 UI 库 Navigation/Layout 组件 |
| 内联样式过多 | 使用 CSS class 或 UI 库原子类 |
| 组件属性写错（如拼写错误） | 查阅组件库文档，开启 `strictTemplates` 检查 |
| 直接操作 DOM | 使用 Vue ref + 组件 API |

---

## 九、质量保证

### 9.1 提交前验证（按仓库实际脚本）

在前端根目录根据 `package.json` scripts 执行：

- 若有 `lint` / `typecheck` / `build`：运行对应脚本。
- 若有 `test` / `test:unit`：运行对应脚本。
- 若**无**测试脚本：**不得**要求 `vitest`/`jest`，也不得声称「测试已通过」。
- 仅当依赖中存在 `vue-tsc` 或脚本声明了类型检查时，才运行类型检查。

### 9.2 单元测试规范

| 规则 | 说明 |
|------|------|
| 有测试框架时 | 为变更相关的工具函数与关键交互补充测试 |
| 无测试框架时 | 在验证文档中说明手工验证范围，禁止虚构覆盖率门禁 |

### 9.3 代码审查清单

- [ ] TypeScript 类型完整，无 `any`
- [ ] 网络请求通过统一封装发起
- [ ] 新状态通过 Pinia Store 管理
- [ ] 组件 Props/Emits 类型化
- [ ] 无直接操作 DOM
- [ ] 错误边界处理完善
- [ ] 路由懒加载
- [ ] 三件套全部通过

---

## 十、常见陷阱与避坑

| # | 陷阱 | 解决方案 |
|---|------|---------|
| 1 | Store 解构丢失响应性 | 使用 `storeToRefs()` |
| 2 | `watch` 首次不执行 | 添加 `{ immediate: true }` |
| 3 | 异步组件未处理加载态 | 使用 `<Suspense>` 或手动 loading |
| 4 | 路由参数变化页面不刷新 | `watch(route.params, ...)` 或设置 `:key` |
| 5 | v-model 在自定义组件不工作 | 使用 `defineModel()` (Vue 3.4+) |
| 6 | CSS scoped 穿透失败 | 使用 `:deep()` 选择器 |
| 7 | int64 类型 JSON 精度丢失 | 后端 JSON 中 int64 序列化为 string，前端做 coerce |
| 8 | 401 循环跳转 | 登录页排除拦截器 |
| 9 | 数据双层嵌套 | 检查拦截器是否多剥一层 `.data` |
| 10 | 列表不刷新 | 操作成功后手动调用加载 + 轮询兜底 |

---

## 规范落地优先级

> 文首「本次必读决策表」为 P0/P1 快速入口；下表为完整落地优先级。

| 优先级 | 条款类型 | 落地方式 |
|--------|----------|----------|
| P0 机器门禁 | ESLint/Stylelint/Prettier；`package.json` 已有 `lint`/`build`/`typecheck` | CI 或提交前跑对应 script；无 script 不得虚构通过 |
| P0 机器门禁 | 密钥/Token 不进仓 | 仓库 secret 扫描 / pre-commit（若有） |
| P1 Agent 必读 | 组件/目录约定、请求封装、Store、路由鉴权 | 改前端前 Read 本文件相关节；IDE `standards-frontend` 仅督促加载 |
| P2 参考 | 布局惯例、非强制性能建议 | 可偏离，偏离时在 MR 说明 |

完整条款以本文件为准；短 Rules / AGENTS 门闩不含正文复述。
