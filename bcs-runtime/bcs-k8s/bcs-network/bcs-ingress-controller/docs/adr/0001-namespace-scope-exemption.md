# 0001. Namespace Scope 豁免机制

## 状态：已接受

## 背景

BCS Ingress Controller 支持 `--is_namespace_scope` 模式，限制每个 Ingress 只能引用同 Namespace 的 Service，且非豁免 Namespace 须通过 per-namespace Secret 提供云凭证。

在联邦集群场景下，部分 Namespace（如平台管理 NS）需要：
1. 跨 Namespace 引用 Service（绑定其他 NS 的后端）
2. 使用 Controller 全局云凭证，而非每个 NS 单独配置 Secret

## 决策

引入 **Namespace Scope Exemption** 机制：

- CLI 参数：`--namespace_scope_exempt_namespaces`（逗号分隔 Namespace 列表）
- 解析：`parseExemptNamespaces()` → `map[string]struct{}`
- 豁免效果（同时生效）：
  - **跨 NS 绑定**：Generator 层跳过 cross-namespace 限制
  - **全局云凭证**：NamespacedLB 对豁免 NS 使用 `defaultClient`，其他 NS 走 per-ns Secret

### 数据流

```
ControllerOption.NamespaceScopeExemptNamespaces (string)
  └─ parseExemptNamespaces() → map[string]struct{}
       ├─ IngressConverterOpt.ExemptNamespaces (generator/ingressconverter.go)
       │    └─ RuleConverter / MappingConverter.isIngressNamespaceExempt()
       └─ newNamespacedLBWithExempt() → NamespacedLB (namespacedlb/namespacedclient.go)
            └─ getNsClient(): exempt ns → defaultClient, others → per-ns secret
```

### 新增云厂商接入

1. 在 `main.go` 添加 `initXxxClient`，参照 `initTencentCloudClient`
2. `IsNamespaceScope=true` 时调用 `newNamespacedLBWithExempt`
3. 在 `initClient` switch 中增加 case

## 后果

**正面：**
- 联邦集群平台 NS 无需为每个 NS 复制云 Secret
- 跨 NS Service 引用与凭证策略统一由白名单控制

**负面：**
- 豁免 NS 使用全局凭证，须严格控制白名单范围
- 数据流横跨 generator + cloud 两层，变更需同步多处

## 关联文件

| 层级 | 文件 |
|------|------|
| 入口 | `main.go`（parseExemptNamespaces / initXxxClient） |
| Generator | `internal/generator/{ingress,rule,mapping}converter.go` |
| Cloud | `internal/cloud/namespacedlb/namespacedclient.go` |
| 测试 | `main_test.go`、`namespace_scope_exempt_test.go`、`namespacedclient_test.go` |

## Secret Key 注意

per-ns Secret 中 key 名为 `TENCENTCLOUD_ACESS_KEY`（历史拼写，少一个 C），与代码常量一致。
