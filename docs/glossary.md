# 词汇表（Glossary）

> 本项目涉及的核心概念、术语和缩写定义。Agent 和人类成员均以此为术语的唯一解释来源。

## Harness Engineering 核心概念

| 术语 | 英文 | 定义 |
|------|------|------|
| 驾驭工程 | Harness Engineering | 为 AI Agent 构建可靠运行环境（上下文、约束、工具、验证），而非优化模型本身 |
| 上下文工程 | Context Engineering | 规定 Agent 从何处、以何种预算获取准确适量的项目知识 |
| 架构约束 | Architectural Constraints | 分层、依赖方向与边界规则，防止 Agent 写出结构错误的代码 |
| 熵管理 | Entropy Management | 用园艺、CI 与技术债机制控制系统文档/架构随时间劣化 |
| 工具能力 | Tooling | Agent 可调用的 Skill / MCP / CLI 契约清单（不含本机就绪状态） |
| 执行与验证 | Execution & Verification | Agent Loop、预完成检查与可观测性，确保任务被正确完成 |
| 工作单元 | Work Unit | 带独立 `AGENTS.md` 的子树；局部约定优先于仓库根 |

## 架构与设计模式

| 术语 | 英文 | 定义 |
|------|------|------|
| 控制面 | Control Plane | `bcs-services/` 中面向用户与集群生命周期的微服务集合 |
| 运行时 | Runtime | `bcs-runtime/` 中跑在集群内的 Operator、Watch、网络组件 |
| 共享库 | Shared Library | `bcs-common/`，被各服务引用的公共代码，禁止反向依赖上层服务 |
| API 网关 | API Gateway | 外部请求入口，按集群类型路由到 Mesos 或 Kubernetes 实现 |
| 数据监视 | Data Watch | 将集群对象同步到 BCS Storage 的组件 |
| 原地升级 | InplaceUpdate | 不重建 Pod 更新容器镜像/配置的发布能力 |
| Parse, Don't Validate | Parse, Don't Validate | 在系统边界把原始输入解析为强类型，后续不再重复校验 |

## Skill 相关术语

| 术语 | 英文 | 定义 |
|------|------|------|
| Skill | Skill | 按触发词加载的 Agent 能力包，入口为 `SKILL.md` |
| 安装根 | Skill Install Root | 运行时探测到的项目级 Skill 目录（本仓为 `.agents/skills`） |
| 基线工具 | Harness Baseline | 权威清单中的 TAPD 流水线 / 评审 / Harness 工具 |
| 项目自有工具 | Project-owned Tools | git 跟踪且不在基线白名单中的组件级 Skill |

## 工具与平台

| 术语 | 英文/缩写 | 定义 |
|------|----------|------|
| TAPD | TAPD | 需求与迭代管理平台；Agent 经 MCP 读写需求 |
| 工蜂 | Gongfeng | 腾讯 Git 托管；MR/Issue 经 MCP 查询 |
| graphify | graphify | 代码知识图谱工具；本仓尚未安装，开发地图暂不可用 |
| golangci-lint | golangci-lint | Go 静态检查，提交前建议执行 |
| GameDeployment | GameDeployment | 面向游戏无状态实例的增强 Deployment |
| GameStatefulSet | GameStatefulSet | 面向游戏有状态实例的增强 StatefulSet |

## 工程实践术语

| 术语 | 英文 | 定义 |
|------|------|------|
| 文档园艺 | Document Gardening | 扫描 harness/standards 与代码一致性并修复漂移 |
| ADR | Architecture Decision Record | 记录架构决策的背景、选择与后果 |
| 加载预算 | Loading Budget | 规定某类任务应 Read 的规范章节，禁止默认全文灌入 |
| 门闩 | Gate | 编码前必须执行的最短检查步骤（见根 AGENTS） |

## 信号协议

| 术语 | 格式 | 定义 |
|------|------|------|
| 阻塞 | `blocked: <原因>` | Agent 无法继续，需用户决策或补齐环境 |
| 待确认 | `<!-- TODO: 待补充 -->` | 规范中信息不足的占位，不得当作已确认事实 |

## 业务领域术语

| 术语 | 英文 | 定义 |
|------|------|------|
| BCS | BlueKing Container Service | 蓝鲸容器调度平台 |
| 联邦集群 | Federation | Host Cluster 纳管多个 Member Cluster 的多集群方案 |
| 模板集 | TemplateSet | BCS 图形化/表单化应用编排产物 |
| KubeAgent | KubeAgent | 将 Kubernetes 集群注册到 BCS 网关的组件 |

---

*持续补充中——遇到新术语时请直接在对应分类下添加。*
