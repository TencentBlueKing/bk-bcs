# 后端开发规范（Go + Gin REST）

> **启用前提：** `go.mod` 的 require 含 `github.com/gin-gonic/gin`；**无** tRPC / go-micro 等 RPC 框架主导信号时选用本预设。  
> 若同时存在 `trpc.group/trpc-go` 且仓库以 tRPC 为主，应改用 [`backend-trpc-go.md`](backend-trpc-go.md)。  
> **目录以探测到的后端根为准，禁止写死** `apps/bkms-server`、`internal/server` 等业务路径。  
> 下文分层与目录树为**常见示例**，非强制目录名；以仓库既有 layout 为准。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应章节；禁止默认全文灌入。完整条款见正文；落地优先级见文末「规范落地优先级」。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | Handler 直连 DB；手改 swagger 生成物；写死业务目录；向客户端泄露堆栈 | 二、四、七 |
| 必须 | Handler→Service→Store；`ShouldBind*`+校验；统一错误信封；全链路 `context` | 三、四、五 |
| 验证 | 跑仓库已有 `make lint/test/build` 或 `go test`；改 API 则重生 swagger | 七、[`api-swagger`](api-swagger.md) |

---

## 一、技术栈要求

| 技术 | 版本要求 | 用途 |
|------|---------|------|
| Go | 以 `go.mod` 为准 | 编程语言 |
| Gin | 以 `go.mod` 为准 | HTTP 路由与中间件 |
| 配置 | viper / env / 项目自研——以仓库为准 | 运行时配置 |
| 日志 | zap / logrus / slog——以仓库为准 | 结构化日志 |
| ORM / DB | gorm / sqlx / ent——以仓库为准 | 数据访问 |
| 测试（可选） | `testing` + testify；部分仓库用 Ginkgo/Gomega | 单元 / 集成测试 |
| Lint（可选） | golangci-lint | 静态检查 |

---

## 二、项目结构

> **禁止写死业务目录。** 以下为 Gin REST 单体或模块化单体常见布局**示例**。

```
{backend-root}/
├── cmd/
│   └── server/
│       └── main.go              # 入口：加载配置、组装依赖、启动 HTTP
├── internal/                      # 或 pkg/、app/——以仓库为准
│   └── {feature}/                 # 按业务能力分包（非必须同名）
│       ├── router.go              # 路由注册（可选独立文件）
│       ├── handler.go             # HTTP Handler：绑参、校验、调 Service
│       ├── serializer.go          # 请求/响应 DTO（可选）
│       ├── service.go             # 业务逻辑
│       └── store.go               # 持久化 / 外部依赖访问
├── pkg/                           # 可跨模块复用的库（若有）
├── configs/                       # 配置文件（若有）
├── docs/                          # OpenAPI 生成物（若用 swag，见 api-swagger）
├── go.mod
├── go.sum
├── Makefile                       # 若有：lint / test / build / apidocs
└── README.md
```

| 要点 | 说明 |
|------|------|
| 后端根 | 可能是仓库根、`server/`、`backend/` 等——以探测结果为准 |
| 分层 | Handler → Service → Store/Repository；**禁止** Handler 直连 SQL |
| 路由 | 集中在 `main`、独立 `router` 包或各 feature 的 `Register(r *gin.RouterGroup)`——沿用仓库 |
| 生成物 | `docs/swagger.*` 等若存在则**禁止手改**，只重新生成 |

---

## 三、编码规范

### 3.1 Handler 层

```go
// internal/user/handler.go — 示例，路径以仓库为准
func (h *Handler) GetUser(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest, gin.H{"message": "id required"})
        return
    }
    user, err := h.svc.GetByID(c.Request.Context(), id)
    if err != nil {
        // 错误映射以项目 middleware / errors 包为准
        c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": user})
}
```

| 规则 | 说明 |
|------|------|
| Context | 使用 `c.Request.Context()` 向下传递，支持超时与取消 |
| 绑参 | 见下表 `ShouldBind*`；**禁止**静默忽略 bind 错误 |
| 响应 | 成功/失败均走统一 envelope（字段名以项目为准） |
| 职责 | Handler 不写 SQL、不跨 feature 直接调 Store |

| 绑参方式 | 用途 | 失败时 |
|----------|------|--------|
| `ShouldBindJSON` | JSON body | `400` + 可读校验信息（不回显敏感原文） |
| `ShouldBindQuery` / `ShouldBindUri` | query / path | 同上 |
| `binding` tag | `required`、`min`、`max`、`oneof`、邮箱等 | 与 gin validator 一致；复杂规则下沉 Service |

结构体示例（字段以仓库为准）：

```go
type CreateUserReq struct {
    Name  string `json:"name" binding:"required,max=64"`
    Email string `json:"email" binding:"required,email"`
}
```

### 3.2 Service 与 Store

| 层 | 职责 | 禁止 |
|----|------|------|
| Service | 业务规则、事务边界、编排多个 Store | 依赖 `*gin.Context` |
| Store | CRUD、缓存、外部 HTTP/RPC 客户端 | 返回 HTTP 状态码 |

依赖通过构造函数注入；接口便于 mock（测试以仓库既有风格为准）。

### 3.3 路由与中间件

| 规则 | 说明 |
|------|------|
| 版本前缀 | 常见 `/api/v1/`——以仓库为准 |
| RESTful | 资源名词复数；路径参数语义清晰 |
| 中间件 | 鉴权、Recovery、RequestID、CORS 在路由组统一挂载 |
| 敏感接口 | 必须认证 + 授权——详见 [`security-bk-redlines.md`](security-bk-redlines.md) |

### 3.4 命名约定

| 元素 | 规则 | 示例 |
|------|------|------|
| 包名 | 小写单词 | `user`, `order` |
| Handler 方法 | 动词或 REST 语义 | `ListUsers`, `CreateUser` |
| JSON 字段 | snake_case（与前端/OpenAPI 对齐） | `user_id`, `created_at` |
| 错误变量 | `Err` 前缀或项目 errors 包 | `ErrNotFound` |

---

## 四、错误处理与日志

| 规则 | 说明 |
|------|------|
| 统一信封 | 错误 JSON 形状全项目一致（如 `code`/`message`/`request_id`）；与 [`api-swagger.md`](api-swagger.md) `@Failure` model 对齐 |
| 状态码映射 | 领域错误 → 4xx/5xx 的映射集中在 middleware 或 errors 包；Handler 不散落魔法数字 |
| `c.Error`（可选） | 若仓库采用 Gin 错误中间件：Handler `c.Error(err)` + `return`，由中间件写响应；新旧风格勿混用 |
| 包装错误 | `fmt.Errorf("get user: %w", err)`；日志打完整链，响应只给安全文案 |
| 用户消息 | **禁止**把内部 err/`%+v` 堆栈直接给客户端 |
| 结构化日志 | 带 request_id、path、latency；**禁止**记录密码/Token |
| Panic | 依赖 Recovery 中间件；业务路径禁止 panic |

---

## 五、配置、超时与安全

| 规则 | 说明 |
|------|------|
| 密钥 | DSN、API Key 经环境变量或 Secret 注入，**禁止**硬编码进仓 |
| 启动校验 | 必填配置缺失则 Fatal / 拒绝启动 |
| HTTP 超时 | `ReadHeader`/`Read`/`Write`/`Idle` 超时可配置；禁止无限等客户端 |
| 请求上下文 | 下游调用使用带 deadline 的 `context`（继承 `c.Request.Context()` 或显式 timeout） |
| 优雅退出 | `signal` → `Shutdown`/`Close` 排空连接后再关 DB；禁止 `os.Exit` 打断进行中请求 |
| 输入校验 | 所有外部输入在服务端强校验——见 [`security-bk-redlines.md`](security-bk-redlines.md) |
| SQL | 参数化查询；禁止字符串拼接 SQL |
| 改接口 | 注解变更后必须按 [`api-swagger.md`](api-swagger.md) 重生文档 |

---

## 六、测试（可选）

| 类型 | 工具 | 说明 |
|------|------|------|
| 单元测试 | `go test`, testify | Service / 纯函数优先 |
| Handler 测试 | `httptest` + Gin `ServeHTTP` | 覆盖绑参与状态码 |
| BDD（若有） | Ginkgo + Gomega | 沿用仓库 `go:generate` / suite 布局 |
| 集成测试 | build tag `integration` | 不阻塞默认 `go test ./...` |

**禁止**虚构 `proto/`、`stub/` 或 tRPC 测试脚手架。无测试目录时不得声称测试已通过。

---

## 七、构建与质量门禁

在**探测到的后端根**读取 `Makefile`、`scripts/`、`go.mod`：

| 目标 / 脚本 | 存在则 | 不存在则 |
|-------------|--------|----------|
| `make lint` / `golangci-lint run` | 提交前运行 | 不得虚构 lint 通过 |
| `make test` / `go test ./...` | 提交前运行 | 不得声称测试已通过 |
| `make build` / `go build ./...` | 交付前运行 | 至少保证主入口可编译 |
| `make apidocs` / swag | 改 API 注释后重新生成 | 见 [`api-swagger.md`](api-swagger.md) |

---

## 八、常见陷阱

| # | 陷阱 | 解决方案 |
|---|------|---------|
| 1 | Handler 直连 DB | 下沉到 Store |
| 2 | 忽略 `Context` 取消 | 全链路传递 `context.Context` |
| 3 | 响应格式前后端不一致 | 对齐 OpenAPI / 前端拦截器约定 |
| 4 | int64 JSON 精度 | 大整数序列化为 string（与前端约定） |
| 5 | 手改 swagger 生成物 | 只跑 `make apidocs` / swag |
| 6 | 内网接口无鉴权 | 按安全红线补认证与授权 |
| 7 | bind 失败仍进业务 | 检查 `ShouldBind*` 返回值 |
| 8 | 无超时打满连接 | 配 HTTP/上下文 deadline + 优雅退出 |

---

## 规范落地优先级

> 文首「本次必读决策表」为 P0/P1 快速入口；下表为完整落地优先级。

| 优先级 | 条款类型 | 落地方式 |
|--------|----------|----------|
| P0 机器门禁 | `make lint` / golangci-lint；`make test` / `go test`；`make build` | CI 或提交前跑仓库已有目标；无目标不得虚构通过 |
| P0 机器门禁 | 密钥不进仓；敏感接口鉴权；改 API 后 swagger 可生成 | secret 扫描 + 安全红线 + [`api-swagger.md`](api-swagger.md) |
| P1 Agent 必读 | 分层、`ShouldBind*`、错误信封、超时/优雅退出、Makefile | 改后端前按节 Read；IDE `standards-backend` 仅督促加载 |
| P2 参考 | Ginkgo 风格、目录命名示例、性能微优化 | 可偏离，偏离时在 MR 说明 |

完整条款以本文件为准；短 Rules / AGENTS 门闩不含正文复述。
