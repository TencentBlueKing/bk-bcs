# 0002. HostNetPortPool 动态端口分配

## 状态：已接受

## 背景

使用 `hostNetwork: true` 的 Pod 需要绑定宿主机端口，多个 Pod 可能竞争同一端口段。需要一种机制在集群内动态分配端口，避免冲突，并在 Pod 生命周期变化时自动回收。

## 决策

引入 **HostNetPortPool** CRD 及配套 Controller + 内存缓存：

1. **CRD**：`networkextension.HostNetPortPool` 定义端口池范围与分配策略
2. **Controller**（`hostnetportcontroller/`）：
   - `pool_controller.go`：Pool CRD Reconcile
   - `pod_controller.go`：Pod 变更触发端口分配/回收
   - `node_controller.go`：Node 变更联动
3. **缓存**：`internal/hostnetportpoolcache/` 维护分配状态，冷启动从 API Server 重建
4. **Webhook**：`internal/webhookserver/portallocate.go` 在 Pod 创建时注入端口 Annotation
5. **HTTP API**：`GET /api/v1/hostnetportpool/bindingresult` 查询绑定结果
6. **Metrics**：`internal/metrics/hostnetportpool.go`

### Annotation 约定

- 使用 `internal/constant/constant.go` 中定义的 Key（禁止硬编码）
- 绑定状态通过 Annotation 回写 Pod

## 后果

**正面：**
- hostNetwork Pod 端口自动分配，减少人工配置
- 缓存 + Reconcile 保证 Leader 切换后可恢复

**负面：**
- 增加 Controller 数量与 Watch 开销
- 端口池配置错误可能导致分配失败或泄漏（需 `internal/check/hostnet_segment_checker.go` 巡检）

## 关联文档

- 功能设计：`specs/001-hostnet-port-allocation/`
- 开发地图模块：`docs/dev-map/module-index.md` → hostnetport-controller
