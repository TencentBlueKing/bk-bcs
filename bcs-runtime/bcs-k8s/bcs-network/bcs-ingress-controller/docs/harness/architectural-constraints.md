# 架构约束（Architectural Constraints）

> 目标：通过刚性约束确保 BCS Ingress Controller 代码结构一致、可维护。

## 1. 分层架构模型

### 1.1 层次定义

```
internal/constant + internal/option（配置与常量层，无业务依赖）
  ↓
internal/cloud + internal/*cache + internal/generator（领域与适配层）
  ↓
{name}controller/（Reconcile 控制器层）
  ↓
internal/httpsvr + internal/webhookserver（接入层）
  ↓
main.go（编排层，注册所有组件）
```

### 1.2 依赖规则

- 依赖只能**向下**流动：Controller 可依赖 internal/，internal/ 不得依赖 Controller
- `internal/constant` 和 `internal/option` 为**最底层**，不得 import 其他 internal 包
- 云适配器（`internal/cloud/*`）通过 `internal/cloud/interface.go` 抽象，Controller 不直接引用具体云 SDK
- SSL 证书 API（`internal/cloud/tencentcloud/sslclient.go`）通过 `internal/cloud/namespacedssl/` 按 Namespace 隔离，模式对齐 NamespacedLB
- 同层 cloud 子包（aws/azure/gcp/tencentcloud）之间**不得**互相引用
- CRD 类型定义在外部模块 `../../kubernetes/apis/`，通过 go module replace 引用

### 1.3 目录与层次映射

| 层 | 目录 | 职责 | 允许的依赖 |
|----|------|------|-----------|
| 常量/配置 | `internal/constant/`, `internal/option/` | Annotation Key、CLI 参数 | 仅标准库和外部配置库 |
| 领域逻辑 | `internal/generator/`, `internal/*cache/` | Ingress 转换、端口缓存 | constant, option, CRD types |
| 云适配 | `internal/cloud/` | 多云 LB SDK 封装 | constant, interface |
| 控制器 | `{name}controller/` | CRD Reconcile | internal/*, controller-runtime |
| 接入 | `internal/httpsvr/`, `internal/webhookserver/` | HTTP API、Admission | internal/*, go-restful |
| 编排 | `main.go` | 注册、初始化、启动 | 所有层 |
| 可观测 | `internal/metrics/` | Prometheus 指标 | 各层通过 helper 上报 |
| 检查 | `internal/check/` | 一致性巡检、SSL 证书过期检查 | cloud, cache, client, metrics |

## 2. 自定义 Linter 规则

### 2.1 规则清单

| 规则编号 | 名称 | 描述 | 修复指引 |
|---------|------|------|---------|
| ARCH-001 | 禁止反向依赖 | internal 包不得 import controller 包 | 将共享逻辑下沉到 internal |
| ARCH-002 | 禁止硬编码 Annotation | 不得使用字符串字面量作为 Annotation Key | 添加到 `internal/constant/constant.go` |
| ARCH-003 | 禁止直接 log/klog | 必须使用 `bcs-common/common/blog` | 替换 import 和调用 |
| ARCH-004 | Controller 必须幂等 | Reconcile 多次执行结果一致 | 先 Get 再比较，避免无条件 Update |
| ARCH-005 | 新 Controller 必须注册 | 创建 Controller 后必须在 main.go SetupWithManager | 参考 portpoolcontroller 模式注册 |
| ARCH-006 | initClient 保持纯分发 | main.go initClient 复杂度 ≤ 5 | 云初始化逻辑放入 initXxxClient |
| ARCH-007 | 函数名长度限制 | 所有函数名（含测试函数 `TestXxx`）不得超过 35 字符 | 缩短命名或使用缩写（如 `Alloc`/`Contig`） |
| ARCH-008 | 复杂度控制 | 函数圈复杂度 > 10 必须拆分 | 提取子函数降低复杂度 |
| ARCH-009 | 导出函数注释 | 导出函数（首字母大写）必须有英文 GoDoc 注释 | 补充以导出名称开头的 GoDoc 注释 |

### 2.2 错误信息格式

```
[ARCH-002] 违反常量规则：ingresscontroller/foo.go 使用了硬编码 annotation "networkextension.bkbcs.tencent.com/xxx"
修复方式：在 internal/constant/constant.go 添加 const 并引用
参考文档：docs/harness/architectural-constraints.md#规则清单
```

## 3. Parse, Don't Validate

### 3.1 原则

在数据进入系统的边界处，将原始数据解析为强类型，后续代码只操作解析后的类型。

### 3.2 数据边界

| 边界 | 输入类型 | 解析目标 | 处理位置 |
|------|---------|---------|---------|
| CLI 参数 | flag 字符串 | `ControllerOption` 结构体 | `internal/option/option.go` |
| Ingress Spec | CRD Spec | Listener 生成模型 | `internal/generator/` |
| Webhook 请求 | AdmissionReview | 验证/变更结果 | `internal/webhookserver/` |
| HTTP 请求 | JSON/Query | 领域查询参数 | `internal/httpsvr/` |
| 云 API 响应 | SDK Response | 统一 LB 模型 | `internal/cloud/*/helper.go` |
| SSL 证书 API | DescribeCertificate Response | 证书过期时间 | `internal/cloud/tencentcloud/sslclient.go` |
| Annotation | map[string]string | 类型化配置 | `internal/webhookserver/annotationparse.go` |

## 4. Controller 开发模式

新增 Controller 必须遵循以下模式（参考 `portpoolcontroller/portpool_controller.go`）：

1. Struct 包含：`ctx`, `client.Client`, 领域 cache, `record.EventRecorder`
2. 构造函数：`NewXxxReconciler(ctx, cli, cache, eventer)`
3. `SetupWithManager(mgr)`：`For(primaryCRD)` + `Watches(source.Kind)`
4. `Reconcile()`：fetch → `IsNotFound` 优雅处理 → 业务逻辑 → update status
5. 在 `main.go` 通过 `SetupWithManager(mgr)` 注册

## 5. 架构决策记录（ADR）

### 5.1 管理方式

- 存储位置：`docs/adr/`（[索引](../adr/README.md)）
- 命名格式：`NNNN-标题.md`
- 已收录决策：ADR-0001 Namespace Scope 豁免、ADR-0002 HostNetPortPool 端口分配
- 功能设计细节仍放 `specs/{feature-id}/`

## 检查清单

- [x] 分层架构模型已定义
- [x] 至少 3 条自定义规则已制定（共 9 条）
- [x] 错误信息包含修复指引
- [x] 数据边界 Parse 策略已明确
- [x] Controller 开发模式已文档化
- [x] ADR 目录已建立（`docs/adr/`）
