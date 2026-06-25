# 模块索引（Module Index）

> 按模块列出职责和关联文件，用于评估变更的内部影响范围。

<!-- dev-map:auto -->
## 模块清单

| 模块 | 职责 | 文件数 |
|------|------|-------|
| [entry](#entry) | 程序入口与全局编排 | 2 |
| [ingress-controller](#ingress-controller) | Ingress CRD Reconcile | 6 |
| [listener-controller](#listener-controller) | Listener CRD Reconcile | 3 |
| [portpool-controller](#portpool-controller) | PortPool CRD Reconcile | 4 |
| [portbinding-controller](#portbinding-controller) | PortBinding CRD Reconcile | 10 |
| [hostnetport-controller](#hostnetport-controller) | HostNetPortPool CRD Reconcile | 8 |
| [namespace-controller](#namespace-controller) | Namespace 变更监听 | 2 |
| [node-controller](#node-controller) | Node 元数据缓存 | 1 |
| [generator](#generator) | Ingress → Listener 转换 | 8 |
| [cloud-adapters](#cloud-adapters) | 多云 LB SDK 适配 | 50 |
| [httpsvr](#httpsvr) | REST 管理 API | 10 |
| [webhookserver](#webhookserver) | Admission Webhook | 15 |
| [caches](#caches) | 内存缓存（PortPool/HostNet/Ingress/Node） | 14 |
| [check](#check) | 周期性一致性检查与证书过期巡检 | 14 |
| [metrics](#metrics) | Prometheus 指标 | 9 |
| [foundation](#foundation) | 常量、配置、工具、公共库 | 21 |
| [bcs-ingress-inspector](#bcs-ingress-inspector) | 独立诊断二进制 | 3 |
| [cli-util](#cli-util) | 独立 CLI 工具 | 3 |

---

## entry

**职责**：程序入口，注册所有 Controller、Checker、HTTP、Webhook，初始化云客户端。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`main.go`](../../main.go) | 代码 | 全局编排入口 |
| [`main_test.go`](../../main_test.go) | 测试 | 入口层单元测试 |

---

## ingress-controller

**职责**：监听 Ingress CRD 变更，触发 Listener 生成与云资源同步。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`ingresscontroller/ingress_controller.go`](../../ingresscontroller/ingress_controller.go) | 代码 | Reconcile 主逻辑 |
| [`ingresscontroller/endpointfilter.go`](../../ingresscontroller/endpointfilter.go) | 代码 | Endpoint 过滤 |
| [`ingresscontroller/multiclusereps_filter.go`](../../ingresscontroller/multiclusereps_filter.go) | 代码 | 多集群副本过滤 |
| [`ingresscontroller/podfilter.go`](../../ingresscontroller/podfilter.go) | 代码 | Pod 过滤 |
| [`ingresscontroller/utils.go`](../../ingresscontroller/utils.go) | 代码 | 工具函数 |
| [`ingresscontroller/utils_test.go`](../../ingresscontroller/utils_test.go) | 测试 | 工具函数测试 |

---

## listener-controller

**职责**：管理 Listener CRD 生命周期，对接云 LB 创建/更新/删除。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`listenercontroller/listener_controller.go`](../../listenercontroller/listener_controller.go) | 代码 | Reconcile 主逻辑 |
| [`listenercontroller/listener_bypass_controller.go`](../../listenercontroller/listener_bypass_controller.go) | 代码 | Bypass 模式控制器 |
| [`listenercontroller/listenerHelper.go`](../../listenercontroller/listenerHelper.go) | 代码 | 辅助函数 |

---

## portpool-controller

**职责**：管理 PortPool CRD，维护节点端口池分配状态。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`portpoolcontroller/portpoolcontroller.go`](../../portpoolcontroller/portpoolcontroller.go) | 代码 | Reconcile 主逻辑 |
| [`portpoolcontroller/portpool.go`](../../portpoolcontroller/portpool.go) | 代码 | PortPool 领域逻辑 |
| [`portpoolcontroller/portpoolitem.go`](../../portpoolcontroller/portpoolitem.go) | 代码 | 端口项管理 |
| [`portpoolcontroller/util.go`](../../portpoolcontroller/util.go) | 代码 | 工具函数 |

---

## portbinding-controller

**职责**：管理 PortBinding CRD，关联 Pod 与宿主机端口。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`portbindingcontroller/portbindingcontroller.go`](../../portbindingcontroller/portbindingcontroller.go) | 代码 | Reconcile 主逻辑 |
| [`portbindingcontroller/portbinding.go`](../../portbindingcontroller/portbinding.go) | 代码 | PortBinding 领域逻辑 |
| [`portbindingcontroller/portbindingitem.go`](../../portbindingcontroller/portbindingitem.go) | 代码 | 绑定项管理 |
| [`portbindingcontroller/portbindingbypasscontroller.go`](../../portbindingcontroller/portbindingbypasscontroller.go) | 代码 | Bypass 控制器 |
| [`portbindingcontroller/podportbindinghandler.go`](../../portbindingcontroller/podportbindinghandler.go) | 代码 | Pod 端口绑定处理 |
| [`portbindingcontroller/nodeportbindinghandler.go`](../../portbindingcontroller/nodeportbindinghandler.go) | 代码 | Node 端口绑定处理 |
| [`portbindingcontroller/podfilter.go`](../../portbindingcontroller/podfilter.go) | 代码 | Pod 过滤 |
| [`portbindingcontroller/nodefilter.go`](../../portbindingcontroller/nodefilter.go) | 代码 | Node 过滤 |
| [`portbindingcontroller/event.go`](../../portbindingcontroller/event.go) | 代码 | 事件处理 |
| [`portbindingcontroller/util.go`](../../portbindingcontroller/util.go) | 代码 | 工具函数 |

---

## hostnetport-controller

**职责**：HostNetPortPool 动态端口分配，管理 hostNetwork Pod 端口。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`hostnetportcontroller/pool_controller.go`](../../hostnetportcontroller/pool_controller.go) | 代码 | Pool Reconcile |
| [`hostnetportcontroller/pod_controller.go`](../../hostnetportcontroller/pod_controller.go) | 代码 | Pod Reconcile |
| [`hostnetportcontroller/node_controller.go`](../../hostnetportcontroller/node_controller.go) | 代码 | Node Reconcile |
| [`hostnetportcontroller/podfilter.go`](../../hostnetportcontroller/podfilter.go) | 代码 | Pod 过滤 |
| [`hostnetportcontroller/nodefilter.go`](../../hostnetportcontroller/nodefilter.go) | 代码 | Node 过滤 |
| [`hostnetportcontroller/controller_test.go`](../../hostnetportcontroller/controller_test.go) | 测试 | 控制器测试 |
| [`hostnetportcontroller/podfilter_test.go`](../../hostnetportcontroller/podfilter_test.go) | 测试 | Pod 过滤测试 |
| [`hostnetportcontroller/nodefilter_test.go`](../../hostnetportcontroller/nodefilter_test.go) | 测试 | Node 过滤测试 |

---

## namespace-controller

**职责**：监听 Namespace 变更，触发云客户端重载。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`namespacecontroller/namespacecontroller.go`](../../namespacecontroller/namespacecontroller.go) | 代码 | Reconcile 主逻辑 |
| [`namespacecontroller/utils.go`](../../namespacecontroller/utils.go) | 代码 | 工具函数 |

---

## node-controller

**职责**：缓存 Node 元数据供其他 Controller 使用。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`nodecontroller/nodecontroller.go`](../../nodecontroller/nodecontroller.go) | 代码 | Node 缓存控制器 |

---

## generator

**职责**：将 Ingress Spec 转换为 Listener CRD（L7 规则 / L4 端口映射）。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/generator/ingressconverter.go`](../../internal/generator/ingressconverter.go) | 代码 | 转换入口 |
| [`internal/generator/listenerconverter.go`](../../internal/generator/listenerconverter.go) | 代码 | Listener 转换 |
| [`internal/generator/ruleconverter.go`](../../internal/generator/ruleconverter.go) | 代码 | L7 规则转换 |
| [`internal/generator/mappingconverter.go`](../../internal/generator/mappingconverter.go) | 代码 | L4 映射转换 |
| [`internal/generator/util.go`](../../internal/generator/util.go) | 代码 | 工具函数 |
| [`internal/generator/ingressconverter_test.go`](../../internal/generator/ingressconverter_test.go) | 测试 | 转换测试 |
| [`internal/generator/namespace_scope_exempt_test.go`](../../internal/generator/namespace_scope_exempt_test.go) | 测试 | 豁免特性测试 |
| [`internal/generator/util_test.go`](../../internal/generator/util_test.go) | 测试 | 工具测试 |

---

## cloud-adapters

**职责**：封装 AWS/Azure/GCP/腾讯云 LB SDK，提供统一接口与 Namespace 隔离客户端。

| 文件 | 类型 | 说明 |
|------|------|------|
| `internal/cloud/interface.go` + `mock/` | 代码 | 云接口定义与 Mock |
| `internal/cloud/aws/*.go`（11 文件） | 代码 | AWS ELB/AGA 适配 |
| `internal/cloud/azure/*.go`（10 文件） | 代码 | Azure LB 适配 |
| `internal/cloud/gcp/*.go`（7 文件） | 代码 | GCP CLB 适配 |
| `internal/cloud/tencentcloud/*.go`（14 文件） | 代码 | 腾讯云 CLB/SSL 适配 |
| `internal/cloud/namespacedlb/*.go`（2 文件） | 代码 | Namespace 隔离 LB 客户端 |
| `internal/cloud/namespacedssl/*.go`（2 文件） | 代码 | Namespace 隔离 SSL 证书客户端 |
| `internal/cloudcollector/*.go`（3 文件） | 代码 | 云资源状态采集 |
| `internal/cloudnode/*.go`（2 文件） | 代码 | 云节点客户端 |

---

## httpsvr

**职责**：go-restful HTTP 管理 API，提供 PortPool/Listener/Ingress 等查询接口。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/httpsvr/httpserver.go`](../../internal/httpsvr/httpserver.go) | 代码 | 路由注册入口 |
| [`internal/httpsvr/portpool.go`](../../internal/httpsvr/portpool.go) | 代码 | PortPool API |
| [`internal/httpsvr/listener.go`](../../internal/httpsvr/listener.go) | 代码 | Listener API |
| [`internal/httpsvr/ingress.go`](../../internal/httpsvr/ingress.go) | 代码 | Ingress API |
| [`internal/httpsvr/hostnetportpool.go`](../../internal/httpsvr/hostnetportpool.go) | 代码 | HostNetPortPool API |
| [`internal/httpsvr/node.go`](../../internal/httpsvr/node.go) | 代码 | Node API |
| [`internal/httpsvr/response.go`](../../internal/httpsvr/response.go) | 代码 | 响应封装 |
| [`internal/httpsvr/readiness_probe.go`](../../internal/httpsvr/readiness_probe.go) | 代码 | 健康检查 |
| [`internal/httpsvr/check_bind_status.go`](../../internal/httpsvr/check_bind_status.go) | 代码 | 绑定状态检查 |
| [`internal/httpsvr/aga_support.go`](../../internal/httpsvr/aga_support.go) | 代码 | AGA 支持 API |

---

## webhookserver

**职责**：K8s Admission Webhook，验证和变更 Ingress/PortPool/Node 等资源。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/webhookserver/webhookserver.go`](../../internal/webhookserver/webhookserver.go) | 代码 | Webhook 服务入口 |
| [`internal/webhookserver/validate.go`](../../internal/webhookserver/validate.go) | 代码 | 验证逻辑 |
| [`internal/webhookserver/mutating.go`](../../internal/webhookserver/mutating.go) | 代码 | 变更逻辑 |
| [`internal/webhookserver/mutating_ingress.go`](../../internal/webhookserver/mutating_ingress.go) | 代码 | Ingress 变更 |
| [`internal/webhookserver/mutating_node.go`](../../internal/webhookserver/mutating_node.go) | 代码 | Node 变更 |
| [`internal/webhookserver/portallocate.go`](../../internal/webhookserver/portallocate.go) | 代码 | 端口分配 |
| [`internal/webhookserver/annotationparse.go`](../../internal/webhookserver/annotationparse.go) | 代码 | Annotation 解析 |
| 其余 8 个 .go 文件 | 代码/测试 | 转换、校验、工具 |

---

## caches

**职责**：PortPool、HostNetPortPool、Ingress、Node 的内存缓存，支持冷启动重建。

| 文件 | 类型 | 说明 |
|------|------|------|
| `internal/portpoolcache/*.go`（7 文件） | 代码/测试 | PortPool 缓存 |
| `internal/hostnetportpoolcache/*.go`（3 文件） | 代码/测试 | HostNetPortPool 缓存 |
| `internal/ingresscache/*.go`（3 文件） | 代码 | Ingress 关联缓存 |
| `internal/nodecache/nodecache.go` | 代码 | Node 元数据缓存 |

---

## check

**职责**：周期性一致性检查（Listener、PortBinding、PortLeak、LB Cache 等）及 SSL 证书过期巡检。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/check/checkrunner.go`](../../internal/check/checkrunner.go) | 代码 | 检查调度器 |
| [`internal/check/binding.go`](../../internal/check/binding.go) | 代码 | Ingress 证书 Binding 展开 |
| [`internal/check/certificatechecker.go`](../../internal/check/certificatechecker.go) | 代码 | SSL 证书过期检查 |
| [`internal/check/listenerchecker.go`](../../internal/check/listenerchecker.go) | 代码 | Listener 一致性 |
| [`internal/check/portbindchecker.go`](../../internal/check/portbindchecker.go) | 代码 | PortBinding 一致性 |
| [`internal/check/port_leak_checker.go`](../../internal/check/port_leak_checker.go) | 代码 | 端口泄漏检查 |
| [`internal/check/hostnet_segment_checker.go`](../../internal/check/hostnet_segment_checker.go) | 代码 | HostNet 网段检查 |
| [`internal/check/lbcachechecker.go`](../../internal/check/lbcachechecker.go) | 代码 | LB 缓存检查 |
| 其余 6 个文件 | 代码/测试 | 接口定义与测试 |

---

## metrics

**职责**：Prometheus 指标注册与上报，namespace `bkbcs_ingressctrl`。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/metrics/metric.go`](../../internal/metrics/metric.go) | 代码 | 注册中心 |
| [`internal/metrics/certificate.go`](../../internal/metrics/certificate.go) | 代码 | SSL 证书过期指标 |
| [`internal/metrics/portpool.go`](../../internal/metrics/portpool.go) | 代码 | PortPool 指标 |
| [`internal/metrics/hostnetportpool.go`](../../internal/metrics/hostnetportpool.go) | 代码 | HostNetPortPool 指标 |
| [`internal/metrics/listener_controller.go`](../../internal/metrics/listener_controller.go) | 代码 | Listener 指标 |
| [`internal/metrics/webhook.go`](../../internal/metrics/webhook.go) | 代码 | Webhook 指标 |
| [`internal/metrics/check.go`](../../internal/metrics/check.go) | 代码 | Check 指标 |
| [`internal/metrics/nodeinfoexporter.go`](../../internal/metrics/nodeinfoexporter.go) | 代码 | Node 信息导出 |
| 其余 1 个文件 | 测试 | certificate 指标测试 |

---

## foundation

**职责**：常量、CLI 配置、工具函数、事件、冲突处理、Worker 同步等基础设施。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`internal/constant/constant.go`](../../internal/constant/constant.go) | 代码 | 共享常量 |
| [`internal/constant/webhook.go`](../../internal/constant/webhook.go) | 代码 | Webhook 常量 |
| [`internal/option/option.go`](../../internal/option/option.go) | 代码 | CLI 参数 |
| [`internal/option/option_test.go`](../../internal/option/option_test.go) | 测试 | CLI 参数测试 |
| `internal/utils/*.go`（2 文件） | 代码 | 通用工具 |
| `internal/apiclient/*.go`（7 文件） | 代码 | 外部 API 客户端 |
| `internal/worker/*.go`（8 文件） | 代码/测试 | 同步 Worker |
| `internal/conflicthandler/*.go`（4 文件） | 代码/测试 | 资源冲突处理 |
| `internal/eventer/eventer.go` | 代码 | 事件发送 |
| `internal/common/common.go` | 代码 | 公共定义 |

---

## bcs-ingress-inspector

**职责**：独立诊断二进制，用于 PortBinding 等资源的离线诊断。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`bcs-ingress-inspector/main.go`](../../bcs-ingress-inspector/main.go) | 代码 | 诊断工具入口 |
| [`bcs-ingress-inspector/option/option.go`](../../bcs-ingress-inspector/option/option.go) | 代码 | CLI 参数 |
| [`bcs-ingress-inspector/portbindingcontroller/portbindingcontroller.go`](../../bcs-ingress-inspector/portbindingcontroller/portbindingcontroller.go) | 代码 | PortBinding 诊断 |

---

## cli-util

**职责**：独立 CLI 工具集，不纳入主 Controller 二进制。

| 文件 | 类型 | 说明 |
|------|------|------|
| [`cli-util/validate-listener-name/main.go`](../../cli-util/validate-listener-name/main.go) | 代码 | Listener 名称校验入口 |
| [`cli-util/validate-listener-name/pkg/handler.go`](../../cli-util/validate-listener-name/pkg/handler.go) | 代码 | 校验逻辑 |
| [`cli-util/validate-listener-name/pkg/options.go`](../../cli-util/validate-listener-name/pkg/options.go) | 代码 | CLI 参数 |

<!-- /dev-map:auto -->

---

*由 harness-generating 扫描生成，最后更新：2026-06-11（gardening 增量同步证书过期特性）*
