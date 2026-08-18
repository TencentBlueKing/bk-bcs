# 前后端接口规范（Gin + Swagger / OpenAPI）

> **启用前提：** `go.mod` require 含 `github.com/gin-gonic/gin`，且含 `github.com/swaggo/swag` **或** `github.com/swaggo/gin-swagger`（二者有其一即可）。  
> 接口契约以 **swag 注解 + 生成的 OpenAPI** 为单一事实源；**禁止**写死 `web/src`、`services/generated/` 等前端路径。  
> 安全红线（输入校验、鉴权、敏感数据）横切适用 → [`security-bk-redlines.md`](security-bk-redlines.md)。  
> 后端分层与 Handler 写法 → [`backend-gin.md`](backend-gin.md)。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应章节；禁止默认全文灌入。完整条款见正文；落地优先级见文末「规范落地优先级」。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 手改 swagger 生成物；注解与路由/鉴权不一致；口头约定字段 | 二、三、六 |
| 必须 | 改 API 后重生文档；敏感路由 `@Security`+中间件一致；破坏性变更升版本 | 三、五、六 |
| 验证 | `make apidocs`/`swag init`；若有契约漂移 CI 则跑（无则不得虚构） | 五、九、文末 P0 |

---

## 一、架构概述

| 层级 | 技术 | 职责 |
|------|------|------|
| HTTP 服务 | Gin | 路由、中间件、Handler |
| 契约定义 | swag 注解（Handler / model 注释） | 描述路径、参数、响应 schema |
| 文档生成 | `swag init` / `make apidocs` | 产出 OpenAPI JSON/YAML |
| 在线文档（可选） | gin-swagger | 开发/内网 Swagger UI |
| 前端消费 | 以项目为准（axios + 手写类型 / openapi-codegen 等） | **禁止**本文件强制 vitest 或某一 codegen |

### 1.1 请求链路（示例）

```
Client → Gin Router → Middleware（鉴权等）→ Handler → Service → Store
                ↓
         swag 注解 ──generate──→ docs/swagger.json（示例路径）
```

---

## 二、生成物与目录

> 生成目录以项目 `swag` 配置 / Makefile 为准；以下为**常见示例**。

```
{backend-root}/
├── internal/{feature}/handler.go    # @Router / @Summary 等注解
├── docs/                            # swag 默认输出（示例）
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
└── Makefile                         # apidocs / swagger 目标（若有）
```

| 规则 | 说明 |
|------|------|
| 禁止手改 | `docs/docs.go`、`swagger.json`、`swagger.yaml` 只通过生成更新 |
| 入库策略 | 以团队约定为准（多数团队提交生成物便于 CI 与前端拉取） |
| 版本策略 | **兼容**变更：additive 字段 + 文档说明；**破坏性**变更：升 URL `/api/v2` **或** 明确 bump `info.version` 并在 MR 标注 breaking（二者择一，全仓一致） |

---

## 三、注解与 Handler 规范

### 3.1 主入口注解（示例）

```go
// @title           API 标题
// @version         1.0
// @BasePath        /api/v1
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() { /* ... */ }
```

### 3.2 路由注解（示例）

```go
// ListUsers 获取用户列表
// @Summary      用户列表
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        page      query int false "页码"
// @Param        page_size query int false "每页数量"
// @Success      200 {object} ListUsersResponse
// @Failure      400 {object} ErrorResponse
// @Failure      401 {object} ErrorResponse
// @Security     Bearer
// @Router       /users [get]
func (h *Handler) ListUsers(c *gin.Context) { /* ... */ }
```

| 鉴权注解 | 要求 |
|----------|------|
| 全局定义 | 主入口 `@securityDefinitions.*` 与真实 Header/Cookie 名一致 |
| 需登录路由 | 必须写 `@Security`（名称与定义一致），且挂在同一 Auth 路由组 |
| 公开路由 | **不要**标 `@Security`；文档与中间件白名单一致 |
| 禁止 | 实现有鉴权但文档未标，或文档标了但中间件放行 |

### 3.3 命名与 URL 约定

| 元素 | 规则 | 示例 |
|------|------|------|
| 路径 | 名词复数、版本前缀 | `/api/v1/users` |
| 路径参数 | snake_case，与 JSON 一致 | `/users/{user_id}` |
| Tags | 按业务模块 | `user`, `order` |
| 响应模型 | 独立 struct，字段带 `json` tag | `ListUsersResponse` |
| 错误体 | 统一 `ErrorResponse`（code/message/details） | 与前端拦截器对齐 |

---

## 四、数据类型与 JSON 契约

| 类型 | JSON | 前端 TypeScript（示例） | 注意 |
|------|------|-------------------------|------|
| int32 | number | `number` | |
| int64 | **string**（推荐）或 number | `string` | 防 JS 精度丢失 |
| bool | boolean | `boolean` | |
| time | RFC3339 string | `string` | 统一 UTC 或文档声明时区 |
| 列表 | array | `T[]` | 空列表用 `[]` 非 null（团队约定） |
| 分页 | `items` + `total` 等 | 与现有前端 list 封装对齐 | 字段名变更需同步前端 |

**响应 envelope：** 若统一 `{ "data": ... }` 或 `{ "code", "message", "data" }`，所有 `@Success` model 与中间件剥壳规则一致；前端只消费剥壳后的形状。

---

## 五、生成命令

在**探测到的后端根**查找 `Makefile` 或 `scripts/`：

```bash
# 常见（以仓库为准）
make apidocs
# 或
swag init -g cmd/server/main.go -o docs
```

| 步骤 | 要求 |
|------|------|
| 改 API | 先改 Handler 注解与 model，再跑生成 |
| MR 自检 | diff 含 swagger 生成物（若团队要求提交） |
| 契约漂移 | 若 CI/`make` 有「先 `swag init` 再 `git diff --exit-code`」类检查：**必须跑**；**无则不得虚构「契约已校验」** |
| gin-swagger | 仅 dev/staging 暴露 UI；生产按安全策略关闭或加鉴权 |

---

## 六、前后端契约同步

| 场景 | 做法 |
|------|------|
| 新接口 | 后端合并前生成物更新；前端按 OpenAPI 或联调文档接入 |
| 字段变更 | 优先 additive；删除/重命名走版本或 deprecation 周期 |
| 前端类型 | 有 openapi-codegen 则重新生成；无则手写 types 并对齐 swagger |
| 联调 | 以 swagger.json + 示例 payload 为准，避免口头约定 |
| 鉴权 | 文档 `@Security` 与实现中间件一致；敏感接口见安全红线 |

**禁止：** 前端改路径/字段而后端未更新注解；Hand-written 前端类型与 swagger 长期漂移。

---

## 七、安全与红线（交叉引用）

完整条款见 [`security-bk-redlines.md`](security-bk-redlines.md)。接口层最低要求：

| 红线 | OpenAPI / Gin 落地 |
|------|-------------------|
| 外部输入校验 | `@Param` 描述 + Handler 内强校验（长度/枚举/格式） |
| 敏感接口鉴权 | `@securityDefinitions` + 路由 `@Security` + 路由组 Auth 中间件三者一致 |
| 敏感数据 | 响应 schema 不含 password/token；日志不打印 body 敏感字段 |

---

## 八、常见陷阱

| # | 陷阱 | 解决方案 |
|---|------|---------|
| 1 | 注解与路由不一致 | 改 `@Router` 后跑 apidocs 并 diff |
| 2 | 手改 swagger.json | 只重新 `swag init` |
| 3 | int64 前端精度问题 | JSON string + 前端 coerce |
| 4 | 双层 data 嵌套 | 对齐 Gin 响应 helper 与前端拦截器 |
| 5 | 生产暴露 Swagger UI | 环境开关或 IP 限制 |
| 6 | 未文档化的 breaking change | MR 标注并升 `/api/v2` 或 `info.version` |
| 7 | 有鉴权无 `@Security` | 补注解并重生；与中间件白名单对齐 |
| 8 | 宣称漂移已检但无脚本 | 只陈述已跑 `apidocs`；勿虚构 CI 门禁 |

---

## 九、质量保证

| 检查项 | 说明 |
|--------|------|
| `make apidocs` | 改注解后必须成功生成 |
| `make build` / `go build` | 确保 `docs/docs.go` 编译通过 |
| 契约漂移 CI | **有**则必须绿；**无**不得在报告中写「契约校验通过」 |
| 契约 review | MR 同时审注解、`@Security`、Handler 行为、swagger diff |
| 前端 | 按前端仓 `package.json` scripts 验证；**无** vitest 不得虚构 |

---

## 规范落地优先级

> 文首「本次必读决策表」为 P0/P1 快速入口；下表为完整落地优先级。

| 优先级 | 条款类型 | 落地方式 |
|--------|----------|----------|
| P0 机器门禁 | `make apidocs` / swag 生成成功；`make build`；若有漂移检测则跑 | CI 或提交前执行；无脚本则至少本地 `swag init` 可过 |
| P0 机器门禁 | 安全红线（鉴权、输入校验、密钥）；`@Security` 与中间件一致 | 见 [`security-bk-redlines.md`](security-bk-redlines.md) |
| P1 Agent 必读 | 注解、`@Security`、生成物禁改、envelope、版本策略 | 改 API 前按节 Read；IDE `standards-api` 督促加载 |
| P2 参考 | gin-swagger UI、Tag 命名、示例 payload | 可偏离，偏离时在 MR 说明 |

完整条款以本文件为准；短 Rules / AGENTS 门闩不含正文复述。
