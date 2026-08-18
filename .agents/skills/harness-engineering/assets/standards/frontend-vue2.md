# 前端开发规范（Vue 2）

> 适用于 **Vue 2.x + webpack**（含 Vue CLI / `@vue/cli-service` / `@blueking/cli-service-webpack`）的中后台前端。  
> 对齐 [Vue 2 Style Guide](https://v2.vuejs.org/v2/style-guide/)（Priority A/B）、Vuex、Vue Router 3 社区惯例。  
> **目录与命令以仓库实际前端根（含 `package.json` 的目录）为准**，禁止写死 `web/src` 等业务路径。  
> **不要**套用 Vue 3 Composition API / Pinia / Vite / `script setup` 约定，除非仓库已显式迁移。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应章节；禁止默认全文灌入。完整条款见正文；落地优先级见文末「规范落地优先级」。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 写死目录；误用 Vue 3 / Pinia / Vite / `script setup` | 二、十 |
| 必须 | Options API + Vuex；请求走统一封装；`v-for` 稳定 `:key` | 三、四、五 |
| 验证 | 跑 `package.json` 已有 lint/build/typecheck（无则不得虚构通过） | 九、文末 P0 |

---

## 一、技术栈要求

| 技术 | 版本要求 | 用途 |
|------|---------|------|
| Vue | 2.x（常见 2.6 / 2.7；以 package.json 为准） | 前端框架 |
| 构建 | webpack / Vue CLI 系（非 Vite） | 本地开发与打包 |
| vue-router | 3.x（以项目为准） | 路由 |
| 状态 | 常见 Vuex 3.x；以项目现有为准 | 全局状态 |
| HTTP | axios 等；以项目封装为准 | 网络请求 |
| UI | 以项目依赖为准（如 Element UI、bk-magic-vue） | 组件库 |
| 语言 | JS 或 TS（以仓库为准） | — |

版本号以该前端根目录 `package.json` 锁定为准。

---

## 二、项目结构

> **禁止写死业务目录。** 以下为 Vue CLI / webpack 中后台常见布局**示例**，实际以探测到的前端根为准。

```
{frontend-root}/                 # 含 package.json 的目录
├── public/                      # 静态资源（若使用 Vue CLI）
├── src/
│   ├── main.js|.ts              # 入口
│   ├── App.vue
│   ├── router/                  # 或 router.js
│   ├── store/                   # Vuex（modules 按域拆分）
│   ├── views/|pages/            # 路由页面
│   ├── components/              # 可复用组件
│   ├── api/|http/|services/     # 请求封装（命名以仓库为准）
│   ├── utils/|helpers/
│   └── assets/
├── vue.config.js                # 或项目等价 webpack 配置
└── package.json
```

- 前端根可能是 `webfe/package_vue`、`src/pages`、`frontend` 等，**不一定是** `web/src`。
- **不要**假设存在 `services/generated/` 或强制「所有请求必须经某一固定文件」——以仓库真实入口为准。
- 本地命令：读该目录 `package.json` scripts（常见 `serve` / `dev` / `build` / `lint`），禁止写死他仓命令。

---

## 三、编码规范（Options API）

### 3.1 组件选项顺序（Style Guide Priority B）

单文件组件（SFC）推荐按下列顺序声明（与社区 Style Guide 一致）：

1. `name`  
2. `components` / `directives` / `filters`  
3. `extends` / `mixins`  
4. `props` / `propsData`  
5. `data` / `computed`  
6. `watch`  
7. 生命周期钩子（`beforeCreate` → `created` → … → `destroyed`）  
8. `methods`  

```vue
<template>
  <div class="user-list">
    <!-- ... -->
  </div>
</template>

<script>
export default {
  name: 'UserList',
  components: { /* ... */ },
  props: {
    id: { type: String, required: true },
    title: { type: String, default: '' },
  },
  data() {
    return {
      loading: false,
      list: [],
    };
  },
  computed: {
    isEmpty() {
      return this.list.length === 0;
    },
  },
  created() {
    this.fetchData();
  },
  methods: {
    async fetchData() {
      this.loading = true;
      try {
        this.list = await this.$api /* 或项目封装 */.list(this.id);
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>

<style scoped>
.user-list { /* ... */ }
</style>
```

**要点：**

- 以 **Options API** 为主；Vue 2.7 若已使用 Composition API，仅沿用仓库既有写法，不主动全面切换。
- `props` 必须声明 `type`；必需项加 `required: true`，可选给 `default`。
- 避免在 `created`/`mounted` 以外随意操作 DOM；优先数据驱动。
- 过滤器 `filters`：新代码优先用 methods/computed；存量可保留。

### 3.2 命名约定（Style Guide）

| 类型 | 规则 | 示例 |
|------|------|------|
| 组件文件 / `name` | PascalCase 或多词名称（避免单个词根组件名） | `UserList.vue` / `UserList` |
| 组件在模板中 | PascalCase 或 kebab-case，项目内统一 | `<UserList />` 或 `<user-list />` |
| props | camelCase 定义，模板可用 kebab-case | `userId` / `user-id` |
| 事件名 | kebab-case | `update:title`、`item-click` |
| Vuex mutation | UPPER_SNAKE_CASE | `SET_USER_INFO` |
| Vuex action | camelCase | `fetchUserInfo` |
| CSS class | kebab-case | `.page-header` |
| 常量 | UPPER_SNAKE_CASE | `MAX_RETRY_COUNT` |

### 3.3 模板与指令

| 规则 | 说明 |
|------|------|
| `v-for` 必须 `:key` | 优先稳定业务 id，避免仅用 index（列表会重排时） |
| 避免 `v-if` 与 `v-for` 同节点 | 拆开或外包一层（Vue 2 中 `v-for` 优先级更高，易踩坑） |
| 组件通信 | 父→子 props；子→父 `$emit`；跨层 Vuex / provide-inject（慎用） |
| 样式 | 优先 `scoped`；穿透使用仓库已有方案（`::v-deep` / `/deep/`），项目内统一 |

---

## 四、状态管理（Vuex）

### 4.1 模块化

- 按业务域拆 `store/modules/*`，开启 `namespaced: true`（与仓库一致即可）。
- **mutation 同步、纯函数式改 state**；异步只放 **action**。
- 组件内：读 state 用 `mapState` / `mapGetters`；改状态用 `mapActions` / `dispatch`，避免组件直接 `commit` 复杂逻辑（简单赋值可接受，以仓库习惯为准）。

```javascript
// 示例：namespaced module（路径以仓库为准）
export default {
  namespaced: true,
  state: () => ({ userInfo: null, loading: false }),
  getters: {
    username: (s) => (s.userInfo && s.userInfo.name) || '',
  },
  mutations: {
    SET_USER_INFO(state, payload) {
      state.userInfo = payload;
    },
    SET_LOADING(state, v) {
      state.loading = v;
    },
  },
  actions: {
    async fetchUserInfo({ commit, state }) {
      if (state.userInfo) return state.userInfo;
      commit('SET_LOADING', true);
      try {
        const data = await api.getUserInfo(); // api 为项目封装
        commit('SET_USER_INFO', data);
        return data;
      } finally {
        commit('SET_LOADING', false);
      }
    },
  },
};
```

### 4.2 使用规范

| 规则 | 说明 |
|------|------|
| 单一数据源 | 跨页共享状态进 Vuex；仅局部 UI 状态留在组件 `data` |
| 不在 mutation 里发请求 | 请求与副作用放 action |
| 避免巨型 root state | 新域优先新 module |

---

## 五、网络请求规范

### 5.1 统一封装

- 所有业务请求走项目统一 HTTP 封装（axios 实例 + 拦截器）；**禁止**在页面里裸 `axios.create` 散落多套 baseURL。
- 拦截器职责：鉴权头、统一错误提示、401 跳转登录（登录页需排除循环）、响应剥壳（按项目协议）。
- 封装路径以仓库为准（`src/api/http.js`、`src/http/index.js` 等均可）。

### 5.2 核心规则

| 规则 | 说明 |
|------|------|
| 统一入口 | 业务只调封装后的 `get/post` 或 API 模块方法 |
| 错误提示 | Toast/Message 统一，避免每页复制粘贴 |
| 超时与取消 | 长请求按项目现有策略；勿静默吞错 |
| int64 | 若后端以 string 传大整数，前端勿强转 Number 丢精度 |

---

## 六、路由（vue-router 3）

| 规则 | 说明 |
|------|------|
| 懒加载 | 路由 `component: () => import(...)`，减小首包 |
| 守卫 | 鉴权、标题等放 `beforeEach`；与仓库现有守卫合并，勿重复造轮子 |
| 命名路由 | 复杂跳转优先 `name` + `params`，避免魔法路径字符串散落 |
| 参数变化 | 同一组件复用时 `watch: { '$route' }` 或 `beforeRouteUpdate` 拉数 |

---

## 七、构建与工程（webpack / Vue CLI）

| 规则 | 说明 |
|------|------|
| 配置入口 | `vue.config.js` / 项目 webpack 配置；改代理、alias、publicPath 走现有文件 |
| 环境变量 | `VUE_APP_*`（Vue CLI）或项目既有约定；勿把密钥写入前端仓 |
| 依赖升级 | Vue 2 生态注意 peer 版本；勿在未评估时引入仅 Vue 3 的库 |
| 与 Vite 预设隔离 | 无 `vite.config.*`；禁止按 Vite 插件/约定改本仓库 |

---

## 八、UI 组件使用规范

| 原则 | 说明 |
|------|------|
| 组件优先 | 优先项目 UI 库；避免页面手写一套表格/对话框 |
| 样式统一 | 使用库的主题/变量；减少页面级魔法色值 |
| Icon | 跟项目现有图标方案；禁止随意 CDN 外链图标（安全与稳定性） |

列表页 / 详情页 / 表单页布局模式与中后台惯例一致：搜索区 + 表格 + 分页；面包屑 + 信息区；表单 + 提交区。

---

## 九、质量保证

### 9.1 提交前验证（按仓库实际脚本）

在前端根目录根据 `package.json` scripts 执行：

- 有 `lint` / `build`：运行对应脚本。
- 有 `test` / `test:unit`：运行对应脚本。
- **无**测试脚本：**不得**要求 Jest/Vitest，也不得声称「测试已通过」。
- 仅当存在类型检查脚本时才跑 TS 检查。

### 9.2 代码审查清单

- [ ] 组件 `name` 与文件名合理；props 有类型与默认值
- [ ] `v-for` 有稳定 `:key`；无 `v-if`+`v-for` 同节点
- [ ] 请求走统一封装；无密钥进仓
- [ ] 跨页状态进 Vuex module；mutation 无异步
- [ ] 路由懒加载；鉴权守卫与现网一致
- [ ] 未误用 Vue 3 / Vite / Pinia API
- [ ] `lint`/`build`（若有）通过

---

## 十、常见陷阱与避坑

| # | 陷阱 | 解决方案 |
|---|------|---------|
| 1 | `v-if` + `v-for` 同元素 | 拆节点或计算属性过滤列表 |
| 2 | 直接改 props | 用 data/computed 或 `.sync` / `update:` 事件 |
| 3 | Vuex 解构丢失响应 | 用 `mapState`/`mapGetters` 或在 computed 里读 `$store` |
| 4 | 路由同组件不刷新 | `watch $route` / `beforeRouteUpdate` |
| 5 | 内存泄漏 | `destroyed` 里清理 timer、事件、第三方实例 |
| 6 | 401 死循环 | 登录页排除拦截器 |
| 7 | 响应多剥一层 `.data` | 与拦截器约定对齐 |
| 8 | 把 Vue 3 写法拷进 Vue 2 | 以本文件与仓库现状为准 |

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
