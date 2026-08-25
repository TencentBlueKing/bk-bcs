# 上下文工程（Context Engineering）

> 目标：让 Agent "知道该知道的信息"——确保 Agent 在任务执行中能获取准确、及时、适量的上下文。

## 1. 知识来源定义

### 1.1 唯一知识来源（Single Source of Truth）

| 知识类型 | 存储位置 | 维护责任人 | 更新频率 |
|---------|---------|-----------|---------|
| 项目入口与硬约束 | 根 `AGENTS.md`（短头 + 局部入口索引） | BCS 平台团队 | 入口约定变更时 |
| 工作单元入口 | 非根 `**/AGENTS.md`（局部约定优先） | 模块 Owner | 组件边界变更时 |
| 架构设计 | `docs/overview/README.md`、`docs/features/` | BCS 平台团队 | 架构/特性变更时 |
| API 说明 | `docs/apidoc/`、`docs/openapi/`、各服务 `proto/` | 对应服务 Owner | 接口变更时 |
| 编码/技术规范 | `docs/standards/`（按加载预算按节读） | harness 同步预设 | 技术栈/规范变更时 |
| 前端规范 | `docs/standards/frontend-vue2.md` | 预设库 | 技术栈变更时 |
| 接口规范 | `docs/standards/api-grpc-gateway.md` 及分册 | 预设库 | 协议变更时 |
| 后端规范 | `docs/standards/backend-gin.md` | 预设库 | 技术栈变更时 |
| 安全红线 | `docs/standards/security-bk-redlines.md` | 预设库 | 红线变更时 |
| 代码评审 | `docs/standards/quality-code-review.md` 及分册 | 预设库 | 按需（Review 任务） |
| 贡献与提交 | `CONTRIBUTING.md`、`docs/specification/` | BCS 平台团队 | 流程变更时 |
| 业务规范 | `docs/business-standards/` | 业务 Owner | 按 tags/scenarios 增补 |
| 组件级 harness | 如 `bcs-ingress-controller/docs/harness/` | 该组件 Owner | 与组件同步 |

### 1.2 禁止的知识来源

以下渠道的信息不应作为 Agent 决策依据（容易过时或缺乏版本控制）：
- 即时通讯记录（飞书、微信、企业微信等）
- 未纳入版本控制的外部 Wiki 口述摘要
- 口头约定或未归档会议记录
- 个人本机绝对路径、未跟踪的 `project.json` 内容当作仓库事实

## 2. 渐进式上下文披露

### 2.1 三层结构

```
第一层（入口）：根 AGENTS.md（短头 + 局部入口索引）
  └── 改某路径前：阅读向上最近的工作单元 AGENTS.md（局部优先于根）

第二层（导航）：docs/harness/README.md、docs/standards/README.md、docs/overview
  └── 对第一层只做指针/短摘要，不复制大段局部 AGENTS

第三层（详情）：代码、局部 AGENTS 细则、standards 分册、docs/features、proto
  └── 按任务按节加载，禁止默认全文灌入
```

### 2.2 上下文预算管理

- Agent 的上下文窗口视为有限资源，需精心管理
- 优先加载与当前任务直接相关的文档与最近工作单元 `AGENTS.md`
- 大文件（>300 行）通过目录索引定位相关段落，避免全量加载
- `docs/standards/` 长规范必须按「加载预算」按节 Read

## 3. 动态上下文接入

### 3.1 实时数据源

| 数据源 | 接入方式 | 用途 | 刷新频率 |
|-------|---------|------|---------|
| TAPD 需求/迭代 | TAPD MCP（`get_stories_or_tasks` 等） | 澄清、规划、实现流水线 | 任务开始时拉取 |
| 工蜂 MR/Issue | 工蜂 MCP 只读接口 | 关联代码评审与 Issue | 按任务 |
| 集群对象 | 运行中集群 / kubectl（需授权） | 排查 Operator 行为 | 调试时 |
<!-- TODO: 待补充仓库级 TAPD workspace_id（根 project.json 被 gitignore） -->

### 3.2 可观测性数据

| 数据类型 | 工具 | 访问方式 |
|---------|------|---------|
| 应用日志 | 各服务 `blog` / 组件日志 | 本地运行或集群 `kubectl logs` |
| 性能指标 | Prometheus（部分 Operator 已暴露） | 组件 metrics 端口；见局部 AGENTS |
| 链路追踪 | 以组件实际接入为准 | <!-- TODO: 待补充统一 APM 入口 --> |

## 4. 上下文更新机制

### 4.1 触发条件

- 代码架构发生重大变更（新增模块、调整分层）
- API / proto 接口新增或变更
- 工作单元边界或构建/测试命令变更
- 依赖的外部系统（云厂商、IAM、CMDB）发生变更

### 4.2 更新流程

1. 变更方在 PR 中同步更新相关文档与（若影响入口）工作单元 `AGENTS.md`
2. Code Review 时检查文档是否同步更新
3. 可用「文档巡检 / harness gardening」扫描根 harness 与代码的一致性

## 检查清单

- [x] 所有知识类型都有明确的存储位置
- [x] 根 AGENTS 短头精简，并索引已发现的工作单元
- [x] 已写明局部优先（nearest）
- [ ] 仓库级 TAPD `workspace_id` 未入库（`project.json` gitignore）
- [ ] 统一可观测性入口待补充
