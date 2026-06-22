# 词汇表（Glossary）

> 本项目核心概念、术语和缩写定义。Agent 和人类成员均以此为术语的唯一解释来源。

## Harness Engineering 核心概念

| 术语 | 英文 | 定义 |
|------|------|------|
| 驾驭工程 | Harness Engineering | 为 AI Agent 构建可靠运行环境的系统工程，通过工具链、约束和验证机制保障任务正确完成 |
| 上下文工程 | Context Engineering | 管理 Agent 知识来源与上下文披露策略，确保获取准确、及时、适量的信息 |
| 架构约束 | Architectural Constraints | 通过分层与依赖规则约束代码结构，防止 Agent 产生不可维护的变更 |
| 熵管理 | Entropy Management | 通过自动化检测与园艺机制控制系统复杂度增长，保持文档与代码一致 |
| 工具能力 | Tooling | Agent 可调用的 Skill、MCP、CLI 工具清单及其稳定性保障策略 |
| 执行与验证 | Execution & Verification | Agent Loop 执行循环与预完成强制验证机制 |

## 架构与设计模式

| 术语 | 英文 | 定义 |
|------|------|------|
| ADR | Architecture Decision Record | 架构决策记录，存储于 `docs/adr/`（含 ADR-0001/0002） |
| Reconcile 循环 | Reconcile Loop | controller-runtime 核心模式：获取资源 → 业务逻辑 → 更新 Status，必须幂等 |
| 渐进式披露 | Progressive Disclosure | AGENTS.md → harness/README → 详细文档的三层上下文加载策略 |
| Parse, Don't Validate | Parse, Don't Validate | 在数据边界一次性解析为强类型，后续代码不重复验证 |

## Controller 与 CRD 术语

| 术语 | 英文 | 定义 |
|------|------|------|
| Ingress | networkextension.Ingress | BCS 自定义 Ingress CRD，描述 L4/L7 负载均衡规则 |
| Listener | networkextension.Listener | 由 Ingress 转换生成的监听器 CRD，对接云 LB |
| PortPool | networkextension.PortPool | 节点端口池 CRD，管理宿主机端口分配 |
| PortBinding | networkextension.PortBinding | 端口绑定 CRD，关联 Pod 与端口 |
| HostNetPortPool | networkextension.HostNetPortPool | hostNetwork Pod 动态端口池 CRD |
| Namespace Scope Exemption | Namespace Scope Exemption | 白名单 Namespace 可跨 NS 引用 Service 并使用全局云凭证 |

## 工具与平台

| 术语 | 英文/缩写 | 定义 |
|------|----------|------|
| project.json | project.json | 项目根目录 TAPD 配置文件，含 workspace_id 和 owner，供迭代流水线读取 |
| tapd-story-agent | tapd-story-agent | 单需求实现流水线调度 Agent，定义于 `.cursor/agents/tapd-story-agent.md` |
| speckit-executor-agent | speckit-executor-agent | 隔离的 Spec Kit 命令执行 Agent，定义于 `.cursor/agents/speckit-executor-agent.md` |
| controller-runtime | controller-runtime | K8s Operator 框架，提供 Manager、Client、Reconcile 基础设施 |
| go-restful | go-restful | HTTP WebService 框架，用于管理 API |
| blog | bcs-common/blog | BCS 统一日志库，本项目禁止使用 stdlib log 或 klog |
| Spec Kit | Spec Kit | 基于 `.specify/` 的需求规格化与 TDD 开发工具链 |
| TAPD | TAPD | 腾讯敏捷协作平台，迭代研发流水线的外部需求来源 |

## 工程实践术语

| 术语 | 英文 | 定义 |
|------|------|------|
| 文档园艺 | Harness Gardening | harness-gardening skill 执行的八维度文档一致性巡检 |
| 技术债预算 | Technical Debt Budget | 控制 TODO/FIXME 和架构违规增速的量化阈值 |
| 表驱动测试 | Table-Driven Test | Go 单元测试标准模式，测试用例以结构体切片组织 |
| 幂等性 | Idempotency | Reconcile 多次执行产生相同结果，不产生副作用 |

## 信号协议

| 术语 | 格式 | 定义 |
|------|------|------|
| 待补充标记 | `<!-- TODO: 待补充 -->` | Harness 文档中信息不足处的标准占位 |
| 自动生成段落 | `<!-- dev-map:auto -->` | dev map 中由工具自动维护的段落边界 |

## 业务领域术语

| 术语 | 英文 | 定义 |
|------|------|------|
| CLB | Cloud Load Balancer | 腾讯云负载均衡，tencentcloud 适配器主要对接对象 |
| NamespacedLB | Namespaced Load Balancer Client | 按 Namespace 隔离云凭证的客户端封装 |
| PortPool Cache | PortPool Cache | PortPool 端口分配状态的内存缓存，冷启动从 API Server 重建 |
| HostNetPortPool Cache | HostNetPortPool Cache | HostNetPortPool 端口分配内存缓存 |
| CertificateChecker | Certificate Checker | 周期性检查 Ingress SSL 证书剩余过期天数，上报 Prometheus 指标 |
| NamespacedSSL | Namespaced SSL Client | 按 Namespace 隔离的腾讯云 SSL 证书 API 客户端，支持豁免 NS 使用全局凭证 |
| Cert Binding | Certificate Binding | Ingress 证书配置展开后的去重单元，含 cert_id、cert_role、cert_scope 等维度 |

---

*持续补充中——遇到新术语时请直接在对应分类下添加。*
