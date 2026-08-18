# 架构约束（Architectural Constraints）

> 目标：让 Agent "做正确的事"——通过刚性约束确保代码结构的一致性和可维护性。

本仓是多组件 monorepo。**仓库级**只定义跨组件边界；组件内部层次以最近工作单元 `AGENTS.md` 为准，本文件只做摘要与指针，不另编冲突分层。

仓库级架构说明见 [`docs/overview/README.md`](../overview/README.md)（网关路由、Storage、Watch、KubeAgent / MesosDriver）。

## 1. 分层架构模型

### 1.1 仓库级层次（依赖只允许向下）

```
bcs-ui/            控制台（最上层，面向用户）
  ↓ HTTP / 网关
bcs-services/      控制面微服务
bcs-scenarios/     场景化服务
  ↓ 共享库 / 集群 API
bcs-runtime/       集群内 Operator / Watch / 网络
  ↓
bcs-common/        共享库（不得反向依赖 services / ui / runtime 业务包）
```

`install/`、`scripts/`、`docs/` 不参与运行时依赖图；Agent 不得把部署清单改成业务逻辑，也不得把 `bcs-common` 改成依赖某个具体服务。

### 1.2 依赖规则

- 新代码优先落在**被修改功能所属组件根**（该组件自己的 `go.mod` / `package.json`），禁止为图方便把业务逻辑塞进无关服务
- `bcs-common` 只放真正跨服务复用的能力；引入新的重量级依赖须评估所有引用方
- 上层（ui / services）不得 import 另一服务的 `internal/`；跨服务走已有 API / proto / 网关
- 同层微服务之间不得形成包级循环依赖
- 生成物（`zz_generated.*`、swagger 生成文件、CRD bases）禁止手改——以最近工作单元 AGENTS 为准

### 1.3 目录与层次映射

| 层 | 目录 | 职责 | 允许的依赖 |
|----|------|------|-----------|
| 共享库 | `bcs-common/` | 日志、配置、加密、公共 API 客户端 | 第三方库；不得依赖具体 bcs-services |
| 运行时 | `bcs-runtime/` | Operator、CRD、网络、Watch | `bcs-common`、K8s API |
| 控制面 | `bcs-services/` | 集群/项目/Helm/监控等微服务 | `bcs-common`、必要时 runtime CRD 类型 |
| 场景 | `bcs-scenarios/` | 特定场景微服务 | `bcs-common`、外部云/蓝鲸 API |
| 前端 | `bcs-ui/frontend/` | Vue 2 控制台 | 网关 HTTP API，不直连集群 |
| 部署 | `install/` | 配置与 chart | 无编译期依赖 |

### 1.4 工作单元内部层次（指针，不覆写）

| 组件 | 入口 | 内部约束摘要 |
|------|------|-------------|
| bcs-ingress-controller | [`.../bcs-ingress-controller/AGENTS.md`](../../bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/AGENTS.md) | controller-runtime；Reconcile 幂等；日志用 `bcs-common/common/blog`；禁止硬编码 annotation |
| bcs-drplan-controller | [`.../bcs-drplan-controller/AGENTS.md`](../../bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/AGENTS.md) | Kubebuilder 目录不可挪；禁止手改 generated manifests |
| bcs-terraform-bkprovider | [`.../bcs-terraform-bkprovider/AGENTS.md`](../../bcs-scenarios/bcs-terraform-bkprovider/AGENTS.md) | go-micro gRPC；handler / middleware / proto 分离 |

无工作单元 `AGENTS.md` 的服务：先读该目录 `README.md` 与相邻实现，保持与同目录既有分层一致。

## 2. 自定义 Linter 规则

### 2.1 规则清单

| 规则编号 | 名称 | 描述 | 修复指引 |
|---------|------|------|---------|
| ARCH-001 | 禁止反向依赖 | `bcs-common` 不得 import `bcs-services` / `bcs-ui` / 具体 runtime 业务包 | 把类型下沉到 common 或改走 API |
| ARCH-002 | 禁止跨服务直连 internal | 服务 A 不得引用服务 B 的 `internal/` | 经 proto / 网关 / 已有 SDK |
| ARCH-003 | 禁止手改生成物 | 不编辑 `zz_generated.*`、CRD bases、proto 生成 Go | 改源（types/proto）后跑该组件 `make generate/proto` |
| ARCH-004 | 变更范围收敛 | 一次任务只改相关组件与其测试/文档 | 撤销无关文件；大特性拆 PR（见 CONTRIBUTING） |

Go 风格以 [`docs/specification/blueking-golang-code-conduct1.0.1.md`](../specification/blueking-golang-code-conduct1.0.1.md) 与组件 `.golangci.yml` 为准。

### 2.2 错误信息格式

```
[ARCH-002] 禁止跨服务引用 internal：bcs-services/foo 引用了 bcs-services/bar/internal/...
修复方式：改为调用 bar 已有 API / proto 客户端
参考文档：docs/harness/architectural-constraints.md#依赖规则
```

## 3. Parse, Don't Validate

### 3.1 原则

在数据进入系统的边界处，将原始数据**解析**为强类型的领域模型，后续代码只操作解析后的类型。

### 3.2 数据边界

| 边界 | 输入类型 | 解析目标 | 处理位置 |
|------|---------|---------|---------|
| HTTP / gRPC 请求 | JSON / protobuf | 请求结构体 + 校验 | Handler / proto 入口 |
| K8s CR | unstructured / YAML | CRD Go 类型 | Reconcile 开头 |
| 配置文件 | YAML / JSON / flag | 配置 struct | 进程启动 |
| 云厂商 API 响应 | JSON | 适配层 DTO | cloud client / adapter |
| 前端表单 | 用户输入 | 校验后的视图模型 | 请求封装层（禁止只在 UI 校验） |

安全红线（输入校验 / 鉴权 / 加密）见 `docs/standards/security-bk-redlines.md` 相关节。

## 4. 架构决策记录（ADR）

### 4.1 管理方式

- 仓库级特性设计：`docs/features/`（CONTRIBUTING 要求新特性归档设计文档）
- 已有组件 ADR 示例：`bcs-ingress-controller/docs/adr/`
- 命名建议：`NNNN-标题.md`；Agent 做架构决策前先检索已有 ADR / features 文档

### 4.2 ADR 模板

```
# NNNN. 决策标题

## 状态：已接受 / 已废弃 / 已替代

## 背景
为什么需要做这个决策？

## 决策
选择了什么方案？

## 后果
带来了哪些影响（正面和负面）？
```

## 检查清单

- [x] 仓库级分层与依赖方向已定义
- [x] 工作单元已有分层仅摘要 + 指针
- [x] 至少 3 条架构规则
- [x] 数据边界 Parse 策略已明确
- [ ] 仓库级 `docs/adr/` 尚未建立（现有设计在 `docs/features/`）
