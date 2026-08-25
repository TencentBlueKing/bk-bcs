# 后端开发规范（go-micro）

> **误匹配风险：** 仅当 `go.mod` require 含 `go-micro.dev`（或项目声明的 go-micro 模块）时启用。
> 特征目录常为 `api/proto/`、`cmd/*-server/`；**不要**因存在任意 `.proto` 就选用本预设（可能是 grpc-gateway / tRPC）。
> 下文目录树为典型布局示例，以仓库实际为准。

> 通用后端开发规范文档，适用于基于 Go + gRPC + go-micro 微服务架构的项目。

---

## 分册目录

| 分册 | 说明 |
|------|------|
| [01-技术栈要求.md](./backend-go-micro/01-技术栈要求.md) | 技术栈要求 |
| [02-项目结构.md](./backend-go-micro/02-项目结构.md) | 项目结构 |
| [03-proto-文件规范.md](./backend-go-micro/03-proto-文件规范.md) | Proto 文件规范 |
| [04-代码生成.md](./backend-go-micro/04-代码生成.md) | 代码生成 |
| [05-分层架构.md](./backend-go-micro/05-分层架构.md) | 分层架构 |
| [06-服务注册与发现.md](./backend-go-micro/06-服务注册与发现.md) | 服务注册与发现 |
| [07-异步任务开发规范.md](./backend-go-micro/07-异步任务开发规范.md) | 异步任务开发规范 |
| [08-配置管理.md](./backend-go-micro/08-配置管理.md) | 配置管理 |
| [09-日志规范.md](./backend-go-micro/09-日志规范.md) | 日志规范 |
| [10-错误处理.md](./backend-go-micro/10-错误处理.md) | 错误处理 |
| [11-测试规范.md](./backend-go-micro/11-测试规范.md) | 测试规范 |
| [12-构建与部署.md](./backend-go-micro/12-构建与部署.md) | 构建与部署 |
| [13-安全规范.md](./backend-go-micro/13-安全规范.md) | 安全规范 |

---

## 章节快速索引

> 接入仓 `docs/standards/README.md` 的「章节快速索引」会汇总本入口与各分册标题；按任务 **Read 对应分册**（可用 offset/limit），禁止默认全文灌入所有分册。
