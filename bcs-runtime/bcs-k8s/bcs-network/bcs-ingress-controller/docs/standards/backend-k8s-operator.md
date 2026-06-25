# 后端开发规范（K8s Operator）

<!--
  项目定制规范，基于 controller-runtime v0.6.3 实践编写。
  可贡献回 harness-engineering/assets/standards/ 预设库。
-->

> BCS Ingress Controller 后端开发规范。技术栈：Go 1.20+、controller-runtime v0.6.3、kubebuilder 模式、Prometheus、go-restful。

---

## 一、技术栈要求

| 技术 | 版本要求 | 用途 |
|------|---------|------|
| Go | ≥ 1.20 | 主语言 |
| controller-runtime | v0.6.3 | Operator 框架（Manager / Reconcile / Webhook） |
| k8s client-go | v0.18.6（间接） | K8s API 访问 |
| CRD 类型 | `../../kubernetes/apis/networkextension/v1/` | 独立 Go module，通过 replace 引用 |
| go-restful | v2.15.0 | HTTP 管理 API |
| prometheus/client_golang | v1.16.0 | 指标导出，namespace `bkbcs_ingressctrl` |
| bcs-common/blog | — | 统一日志（**禁止** log/klog） |

**模块路径注意**：`go.mod` 位于 `../`（bcs-network/），构建和测试须从该目录执行。

---

## 二、项目结构

```
bcs-ingress-controller/
├── main.go                     # 入口：注册 Controller / Checker / HTTP / Webhook
├── {name}controller/           # 每个 CRD 一个 Reconcile 控制器目录
├── internal/
│   ├── constant/               # Annotation Key（唯一事实源）
│   ├── option/                 # CLI 参数 → ControllerOption
│   ├── generator/                # Ingress → Listener 转换
│   ├── cloud/                  # 多云 LB 适配（interface + 子包）
│   ├── httpsvr/                # REST 管理 API
│   ├── webhookserver/          # Admission Webhook
│   ├── portpoolcache/          # PortPool 内存缓存
│   ├── hostnetportpoolcache/   # HostNetPortPool 内存缓存
│   ├── metrics/                # Prometheus 指标
│   └── check/                  # 周期性一致性检查
├── specs/                      # 功能设计文档
└── docs/adr/                   # 架构决策记录
```

### 目录职责原则

- `{name}controller/`：仅含 Reconcile 逻辑，不直接引用云 SDK 具体实现
- `internal/cloud/interface.go`：云适配唯一抽象入口
- `internal/constant/constant.go`：所有 Annotation Key 集中定义
- `main.go`：纯编排，云初始化逻辑放 `initXxxClient`，`initClient` 保持纯分发（复杂度 ≤ 5）

---

## 三、Controller 开发规范

### 3.1 Reconciler 结构模板

参考 `portpoolcontroller/portpoolcontroller.go`：

```go
type XxxReconciler struct {
    ctx       context.Context
    k8sClient client.Client
    eventer   record.EventRecorder
    // 领域依赖：cache、cloud client、opts 等
}

func NewXxxReconciler(ctx context.Context, cli client.Client, eventer record.EventRecorder, ...) *XxxReconciler

func (r *XxxReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error)
func (r *XxxReconciler) SetupWithManager(mgr ctrl.Manager) error
```

### 3.2 Reconcile 必须遵守

| 规则 | 说明 |
|------|------|
| 幂等性 | 多次 Reconcile 产生相同结果，避免无条件 Update |
| IsNotFound 优雅处理 | `client.Get` 返回 NotFound 时返回 `ctrl.Result{}` 而非 error |
| 错误时 Requeue | 临时错误返回 `RequeueAfter`，永久错误记录 Event 并停止重试 |
| 先 Get 再比较 | Status 更新前比较 `reflect.DeepEqual` 或字段级 diff，减少 etcd 写入 |
| 注册到 main.go | 新建 Controller 必须在 `main.go` 调用 `SetupWithManager(mgr)` |

### 3.3 SetupWithManager 模式

```go
ctrl.NewControllerManagedBy(mgr).
    For(&networkextensionv1.PortPool{}).
    WithEventFilter(predicate.Funcs{...}).  // 按需
    Complete(r)
```

跨资源 Watch 使用 `Watches(source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{...})`。

### 3.4 命名与注释约束

- **导出函数**（首字母大写）**必须**有 GoDoc 注释，注释语言为英文，以被导出名称开头（符合 Go 惯例）
- 所有函数名不得超过 **35 字符**，**包括测试函数**（`func TestXxx`）；超长时使用缩写（如 `Alloc`/`Contig`/`Seg`）
- 圈复杂度 > 10 的函数必须拆分
- 包注释说明 package 职责（英文）

---

## 四、分层架构

### 4.1 依赖方向

```
internal/constant + internal/option（最底层）
  ↓
internal/generator / internal/*cache / internal/cloud
  ↓
{name}controller/（Reconcile 层）
  ↓
internal/httpsvr + internal/webhookserver（接入层）
  ↓
main.go（编排层）
```

| 层 | 允许依赖 | 禁止依赖 |
|----|---------|---------|
| constant/option | 标准库 | 其他 internal 包 |
| cloud 子包 | constant、interface | 其他云子包、controller |
| controller | internal/*、CRD types | 反向依赖上层 |
| httpsvr/webhookserver | internal/*、go-restful | controller 包 |

### 4.2 云适配器规范

- 新增云厂商：在 `internal/cloud/{vendor}/` 实现 `cloud.LoadBalance` 接口
- `main.go` 添加 `initXxxClient`，在 `IsNamespaceScope=true` 时调用 `newNamespacedLBWithExempt`
- **禁止**在 `initClient` 中直接写业务逻辑

---

## 五、Webhook 规范

| 规则 | 说明 |
|------|------|
| 入口 | `internal/webhookserver/webhookserver.go` |
| Annotation 解析 | 统一走 `annotationparse.go`，Key 来自 `constant.go` |
| 验证与变更分离 | `validate.go` / `mutating.go` 按资源类型拆分 |
| 指标 | 在 `internal/metrics/webhook.go` 注册并上报 |

---

## 六、Metrics 规范

### 6.1 注册模式

每个子系统在 `internal/metrics/{subsystem}.go` 中：

```go
var myMetric = prometheus.NewGaugeVec(...)

func init() {
    metrics.Registry.MustRegister(myMetric)
}

func ReportXxx(...) { ... }  // 导出 helper
func CleanXxx(...)  { ... }  // 资源删除时清理
```

### 6.2 命名空间

所有指标 namespace 为 `bkbcs_ingressctrl`，subsystem 按模块划分（`portpool`、`api`、`controller` 等）。

---

## 七、缓存规范

| 缓存 | 目录 | 冷启动 |
|------|------|--------|
| PortPool | `internal/portpoolcache/` | `RebuildFromAPIServer()` |
| HostNetPortPool | `internal/hostnetportpoolcache/` | 同上 |
| Ingress 关联 | `internal/ingresscache/` | Leader 选举后重建 |
| Node 元数据 | `internal/nodecache/` | Controller Reconcile 填充 |

- 使用 `sync.RWMutex` 保证线程安全
- 类型定义放独立 `types.go`

---

## 八、编码规范

| 规则 | 说明 |
|------|------|
| 日志 | 仅用 `blog`，禁止 `log`/`klog` |
| 常量 | Annotation Key 禁止硬编码字符串 |
| 错误处理 | 所有 `client.Get/Update/List` 必须检查 error |
| 格式 | gofmt / goimports |
| 注释 | 代码注释英文；文档/PR 中文 |
| 导出注释 | 导出类型和函数必须有 GoDoc 注释（英文） |

---

## 九、测试规范

### 9.1 模式

- **表驱动测试**：测试用例以 `[]struct{ name string; ... }` 组织
- **Fake Client**：使用 `controller-runtime/pkg/client/fake` 构造测试环境
- **测试文件**：与源码同目录，命名 `*_test.go`
- **测试函数命名**：`func TestXxx` 名称不得超过 35 字符（与 §3.4 一致）

### 9.2 构建命令

```bash
cd .. && make test-ingress-controller          # 全量测试 + 覆盖率
go test -v -run TestReconcile ./hostnetportcontroller/...  # 单包
```

### 9.3 特性回归（Namespace Scope 等）

```bash
go test -count=1 -run 'TestParseExemptNamespaces|TestRuleConverter|TestMappingConverter|TestIsExempt|TestGetNsClient|TestReloadNsClient|TestNewNamespacedLB' \
  ./internal/cloud/namespacedlb/... ./internal/generator/... .
```

---

## 十、配置管理

| 配置来源 | 位置 | 说明 |
|---------|------|------|
| CLI 参数 | `internal/option/option.go` | 通过 flag 解析为 `ControllerOption` |
| Namespace Scope | `--is_namespace_scope` | 限制 Ingress 引用同 NS Service |
| 豁免命名空间 | `--namespace_scope_exempt_namespaces` | 逗号分隔，见 ADR-0001 |
| 云凭证（全局） | 环境变量 `TENCENTCLOUD_ACCESS_KEY_ID` / `TENCENTCLOUD_ACESS_KEY` | 注意 ACESS 拼写 |
| 云凭证（per-ns） | Secret `ingress-secret.networkextension.bkbcs.tencent.com` | 非豁免 NS 使用 |

---

## 十一、构建与部署

```bash
cd .. && make ingress-controller              # 构建二进制
kubectl rollout restart -n bcs-system deployment/bcsingresscontroller
```

CRD 类型变更时，须在 `../../kubernetes/` 重新生成 deepcopy 和 manifest。

---

## 十二、新增功能检查清单

- [ ] 新常量已加入 `internal/constant/constant.go`
- [ ] 新 Controller 已在 `main.go` 注册 `SetupWithManager`
- [ ] 新 metrics 已在 `internal/metrics/` 通过 `init()` 注册
- [ ] 新 HTTP 路由已在 `httpserver.go` `InitRouters()` 注册
- [ ] 单元测试表驱动，fake client
- [ ] gofmt / goimports 干净
- [ ] 导出函数有 GoDoc 注释（英文）
- [ ] 函数名 ≤ 35 字符（含测试函数），复杂度 ≤ 10
- [ ] 关联 ADR / specs 文档已更新
