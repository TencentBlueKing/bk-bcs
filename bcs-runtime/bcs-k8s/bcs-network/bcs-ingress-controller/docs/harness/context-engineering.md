# 上下文工程（Context Engineering）

> 目标：让 Agent 获取准确、及时、适量的上下文，避免信息过载或知识过时。

## 1. 知识来源定义

### 1.1 唯一知识来源（Single Source of Truth）

| 知识类型 | 存储位置 | 维护责任人 | 更新频率 |
|---------|---------|-----------|---------|
| 项目入口与速查 | `AGENTS.md` | 研发团队 | 架构/流程变更时 |
| Harness 规范 | `docs/harness/` | 研发团队 + harness-gardening | 规范变更时 |
| 开发地图 | `docs/dev-map/` | harness-generating / harness-gardening | 代码结构变更时 |
| 技术规范 | `docs/standards/` | 预设库同步 | 技术栈变更时 |
| 功能设计 | `specs/{feature-id}/` | 功能负责人 | 新功能启动时 |
| 架构决策（ADR） | `docs/adr/` | 研发团队 | 架构变更时 |
| CRD 类型定义 | `../../kubernetes/apis/networkextension/v1/` | K8s API 团队 | CRD 变更时 |
| CRD YAML 清单 | `../../kubernetes/config/crd/bases/` | K8s API 团队 | CRD 变更时 |
| 共享常量 | `internal/constant/constant.go` | 研发团队 | 新增 Annotation 时 |
| 后端规范 | `docs/standards/backend-k8s-operator.md` | 研发团队 | 技术栈变更时 |
| HTTP API 规范 | `docs/standards/api-go-restful.md` | 研发团队 | 接口变更时 |
| 安全规范 | `docs/standards/security-bk-redlines.md` | 安全团队（预设同步） | 预设更新时 |
| 代码评审规范 | `docs/standards/quality-code-review.md` | 研发团队（预设同步） | 预设更新时 |

### 1.2 禁止的知识来源

以下渠道的信息不应作为 Agent 决策依据：

- 即时通讯记录（飞书、微信等）
- 未纳入版本控制的外部 Wiki
- 口头约定或会议记录
- 过时的 `.cursor/rules/` 中与 `docs/harness/` 冲突的内容

## 2. 渐进式上下文披露

### 2.1 三层结构

```
第一层（入口）：AGENTS.md / docs/harness/README.md
  ├── 项目概述与目录结构（~30行）
  ├── 关键构建命令（~10行）
  ├── 规范导航索引
  └── 总计不超过 100 行

第二层（导航）：docs/harness/、docs/standards/、docs/dev-map/
  ├── 各组件文档控制在 300 行以内
  └── 按任务类型按需加载

第三层（详情）：代码注释 + specs/ 设计文档 + CRD 定义
  └── 仅在需要时访问，不主动全量加载
```

### 2.2 上下文预算管理

- 优先加载与当前任务直接相关的文档（如改 Controller 则加载 architectural-constraints + dev-map 对应模块）
- 大文件（>300行）通过 dev-map 索引定位相关段落
- `AGENTS.md` 历史详细内容已拆分至 harness 组件文档，避免重复加载

## 3. 动态上下文接入

### 3.1 实时数据源

| 数据源 | 接入方式 | 用途 | 刷新频率 |
|-------|---------|------|---------|
| K8s 集群状态 | kubectl / BCS MCP | 验证 Controller 行为、排查 Reconcile 问题 | 按需 |
| TAPD 需求 | TAPD MCP `stories_get` | 迭代研发流水线需求拉取 | 按需 |
| 代码索引 | Codegraph MCP | 符号搜索、调用链追踪、影响分析 | 实时（~1s 延迟） |
| Git 变更 | `git diff` / `git log` | 代码评审、提交范围确认 | 按需 |

### 3.2 可观测性数据

| 数据类型 | 工具 | 访问方式 |
|---------|------|---------|
| 应用日志 | blog（bcs-common） | 集群 Pod 日志 `kubectl logs` |
| Prometheus 指标 | prometheus/client_golang | namespace `bkbcs_ingressctrl`，`/metrics` 端点 |
| Controller 事件 | k8s EventRecorder | `kubectl get events` |
| HTTP 健康检查 | go-restful readiness | `internal/httpsvr/readiness_probe.go` |

## 4. 上下文更新机制

### 4.1 触发条件

- 新增/删除 Controller 或 CRD
- 云适配器或 Namespace Scope 逻辑变更
- HTTP API 路由新增
- Skill/MCP 工具链变更
- 完成 Spec Kit 功能开发后

### 4.2 更新流程

1. 变更方在 PR 中同步更新 `AGENTS.md`、`docs/dev-map/` 或 `docs/harness/` 相关章节
2. Code Review 时检查文档是否同步（参考 `docs/standards/quality-code-review.md`）
3. 对我说「文档巡检」触发 harness-gardening 八维度扫描
4. 新增功能设计文档放入 `specs/{feature-id}/`

## 检查清单

- [x] 知识类型均有明确存储位置
- [x] AGENTS.md 控制在 100 行以内
- [x] 动态数据源已定义接入方式
- [x] ADR 目录 `docs/adr/` 已建立
- [x] 上下文更新机制已文档化
