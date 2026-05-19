<!--
Sync Impact Report
===================
- Version change: 1.1.0 → 1.2.0
- Added rules:
  - 错误 MUST 被显式处理或记录；仅检查而不处理时 MUST 添加注释说明原因（I 节）
- Removed sections: N/A
- Follow-up TODOs: none
-->

# bcs-ingress-controller Constitution

## Core Principles

### I. Go 语言工程规范

本项目使用 Go 语言开发，所有代码 MUST 遵循以下规范：

- 代码 MUST 通过 `golangci-lint` 静态检查，零容忍 lint 错误
- 代码格式化 MUST 使用 `gofmt` / `goimports`，禁止提交未格式化的代码
- 包命名 MUST 使用小写单词，禁止下划线或驼峰命名（遵循 Go 官方 Effective Go 规范）
- 错误处理 MUST 显式处理，禁止忽略 error 返回值（`_` 丢弃 error 除外需注释说明原因）
- 错误 MUST 被处理（记录日志、返回、重试等），仅检查错误而不做任何处理（静默忽略）MUST 添加注释说明原因；条件分支中忽略错误分支时 MUST 至少记录一条日志
- 导出类型和函数 MUST 有 GoDoc 注释，注释语言为英文
- 项目依赖管理 MUST 使用 Go Modules，`go.mod` 和 `go.sum` MUST 保持同步
- 并发编程 MUST 使用 `context.Context` 进行生命周期管理，禁止裸 goroutine 泄漏

**理由**: Go 语言的工程化规范是保证代码一致性和可维护性的基础。BCS 项目作为大型开源项目，
统一的 Go 编码风格能降低多人协作的认知成本。

### II. Kubernetes 控制器开发规范

本项目是 Kubernetes 控制器，MUST 遵循 K8s 生态的开发惯例：

- 控制器 MUST 基于 `controller-runtime` 框架实现 Reconcile 模式
- CRD（Custom Resource Definition）MUST 使用 `kubebuilder` 标记进行代码生成
- 资源状态管理 MUST 区分 `spec`（期望状态）和 `status`（实际状态）
- Finalizer 的添加与移除 MUST 遵循 K8s Finalizer 模式，确保资源清理的可靠性
- 控制器 MUST 实现幂等性：对同一资源的多次 Reconcile MUST 产生相同的结果
- 事件记录 MUST 使用 K8s Event Recorder，关键操作需记录 Normal/Warning 事件
- RBAC 权限 MUST 使用 `kubebuilder` 注解声明，遵循最小权限原则
- Leader Election MUST 启用以支持高可用部署
- 对外部云资源（如 CLB/LB）的操作 MUST 具备重试与降级机制

**理由**: 遵循 K8s 控制器的标准模式可以确保控制器的稳定性和可预测性。
bcs-ingress-controller 管理着网络负载均衡等关键基础设施资源，
任何不符合 K8s 惯例的实现都可能导致资源泄漏或状态不一致。

### III. 代码质量优先

代码质量是本项目的首要关注点，所有变更 MUST 满足以下标准：

- 单元测试覆盖率 MUST 不低于增量代码的 60%，核心逻辑（Reconcile、云操作）SHOULD 达到 80%
- 函数圈复杂度 MUST 不超过 15，超过时 MUST 拆分为子函数
- 单个函数体 SHOULD 不超过 80 行（不含注释和空行），超过时需评估是否可拆分
- 单个文件 SHOULD 不超过 500 行，超过时需评估是否可按职责拆分
- 公共接口的变更 MUST 保持向后兼容，破坏性变更 MUST 在 PR 描述中明确标注
- 所有 error MUST 包含足够的上下文信息（使用 `fmt.Errorf` 或 `errors.Wrap`）
- Magic number 和 Magic string MUST 定义为命名常量
- 重复代码块（超过 5 行的相同逻辑）MUST 抽取为公共函数
- 测试函数名（`func TestXxx`）长度 MUST 不超过 35 个字符；超长时使用缩写（如 `Alloc`/`Contig`/`Seg`）

**理由**: 代码质量直接影响系统的可靠性和可维护性。
bcs-ingress-controller 作为网络基础设施的关键组件，
任何代码缺陷都可能导致服务不可达或网络中断。高质量的代码是可靠运行的前提。

### IV. 英文代码注释规范

代码中的所有注释 MUST 使用英文编写：

- GoDoc 注释 MUST 使用英文，以 exported 名称开头（符合 Go 惯例）
- 行内注释 MUST 使用英文，简洁清晰地说明 **why** 而非 **what**
- TODO/FIXME/HACK 注释 MUST 使用英文，格式为 `// TODO(author): description`
- 文件头部版权声明 MUST 保持现有的 Tencent 开源协议英文模板
- 变量、函数、结构体命名 MUST 使用英文，表意明确，禁止拼音命名
- 注释 MUST NOT 重复代码本身已表达的信息（如 `// increment counter` 是冗余注释）
- 复杂算法或业务逻辑 SHOULD 添加注释说明设计意图和权衡取舍

**理由**: 英文注释确保代码对国际开源社区的友好性。BCS 是 Tencent 的开源项目，
托管于 GitHub，英文注释是开源项目的通行标准，也便于全球开发者参与贡献。

### V. 中文沟通语言

项目协作过程中的非代码沟通 MUST 使用中文：

- PR/MR 描述和讨论 MUST 使用中文
- Issue 描述和讨论 MUST 使用中文
- Commit message MUST 使用中文描述（type 标记使用英文，如 `fix: [模块名] 修复XXX问题`）
- 设计文档、技术方案 MUST 使用中文编写
- Code Review 评论 MUST 使用中文
- AI 辅助工具的交互 MUST 使用中文
- 项目 Constitution、Specification 等治理文档 MUST 使用中文

**理由**: 团队主要成员为中文母语使用者，中文沟通能最大化信息传递效率，
减少误解。明确沟通语言的边界（代码=英文，协作=中文）可以避免混乱。

## 技术栈与依赖约束

本项目的技术栈和依赖 MUST 遵循以下约束：

- **运行时**: Go 1.20+（跟随 BCS 主仓库版本）
- **控制器框架**: `sigs.k8s.io/controller-runtime`，版本跟随 BCS 主仓库统一管理
- **K8s 客户端**: `k8s.io/client-go`，版本与 controller-runtime 兼容
- **日志框架**: `github.com/Tencent/bk-bcs/bcs-common/common/blog`（BCS 统一日志组件）
- **HTTP 框架**: `github.com/emicklei/go-restful`（运维接口）
- **监控指标**: `github.com/prometheus/client_golang`（Prometheus 指标暴露）
- **云厂商 SDK**: 按需引入（腾讯云、AWS、GCP、Azure），MUST 通过接口抽象隔离
- 新增第三方依赖 MUST 经过团队评审，优先选择社区活跃度高、License 兼容的库
- 禁止引入与现有功能重叠的冗余依赖

## 开发流程与质量门禁

所有代码变更 MUST 通过以下流程：

- 代码变更 MUST 提交 PR/MR，禁止直接推送到主分支
- PR/MR MUST 至少经过 1 位 Reviewer 审核通过后方可合并
- CI 流水线 MUST 包含：编译检查、lint 检查、单元测试、敏感信息扫描
- Commit message MUST 遵循 BCS 项目的提交规范（参见 BCSDev iWiki 代码提交规范）
- 提交前 MUST 确保无敏感信息（IP、密码、密钥等）泄漏
- 重大架构变更 MUST 先编写设计文档，经团队讨论后再实施
- CRD 字段变更 MUST 保持向后兼容，废弃字段使用 `deprecated` 注释标记

## Governance

本 Constitution 是 bcs-ingress-controller 项目的最高治理准则：

- 所有 PR/MR 审核 MUST 验证变更是否符合本 Constitution 的原则
- Constitution 的修订 MUST 提交 PR 并经过团队讨论批准
- 版本号遵循语义化版本规范：
  - MAJOR：原则的移除或不兼容重定义
  - MINOR：新增原则或对现有原则的实质性扩展
  - PATCH：措辞澄清、排版修正等非语义性变更
- 任何与 Constitution 冲突的实践，以 Constitution 为准
- 复杂度 MUST 有合理理由，遵循 YAGNI（You Aren't Gonna Need It）原则

**Version**: 1.1.0 | **Ratified**: 2026-03-16 | **Last Amended**: 2026-03-27
