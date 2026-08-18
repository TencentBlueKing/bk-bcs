# 前端开发规范（Vue 3 + webpack）

> **启用前提：** `package.json` 含 `vue@>=3`，**无** `vite.config.*`，且存在 webpack 类构建（如 `webpack` 依赖、`vue.config.js`、`webpack.config.*`、或 `@blueking/cli-service-webpack`）。  
> 若存在 `vite.config.*`，应改用 [`frontend-vue3.md`](frontend-vue3.md)（Vite 预设）。  
> 对齐 [Vue 3 Style Guide](https://vuejs.org/style-guide/)、Composition API、Pinia（若项目使用）、Vue Router 4、Vue CLI / webpack 社区惯例。  
> **目录以探测到的前端根为准，禁止写死** `web/src` 等业务路径。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应章节；禁止默认全文灌入。完整条款见正文；落地优先级见文末「规范落地优先级」。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 写死目录；误用 Vite-only API；无 test 脚本时要求 Vitest/Jest | 二、十一 |
| 必须 | webpack/`vue.config.js` 构建；Pinia 用 `storeToRefs`；统一请求封装 | 三、五、七 |
| 验证 | 跑 `package.json` 已有 lint/build（无则不得虚构通过） | 十一、文末 P0 |

---

## 一、技术栈要求

| 技术 | 版本要求 | 用途 |
|------|---------|------|
| Vue | 3.x（以 package.json 为准） | 前端框架 |
| TypeScript | 以 package.json 为准（有则启用严格习惯） | 类型安全 |
| 构建 | webpack / Vue CLI 系（**非** Vite） | 本地开发与打包 |
| vue-router | 4.x（以项目为准） | 路由 |
| 状态 | Pinia 优先；存量可为 Vuex 4——以仓库为准 | 全局状态 |
| HTTP | axios 等；以项目封装为准 | 网络请求 |
| UI | 以项目依赖为准 | 组件库 |

---

## 二、项目结构

> **禁止写死业务目录。** 以下为 Vue CLI + Vue 3 + webpack 常见布局**示例**。

```
{frontend-root}/
├── public/
├── src/
│   ├── main.ts
│   ├── App.vue
│   ├── router/
│   ├── stores/|store/          # Pinia 或 Vuex（以仓库为准）
│   ├── views/|pages/
│   ├── components/
│   ├── api/|http/|services/
│   ├── composables/|hooks/     # 组合式函数（若有）
│   └── assets/
├── vue.config.js               # 或等价 webpack 配置
└── package.json
```

- 前端根可能是 `src/`、`frontend/`、`bk-user-web` 等。
- 请求 / 路由 / 页面 / store 命名沿用仓库；**不要**假设 Vite 专用目录或 `services/generated/`。
- 本地命令：`package.json` scripts（常见 `serve` / `dev` / `build` / `lint`），禁止写死他仓或强制 `vitest`。

---

## 三、编码规范

### 3.1 组件编写（Composition API）

优先 `<script setup lang="ts">`（与仓库既有 Options API 一致时可沿用 Options，不强制大爆炸式迁移）。

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/stores/user';
import type { ResourceItem } from '@/types/resource';

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

const userStore = useUserStore();
const { userInfo } = storeToRefs(userStore);

const loading = ref(false);
const list = ref<ResourceItem[]>([]);
const isEmpty = computed(() => list.value.length === 0);

const fetchData = async () => {
  loading.value = true;
  try {
    list.value = await resourceService.list(props.id); // 封装以仓库为准
  } finally {
    loading.value = false;
  }
};

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

- Props / Emits 类型化；区块顺序：导入 → Props/Emits → Store → 状态 → computed → methods → 生命周期  
- `@` 等 alias 以 `vue.config.js` / tsconfig 为准（webpack 项目常见），勿假设 Vite 默认  
- 样式穿透使用 `:deep()`；与 Vue 2 的 `/deep/` 区分  

### 3.2 TypeScript（若项目启用）

| 规则 | 说明 |
|------|------|
| 模板检查 | 有 `vue-tsc` / 类型检查脚本时启用；勿强加仓库没有的门禁 |
| 避免无必要的 `any` | 以项目 ESLint 为准 |
| `ref<T>()` | 复杂类型显式泛型 |
| 路由参数 | 显式校验或断言，避免信任 URL |

### 3.3 命名约定

| 类型 | 规则 | 示例 |
|------|------|------|
| 组件文件 | PascalCase.vue | `UserList.vue` |
| 组合式函数 | use + PascalCase | `usePolling.ts` |
| Pinia store 文件 | 小写名词 | `user.ts` |
| Store 导出 | use + PascalCase + Store | `useUserStore` |
| CSS class | kebab-case | `.page-header` |
| 常量 | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |

---

## 四、状态管理

### 4.1 Pinia（项目已使用时）

推荐 Setup Store；解构 state/getter 用 `storeToRefs`；action 直接调用。

```typescript
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

export const useUserStore = defineStore('user', () => {
  const userInfo = ref<UserInfo | null>(null);
  const loading = ref(false);
  const username = computed(() => userInfo.value?.name ?? '');

  const fetchUserInfo = async () => {
    if (userInfo.value) return userInfo.value;
    loading.value = true;
    try {
      userInfo.value = await http.get('/user/info'); // http 为项目封装
    } finally {
      loading.value = false;
    }
    return userInfo.value;
  };

  return { userInfo, loading, username, fetchUserInfo };
});
```

| 规则 | 说明 |
|------|------|
| `storeToRefs` | 解构保持响应性 |
| 单例 | 直接 `useXxxStore()` |
| 异步初始化 | `App.vue` 或路由守卫 |

### 4.2 Vuex 4（存量）

若仓库仍为 Vuex：遵循 namespaced modules、mutation 同步 / action 异步；新功能优先评估迁移 Pinia，**不**在未共识下强行双栈。

---

## 五、网络请求规范

| 规则 | 说明 |
|------|------|
| 统一入口 | 经项目 HTTP 封装；禁止业务裸 `import axios` 多实例 |
| 拦截器 | 鉴权、剥壳、统一错误、401 跳转（排除登录页） |
| 路径 | `api`/`http`/`services` 以仓库为准 |
| int64 | 大整数按后端约定用 string，避免精度丢失 |

---

## 六、分层与接口接入

```
Pages/Views  →  调用 Services/API 桥接  →  （可选）Generated SDK
```

| 层级 | 职责 | 修改规则 |
|------|------|---------|
| Generated（可选） | OpenAPI/Proto 生成 | 有则禁止手改，只重新生成 |
| Services / API | 封装请求、类型适配 | 可修改 |
| Pages / Views | UI 与交互 | 自由修改 |

无 codegen 时忽略 Generated 行，按现有 `http`/`api` 组织即可。旧接口可 `@deprecated`，确认无引用再删。

---

## 七、Vue 3 实践要点

### 7.1 响应式

| 规则 | 说明 |
|------|------|
| 优先 `ref` | 避免 `reactive` 解构丢响应 |
| `storeToRefs` | 解构 Pinia state/getter |
| `computed` | 只读派生 |
| `watch` | 需要时加 `deep` / `immediate` |

### 7.2 组件通信

| 场景 | 方案 |
|------|------|
| 父 → 子 | Props |
| 子 → 父 | Emits |
| 跨层级 | Provide/Inject 或 Pinia |
| 兄弟 | 状态提升或 Pinia |

### 7.3 性能

| 技术 | 适用 |
|------|------|
| 路由懒加载 | 页面组件 |
| 异步组件 | 弹窗、低频区块 |
| `shallowRef` | 大体量整体替换数据 |
| 虚拟滚动 | 超长列表（按项目库） |

---

## 八、构建与工程（webpack / Vue CLI）

| 规则 | 说明 |
|------|------|
| 配置 | `vue.config.js` / webpack 配置改代理、alias、`publicPath` |
| 环境变量 | `VUE_APP_*` 或项目约定；密钥不进仓 |
| 与 Vite 隔离 | **禁止**按 `vite.config`、Vite 插件、`import.meta.env` 习惯改本仓（除非仓库已混用且文档声明） |
| 依赖 | 注意 Vue 3 生态 peer；勿引入仅 Vue 2 的 UI 主版本 |

---

## 九、UI 组件使用规范

| 原则 | 说明 |
|------|------|
| 组件优先 | 优先项目 UI 库控件与布局 |
| 样式统一 | Design Token / 主题变量；减少魔法数 |
| Icon | 项目图标方案；慎 CDN |
| 布局模式 | 列表：搜索 + 表 + 分页；详情：面包屑 + 卡片/Tab；表单：分区 + 提交 |

---

## 十、与 Vite 预设的差异（Agent 须知）

| 项 | 本预设（webpack） | `frontend-vue3`（Vite） |
|----|-------------------|-------------------------|
| 构建入口 | `vue.config.js` / webpack | `vite.config.*` |
| 检测 | vue≥3 且无 vite.config + webpack 信号 | vue≥3 且有 vite.config |
| 环境变量 | 常见 `VUE_APP_*` / `process.env` | 常见 `import.meta.env` |
| 本地命令 | 读 package.json（serve/build） | 通常 vite 脚本（仍以项目为准） |
| 测试 | **禁止**强制某一测试运行器；有 scripts 才跑 | 同左——以 scripts 为准 |

---

## 十一、质量保证

### 11.1 提交前验证

在前端根按 `package.json` scripts：

- 有 `lint` / `typecheck` / `build` → 运行  
- 有 `test` / `test:unit` → 运行  
- **无**测试脚本 → 不得要求 Vitest/Jest，不得声称测试已通过  

### 11.2 代码审查清单

- [ ] Props/Emits 类型化；关键路径无不当 `any`  
- [ ] 请求走统一封装  
- [ ] Pinia 用 `storeToRefs`（若适用）  
- [ ] 无误用 Vite-only API  
- [ ] 路由懒加载；错误与 401 处理完善  
- [ ] 构建/lint（若有）通过  

---

## 十二、常见陷阱与避坑

| # | 陷阱 | 解决方案 |
|---|------|---------|
| 1 | Store 解构丢响应 | `storeToRefs()` |
| 2 | 按 Vite 改 webpack 仓 | 改 `vue.config.js` / 现有 webpack |
| 3 | `watch` 不立即执行 | `{ immediate: true }` |
| 4 | 路由复用不刷新 | `watch(route)` 或 `:key` |
| 5 | scoped 穿透失败 | `:deep()` |
| 6 | int64 精度 | 后端 string + 前端 coerce |
| 7 | 401 循环 | 登录页排除拦截 |
| 8 | 双层 `.data` | 对齐拦截器剥壳约定 |

---

## 规范落地优先级

> 文首「本次必读决策表」为 P0/P1 快速入口；下表为完整落地优先级。

| 优先级 | 条款类型 | 落地方式 |
|--------|----------|----------|
| P0 机器门禁 | ESLint/Stylelint/Prettier；`package.json` 已有 `lint`/`build`/`typecheck` | CI 或提交前跑对应 script；无 script 不得虚构通过 |
| P0 机器门禁 | 密钥/Token 不进仓 | 仓库 secret 扫描 / pre-commit（若有） |
| P1 Agent 必读 | 组件/目录约定、请求封装、Store、路由鉴权、webpack 构建约定 | 改前端前 Read 本文件相关节；IDE `standards-frontend` 仅督促加载 |
| P2 参考 | 布局惯例、非强制性能建议 | 可偏离，偏离时在 MR 说明 |

完整条款以本文件为准；短 Rules / AGENTS 门闩不含正文复述。
