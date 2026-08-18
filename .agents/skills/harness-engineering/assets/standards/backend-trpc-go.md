# 后端开发规范（tRPC-Go）

> **启用前提（generating 必须已校验）：** 仓库存在 `proto/` 与 `stub/` 目录（或项目声明的等价生成目录）。
> 若服务为 Gin/其他 REST 且无上述目录，**不要使用本文件**；backend 分类保持未匹配（不设规范），或待专用预设入库后再生成。
> 下文目录树为典型 tRPC 布局**示例**；模块路径以 `go.mod` 的 module 与仓库实际目录为准。


> 通用后端开发规范文档，适用于基于 Go + tRPC-Go 微服务架构的项目。

## 本次必读决策表

> Agent：先读本表，再按任务 Read 对应分册；禁止默认全文灌入。完整条款见分册；落地优先级见下表。

| 类型 | 决策（摘要） | 详见 |
|------|-------------|------|
| 禁止 | 手改 `stub/`；生产环境直连 IP；Handler 内重复鉴权 | [02-项目结构](./backend-trpc-go/02-项目结构.md)、[14-构建与部署](./backend-trpc-go/14-构建与部署.md) |
| 必须 | Proto 单一事实源；下游调用透传上游 `ctx`；统一 `trpc-filter/auth` | [03-proto-文件规范](./backend-trpc-go/03-proto-文件规范.md)、[10-客户端调用规范](./backend-trpc-go/10-客户端调用规范.md)、[14-构建与部署](./backend-trpc-go/14-构建与部署.md) |
| 验证 | 跑 `golangci-lint` / `go test` / Makefile 已有目标（无则不得虚构通过） | [14-构建与部署](./backend-trpc-go/14-构建与部署.md)、下表 P0 |

---

## 规范落地优先级

> 文首「本次必读决策表」为 P0/P1 快速入口；下表为完整落地优先级。

| 优先级 | 条款类型 | 落地方式 |
|--------|----------|----------|
| P0 机器门禁 | `golangci-lint` / `go test` / Makefile 已有目标；密钥不进仓 | CI 或提交前跑仓库已有命令；无目标不得虚构通过 |
| P0 机器门禁 | Proto/stub 契约一致；禁止手改 `stub/` | `trpc-cli` 重新生成 + CI 校验（若有） |
| P1 Agent 必读 | 分层、ctx 透传、filter 鉴权、北极星服务发现 | 改后端前 Read 本文件相关节；IDE `standards-backend` 仅督促加载 |
| P2 参考 | Docker 镜像惯例、目录命名示例 | 可偏离，偏离时在 MR 说明 |

完整条款见各分册；短 Rules / AGENTS 门闩不含正文复述。

---

## 分册目录

| 分册 | 说明 |
|------|------|
| [01-技术栈要求.md](./backend-trpc-go/01-技术栈要求.md) | 技术栈要求 |
| [02-项目结构.md](./backend-trpc-go/02-项目结构.md) | 项目结构 |
| [03-proto-文件规范.md](./backend-trpc-go/03-proto-文件规范.md) | Proto 文件规范 |
| [04-代码生成-trpc-cli.md](./backend-trpc-go/04-代码生成-trpc-cli.md) | 代码生成（trpc-cli） |
| [05-分层架构.md](./backend-trpc-go/05-分层架构.md) | 分层架构 |
| [06-服务配置-trpc_go-yaml.md](./backend-trpc-go/06-服务配置-trpc_go-yaml.md) | 服务配置（trpc_go.yaml） |
| [07-服务注册与发现-北极星.md](./backend-trpc-go/07-服务注册与发现-北极星.md) | 服务注册与发现（北极星） |
| [08-服务发现扩展.md](./backend-trpc-go/08-服务发现扩展.md) | 服务发现扩展 |
| [09-filter-拦截器-规范.md](./backend-trpc-go/09-filter-拦截器-规范.md) | Filter（拦截器）规范 |
| [10-客户端调用规范.md](./backend-trpc-go/10-客户端调用规范.md) | 客户端调用规范 |
| [11-日志规范.md](./backend-trpc-go/11-日志规范.md) | 日志规范 |
| [12-错误处理.md](./backend-trpc-go/12-错误处理.md) | 错误处理 |
| [13-测试规范.md](./backend-trpc-go/13-测试规范.md) | 测试规范 |
| [14-构建与部署.md](./backend-trpc-go/14-构建与部署.md) | 构建与部署 |

---

## 章节快速索引

> 接入仓 `docs/standards/README.md` 的「章节快速索引」会汇总本入口与各分册标题；按任务 **Read 对应分册**（可用 offset/limit），禁止默认全文灌入所有分册。
