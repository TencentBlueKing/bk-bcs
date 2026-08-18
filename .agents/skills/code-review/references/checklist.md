# 代码评审检查清单

完整的代码评审规则索引和详细说明。

## 严重级别

本 skill 以 CRITICAL/HIGH/MEDIUM/LOW 四级严重级别为唯一权威分级体系（配套 `scoring-standard.md` 的评分与一票否决项），不引入并行分级体系：

| 级别 | 处理 |
|------|------|
| CRITICAL | 阻止合入，必须修复 |
| HIGH | 阻止合入，必须修复 |
| MEDIUM | 讨论后决定是否修改 |
| LOW | 可忽略，作者自行决定 |

## 设计与架构类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **DESIGN-001** | 模块设计 | 代码是否属于正确的模块/库 | 🔴 高 |
| **DESIGN-002** | 系统集成 | 是否能与系统其他部分良好集成 | 🔴 高 |
| **DESIGN-003** | 时机选择 | 现在是否是添加此功能的合适时机 | 🟡 中 |
| **DESIGN-004** | 过度工程 | 是否存在不必要的复杂设计 | 🔴 高 |

### 检查要点
- CL 中各种代码的交互是否有意义
- 此变更是属于代码库还是应该属于独立库
- 它是否能与系统其他部分很好地集成
- 现在是添加此功能的好时机吗

## 功能与正确性类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **FUNC-001** | 行为正确 | 代码行为是否符合预期 | 🔴 高 |
| **FUNC-002** | 边缘情况 | 是否处理边缘情况和异常 | 🔴 高 |
| **FUNC-003** | 并发安全 | 并发操作是否安全 | 🔴 高 |
| **FUNC-004** | 用户友好 | 对用户是否友好 | 🟡 中 |

### 检查要点
- 代码的行为是否符合作者的预期
- 代码的行为方式是否对用户友好
- 考虑边缘情况、并发问题
- 对于 UI 变更，验证变化的合理性

## 复杂度与可维护性类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **COMP-001** | 代码简化 | 代码是否可以被简化 | 🟡 中 |
| **COMP-002** | 可理解性 | 其他开发人员能否轻松理解 | 🔴 高 |
| **COMP-003** | 未来功能 | 是否实现了不确定的未来功能 | 🟡 中 |

### 检查要点
- 代码是否可以被简化
- 其他开发人员是否能够轻松理解和使用该代码
- 避免过度工程（over-engineering）
- 警惕开发者实现未来可能需要但现在不确定的功能

## 测试类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **TEST-001** | 测试存在 | 是否包含自动化测试 | 🔴 高 |
| **TEST-002** | 测试有效 | 测试是否能在代码破坏时失败 | 🔴 高 |
| **TEST-003** | 测试设计 | 测试是否设计良好 | 🟡 中 |
| **TEST-004** | 误报风险 | 测试是否会产生误报 | 🟡 中 |

### 检查要点
- 代码中是否包含正确且设计良好的自动化测试
- 测试是否真的会在代码被破坏时失败
- 测试是否会产生误报
- 每个测试是否有简单而有用的断言

## 命名与注释类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **NAME-001** | 命名清晰 | 变量/类/方法名是否清晰 | 🔴 高 |
| **NAME-002** | 命名长度 | 名称长度是否合适 | 🟡 中 |
| **COMMENT-001** | 注释有用 | 注释是否清晰有用 | 🟡 中 |
| **COMMENT-002** | 解释为什么 | 注释是否解释"为什么"而非"做什么" | 🟡 中 |

### 检查要点
- 变量、类、方法等是否选择了清晰的名称
- 名称应该足够长以完全传达内容或作用
- 但不会太长难以阅读
- 注释应该解释"为什么"而不是"做什么"

## 代码规范类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **STYLE-001** | 规范遵守 | 是否遵循公司代码规范 | 🟡 中 |
| **STYLE-002** | 风格一致 | 代码整体风格是否一致 | 🟡 中 |
| **STYLE-003** | 文件大小 | 文件是否超出项目约定行数上限（通常 200-400 行，最多 800 行） | 🟡 中 |
| **STYLE-004** | Emoji 策略 | 代码中是否违反项目 emoji 约定（许多项目禁止在代码中使用 emoji） | ⚪ 建议 |
| **STYLE-005** | 不可变性 | 是否遵循项目不可变性要求（用展开运算符而非可变操作） | 🟡 中 |
| **DOC-001** | 文档更新 | 相关文档是否同步更新 | ⚪ 建议 |

> **项目专属通用项说明**：STYLE-003~005 源自项目专属准则中可通用化的检查项（文件大小限制、Emoji 策略、不可变性要求）。评审时以项目 `AGENTS.md` 或项目规则中的具体阈值为准；拿不准时与代码库其余部分保持一致。数据库策略（RLS、迁移模式）、错误处理模式（自定义错误类、错误边界）、状态管理约定（Zustand、Redux、Context）等项目专属约定，按项目实际约定核查。

### JavaScript/TypeScript 必须规则

| 规则 | 说明 | 示例 |
|------|------|------|
| 使用 `const` 定义引用 | 避免使用 `var` | `const name = 'John';` |
| 重新赋值时使用 `let` | 代替 `var` | `let count = 0;` |
| 使用字面量语法 | 创建对象和数组 | `const obj = {};` |
| 使用字符串模板 | 代替字符串拼接 | `` `Hello, ${name}` `` |
| 永远不要使用 `eval()` | 安全风险 | - |
| 使用 `===` 和 `!==` | 而不是 `==` 和 `!=` | `if (a === b)` |
| 多行代码块使用大括号 | 包裹代码块 | `if (x) { ... }` |
| 要加分号 | 语句结束 | `const x = 1;` |

## Git 提交规范类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **GIT-001** | Commit 规范 | Commit message 是否符合 Conventional Commits 规范 | 🟡 中 |
| **GIT-002** | 变更范围 | 单次变更是否过大（建议 < 400 行代码变更） | 🟡 中 |
| **GIT-003** | 冲突风险 | 是否存在潜在的合并冲突或破坏性变更 | 🔴 高 |
| **GIT-004** | 敏感信息 | 是否包含敏感文件或信息（.env、密钥、凭证等） | 🔴 高 |
| **GIT-005** | 调试代码 | 是否残留调试代码（console.log、debugger 等） | 🟡 中 |
| **GIT-006** | 单一职责 | Commit 是否遵循单一职责原则 | 🟡 中 |

## 安全审计类 (OWASP Top 10 前端相关)

| 规则 ID | 类别 | 描述 | 优先级 | OWASP |
|---------|------|------|--------|-------|
| **SEC-001** | XSS 防护 | 是否存在跨站脚本攻击风险（innerHTML、v-html、dangerouslySetInnerHTML） | 🔴 高 | A03 |
| **SEC-002** | 输入校验 | 用户输入是否经过校验和转义处理 | 🔴 高 | A03 |
| **SEC-003** | URL 注入 | 动态 URL 是否经过校验（javascript:、data: 协议） | 🔴 高 | A03 |
| **SEC-004** | 敏感数据存储 | 敏感信息是否存储在 localStorage/sessionStorage | 🔴 高 | A02 |
| **SEC-005** | Token 安全 | JWT/Token 是否安全存储和传输 | 🔴 高 | A07 |
| **SEC-006** | 前端鉴权 | 是否仅依赖前端进行权限控制 | 🔴 高 | A01 |
| **SEC-007** | CSRF 防护 | 敏感操作是否有 CSRF Token 保护 | 🔴 高 | A01 |
| **SEC-008** | 依赖安全 | 第三方依赖是否存在已知漏洞 | 🔴 高 | A06 |
| **SEC-009** | 资源完整性 | 外部 CDN 资源是否使用 SRI 校验 | 🟡 中 | A08 |
| **SEC-010** | 敏感信息泄露 | 是否在前端暴露敏感配置或 API 密钥 | 🔴 高 | A02 |
| **SEC-011** | 点击劫持 | 是否考虑 X-Frame-Options 或 CSP frame-ancestors | 🟡 中 | A05 |
| **SEC-012** | 开放重定向 | URL 跳转是否校验目标地址白名单 | 🔴 高 | A01 |

### 检查要点

**A01 - 权限控制失效 (Broken Access Control)**
- 前端路由守卫不能作为唯一的权限控制手段
- 敏感数据和操作必须有后端校验
- URL 跳转需要校验目标地址

**A02 - 加密失败 (Cryptographic Failures)**
- 敏感数据（密码、Token）不应存储在 localStorage
- 避免在前端硬编码 API 密钥、加密密钥
- 确保使用 HTTPS 传输敏感信息

**A03 - 注入 (Injection)**
- 避免使用 `innerHTML`、`v-html`、`dangerouslySetInnerHTML`
- 用户输入必须转义后再渲染
- 动态 URL 必须校验协议（禁止 javascript:、data:）
- 避免使用 `eval()`、`new Function()`、`setTimeout(string)`

```javascript
// ❌ 危险：XSS 风险
element.innerHTML = userInput;
<div v-html="userContent"></div>
<div dangerouslySetInnerHTML={{ __html: userContent }} />

// ✅ 安全：使用 textContent 或转义
element.textContent = userInput;
<div>{{ userContent }}</div>  // Vue 自动转义
<div>{userContent}</div>      // React 自动转义
```

**A05 - 安全配置错误 (Security Misconfiguration)**
- 生产环境应配置 CSP (Content-Security-Policy)
- 禁用不必要的 CORS 允许来源
- 移除开发环境的调试信息

**A06 - 易受攻击的组件 (Vulnerable Components)**
- 定期运行 `npm audit` 检查依赖漏洞
- 及时更新有安全漏洞的依赖
- 谨慎引入来源不明的第三方库

**A07 - 认证失败 (Authentication Failures)**
- Token 应存储在 httpOnly Cookie 中（而非 localStorage）
- 实现 Token 刷新机制
- 敏感操作需要二次验证

**A08 - 软件完整性失败 (Software Integrity Failures)**
- 外部 CDN 资源使用 SRI (Subresource Integrity)

```html
<!-- ✅ 使用 SRI 校验 -->
<script src="https://cdn.example.com/lib.js"
        integrity="sha384-xxxx"
        crossorigin="anonymous"></script>
```

## 性能检查类

| 规则 ID | 类别 | 描述 | 优先级 |
|---------|------|------|--------|
| **PERF-001** | 内存泄漏 | 事件监听器、定时器、订阅是否正确清理 | 🔴 高 |
| **PERF-002** | 大列表渲染 | 长列表是否使用虚拟滚动 | 🔴 高 |
| **PERF-003** | 不必要渲染 | 组件是否存在不必要的重复渲染 | 🟡 中 |
| **PERF-004** | 图片优化 | 图片是否懒加载、是否使用合适的格式和尺寸 | 🟡 中 |
| **PERF-005** | 代码分割 | 是否按路由/功能进行代码分割 | 🟡 中 |
| **PERF-006** | 网络请求 | 是否有重复请求、是否使用缓存和防抖 | 🟡 中 |
| **PERF-007** | 计算缓存 | 复杂计算是否使用 useMemo/computed 缓存 | 🟡 中 |
| **PERF-008** | Bundle 体积 | 是否引入了过大的依赖包 | 🟡 中 |
| **PERF-009** | DOM 操作 | 是否存在频繁的 DOM 操作导致重排重绘 | 🟡 中 |
| **PERF-010** | 异步加载 | 非关键资源是否异步/延迟加载 | 🟡 中 |

### 检查要点

**PERF-001 内存泄漏**
- 组件卸载时清理事件监听器
- 清理定时器（setTimeout、setInterval）
- 取消未完成的网络请求
- 取消订阅（RxJS、EventEmitter）

```javascript
// ❌ 内存泄漏：未清理
useEffect(() => {
  window.addEventListener('resize', handleResize);
  const timer = setInterval(poll, 1000);
}, []);

// ✅ 正确清理
useEffect(() => {
  window.addEventListener('resize', handleResize);
  const timer = setInterval(poll, 1000);
  return () => {
    window.removeEventListener('resize', handleResize);
    clearInterval(timer);
  };
}, []);
```

```vue
<!-- Vue 3 示例 -->
<script setup>
import { onMounted, onUnmounted } from 'vue';

let timer;
onMounted(() => {
  timer = setInterval(poll, 1000);
});
onUnmounted(() => {
  clearInterval(timer);
});
</script>
```

**PERF-002 大列表渲染**
- 超过 100 条数据的列表考虑虚拟滚动
- 推荐使用 `vue-virtual-scroller`、`react-window`、`@tanstack/virtual`

**PERF-003 不必要渲染**
- React：使用 `React.memo`、`useMemo`、`useCallback`
- Vue：避免在 template 中调用方法，使用 computed

```javascript
// ❌ 每次渲染都重新计算
function Component({ items }) {
  const sorted = items.sort((a, b) => a.name.localeCompare(b.name));
  return <List items={sorted} />;
}

// ✅ 缓存计算结果
function Component({ items }) {
  const sorted = useMemo(
    () => [...items].sort((a, b) => a.name.localeCompare(b.name)),
    [items]
  );
  return <List items={sorted} />;
}
```

**PERF-004 图片优化**
- 使用 `loading="lazy"` 懒加载图片
- 使用 WebP/AVIF 格式
- 提供合适尺寸的图片（响应式图片）
- 使用 `<picture>` 或 srcset 适配不同设备

```html
<!-- ✅ 图片懒加载 + 响应式 -->
<img 
  src="image.webp" 
  loading="lazy"
  srcset="image-320.webp 320w, image-640.webp 640w"
  sizes="(max-width: 600px) 320px, 640px"
  alt="描述"
/>
```

**PERF-005 代码分割**
- 路由级别代码分割
- 大型第三方库按需加载

```javascript
// ✅ 路由懒加载
const UserProfile = () => import('./views/UserProfile.vue');
const Dashboard = React.lazy(() => import('./pages/Dashboard'));
```

**PERF-006 网络请求优化**
- 搜索/输入使用防抖（debounce）
- 滚动/拖拽使用节流（throttle）
- 合并多个请求
- 使用请求缓存（SWR、React Query、VueQuery）

```javascript
// ✅ 搜索防抖
const debouncedSearch = useDebouncedCallback(
  (value) => fetchResults(value),
  300
);
```

**PERF-008 Bundle 体积**
- 避免引入整个 lodash，使用 `lodash-es` 或单独导入
- 使用 `date-fns` 代替 `moment.js`
- 分析 bundle 体积：`webpack-bundle-analyzer`、`vite-bundle-visualizer`

```javascript
// ❌ 引入整个 lodash (约 70KB)
import _ from 'lodash';
_.debounce(fn, 300);

// ✅ 按需引入 (约 2KB)
import debounce from 'lodash-es/debounce';
debounce(fn, 300);
```

**PERF-009 DOM 操作**
- 批量 DOM 更新使用 `requestAnimationFrame`
- 避免在循环中读写 DOM（触发强制同步布局）
- 使用 CSS transform 代替 top/left 动画

## Node.js 后端专项（HIGH）

评审 Node.js / 后端代码时，还需检查：

- **未校验的输入** —— 请求体/参数未经 schema 校验便使用
- **缺少限流** —— 公开端点没有节流
- **无界查询** —— 面向用户的端点使用 `SELECT *` 或没有 LIMIT 的查询
- **N+1 查询** —— 在循环中获取关联数据，而非使用 join/批处理
- **缺少超时** —— 外部 HTTP 调用没有配置超时
- **错误信息泄露** —— 将内部错误细节发送给客户端
- **缺少 CORS 配置** —— API 可被非预期来源访问

```typescript
// 反例：N+1 查询模式
const users = await db.query('SELECT * FROM users');
for (const user of users) {
  user.posts = await db.query('SELECT * FROM posts WHERE user_id = $1', [user.id]);
}

// 正例：使用 JOIN 或批处理的单次查询
const usersWithPosts = await db.query(`
  SELECT u.*, json_agg(p.*) as posts
  FROM users u
  LEFT JOIN posts p ON p.user_id = u.id
  GROUP BY u.id
`);
```

## Golang 专项

### 严重 —— 安全问题（CRITICAL/HIGH）
- **SQL 注入**：在 `database/sql` 查询中使用字符串拼接
- **命令注入**：在 `os/exec` 中使用未经验证的输入
- **路径遍历**：用户控制的文件路径未进行 `filepath.Clean` 及前缀检查
- **竞态条件**：共享状态未做同步处理
- **unsafe 包**：无充分理由使用
- **硬编码密钥**：源码中包含 API 密钥、密码
- **不安全的 TLS**：设置 `InsecureSkipVerify: true`

### 严重 —— 错误处理（CRITICAL/HIGH）
- **忽略错误**：使用 `_` 丢弃错误
- **缺少错误包装**：直接 `return err`，未使用 `fmt.Errorf("context: %w", err)`
- **对可恢复错误使用 panic**：应使用错误返回
- **缺少 errors.Is/As**：应使用 `errors.Is(err, target)`，而非 `err == target`

### 高危 —— 并发问题（HIGH）
- **协程泄漏**：无取消机制（应使用 `context.Context`）
- **无缓冲通道死锁**：发送操作没有接收方
- **缺少 sync.WaitGroup**：协程间缺乏协调
- **互斥锁误用**：未使用 `defer mu.Unlock()`

### 高危 —— 代码质量（HIGH）
- **函数过大**：超过 50 行
- **嵌套过深**：超过 4 层
- **不符合语言习惯**：使用 `if/else` 而非提前返回
- **包级变量**：可变的全局状态
- **接口污染**：定义未使用的抽象

### 中等 —— 性能问题（MEDIUM）
- **循环中字符串拼接**：应使用 `strings.Builder`
- **切片未预分配容量**：应使用 `make([]T, 0, cap)`
- **N+1 查询**：在循环中进行数据库查询
- **不必要的内存分配**：热点路径中创建对象

### 中等 —— 最佳实践（MEDIUM）
- **Context 应作为首参数**：`ctx context.Context` 应为第一个参数
- **表驱动测试**：测试应采用表驱动模式
- **错误消息格式**：小写字母，无标点
- **包命名规范**：简短、全小写、不含下划线
- **循环中延迟调用**：存在资源累积风险

## AI 生成代码评审

评审 AI 生成的变更时，优先关注：

1. 行为回归与边界情况处理
2. 安全假设与信任边界
3. 隐藏的耦合或意外的架构漂移
4. 不必要且推高模型成本的复杂度

成本意识检查：
- 标记那些在没有明确推理需求的情况下就升级到更高成本模型的工作流。
- 对确定性的重构，建议默认使用更低成本的模型档位。

---

> 更多详细内容请按需加载：
> - 评分标准：`./references/scoring-standard.md`
> - 撰写规范：`./references/writing-guidelines.md`
> - 置信度与误报控制：`./references/confidence-filtering.md`
