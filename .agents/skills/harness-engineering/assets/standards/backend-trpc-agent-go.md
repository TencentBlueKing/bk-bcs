# 后端开发规范（tRPC-Agent-Go）

> **启用前提（generating 必须已校验）：** `go.mod` 直接依赖含 `trpc-agent-go`（或等价模块路径）。
> 若仅为普通 tRPC-Go 服务而无 Agent 框架依赖，**不要使用本文件**，应改用 `backend-trpc-go.md`（满足其启用前提时）；否则 backend 分类不设规范。
> 下文覆盖率与目录树为条件性示例；以仓库实际 `proto/`/`stub/`（若有）与模块路径为准。

> 通用后端开发规范文档，适用于基于 Go + tRPC-Agent-Go 构建 AI Agent 微服务的项目。

---

## 分册目录

| 分册 | 说明 |
|------|------|
| [01-技术栈要求.md](./backend-trpc-agent-go/01-技术栈要求.md) | 技术栈要求 |
| [02-项目结构.md](./backend-trpc-agent-go/02-项目结构.md) | 项目结构 |
| [03-agent-定义规范.md](./backend-trpc-agent-go/03-agent-定义规范.md) | Agent 定义规范 |
| [04-tool-定义规范.md](./backend-trpc-agent-go/04-tool-定义规范.md) | Tool 定义规范 |
| [05-agentic-loop-调用链路.md](./backend-trpc-agent-go/05-agentic-loop-调用链路.md) | Agentic Loop（调用链路） |
| [06-服务配置-trpc_go-yaml.md](./backend-trpc-agent-go/06-服务配置-trpc_go-yaml.md) | 服务配置（trpc_go.yaml） |
| [07-llm-模型配置.md](./backend-trpc-agent-go/07-llm-模型配置.md) | LLM 模型配置 |
| [08-会话管理-session.md](./backend-trpc-agent-go/08-会话管理-session.md) | 会话管理（Session） |
| [09-日志与可观测性.md](./backend-trpc-agent-go/09-日志与可观测性.md) | 日志与可观测性 |
| [10-错误处理.md](./backend-trpc-agent-go/10-错误处理.md) | 错误处理 |
| [11-单元测试规范.md](./backend-trpc-agent-go/11-单元测试规范.md) | 单元测试规范 |
| [12-安全规范.md](./backend-trpc-agent-go/12-安全规范.md) | 安全规范 |
| [13-构建与部署.md](./backend-trpc-agent-go/13-构建与部署.md) | 构建与部署 |

---

## 章节快速索引

> 接入仓 `docs/standards/README.md` 的「章节快速索引」会汇总本入口与各分册标题；按任务 **Read 对应分册**（可用 offset/limit），禁止默认全文灌入所有分册。
