# Tasks: Per-Cluster Fan-out Execution

**Input**: Design documents from `openspec/changes/2026-03-24-percluster-fanout/`
**Prerequisites**: proposal.md, design.md, specs/percluster-fanout.md

---

## Phase 1: API Layer（基础设施）

**Purpose**: 新增 API 字段和类型，生成 CRD

- [x] T001 Add `ClusterExecutionMode` field to `Action` struct in `api/v1alpha1/common_types.go`
- [x] T002 Add `ClusterActionStatus` type and extend `ActionStatus` with `ClusterStatuses` in `api/v1alpha1/common_types.go`
- [x] T003 Add `ClusterExecutionModeGlobal` and `ClusterExecutionModePerCluster` constants in `api/v1alpha1/constants.go`
- [x] T004 Run `make manifests generate` to regenerate CRD and DeepCopy

**Checkpoint**: API 层变更完成，CRD 正确生成，`make test` 通过（无行为变更）

---

## Phase 2: Subscription Executor Per-Cluster 拆分（核心能力）

**Purpose**: 实现运行时子 Subscription 生成与 per-cluster readiness 检查

- [x] T005 [TDD] Add unit tests for `isPerClusterMode()` helper in `internal/executor/percluster_test.go`
- [x] T006 Implement `isPerClusterMode(action)` helper function in `internal/executor/percluster.go`
- [x] T007 [TDD] Add unit tests for `createChildSubscription()` — verify name, namespace, single-cluster subscriber, OwnerReference
- [x] T008 Implement `buildChildSubscription()` in `internal/executor/percluster.go`
- [x] T009 [TDD] Add unit tests for `resolveBindingClusters()` — verify parsing of `status.bindingClusters`
- [x] T010 Implement `resolveBindingClusters()` in `internal/executor/percluster.go`
- [x] T011 [TDD] Add unit tests for `ExecutePerCluster()` — verify per-cluster child Subscription creation and waitReady
- [x] T012 Implement `ExecutePerCluster(ctx, action, clusterBinding, params)` in `internal/executor/percluster.go`
- [x] T013 Modify `Execute()` to detect PerCluster mode and skip aggregate waitReady (return status for workflow executor to handle)

**代码审查**: Phase 2 完成后执行代码审查

**Checkpoint**: Subscription executor 支持 per-cluster 子 Subscription 创建和 readiness 检查，`make test` 通过

---

## Phase 3: Workflow Executor 调度改造（编排能力）

**Purpose**: 实现 batch 分组和并发 per-cluster 执行

- [x] T014 [TDD] Add unit tests for `groupActionBatches()` — verify correct batch grouping for mixed Global/PerCluster actions
- [x] T015 Implement `groupActionBatches(actions)` in `internal/executor/batch.go`
- [x] T016 [TDD] Add unit tests for `aggregateClusterStatuses()` — verify aggregation rules (all succeeded, any failed, etc.)
- [x] T017 Implement `aggregateClusterStatuses()` in `internal/executor/batch.go`
- [x] T018 [TDD] Add unit tests for `executePerClusterBatch()` — verify concurrent execution per cluster, barrier behavior
- [x] T019 Implement `executePerClusterBatch()` with goroutine-based concurrent per-cluster execution in `internal/executor/native_executor.go`
- [x] T020 Refactor `ExecuteWorkflow()` to use batch-based execution model, maintaining backward compatibility for all-Global workflows

**代码审查**: Phase 3 完成后执行代码审查

**Checkpoint**: Workflow executor 支持 PerCluster batch 并发执行，Global/PerCluster 混合 workflow 正确执行，`make test` 通过

---

## Phase 4: drplan-gen 适配

**Purpose**: drplan-gen 输出 `clusterExecutionMode` 字段

- [x] T021 Update `internal/generator/` to set `clusterExecutionMode: PerCluster` for hook Subscription actions
- [x] T022 Update golden files (`testdata/output/`) to reflect new `clusterExecutionMode` field
- [x] T023 Update generator unit tests to assert `clusterExecutionMode` values

**Checkpoint**: drplan-gen 输出正确的 `clusterExecutionMode`，`make test` 通过

---

## Phase 5: Revert/Rollback 支持

**Purpose**: 确保 PerCluster action 的回滚正确清理子 Subscription

- [x] T024 [TDD] Add unit tests for PerCluster action rollback — verify parent Subscription deletion triggers child cleanup via OwnerReference
- [x] T025 Verified: existing `SubscriptionActionExecutor.Rollback()` already handles PerCluster (deletes parent, K8s cascades children)
- [x] T026 Verified: existing `RevertWorkflow()` operates per-action by name, agnostic to batch model

**代码审查**: Phase 5 完成后执行代码审查

**Checkpoint**: PerCluster action 的 rollback 正确清理所有子 Subscription，`make test` 通过

---

## Phase 6: Documentation Sync

- [ ] T027 Update `docs/waitready-design.md` with per-cluster fan-out execution design
- [ ] T028 Update `docs/drplan-gen-guide.md` with `clusterExecutionMode` usage
- [ ] T029 Update `docs/user-guide.md` with per-cluster execution mode examples
- [ ] T030 Sync OpenSpec proposal/design documents if any decisions changed during implementation

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (API)**: No dependencies — start immediately
- **Phase 2 (Subscription Executor)**: Depends on Phase 1
- **Phase 3 (Workflow Executor)**: Depends on Phase 2
- **Phase 4 (drplan-gen)**: Depends on Phase 1 only (can parallel with Phase 2/3)
- **Phase 5 (Revert)**: Depends on Phase 2 + Phase 3
- **Phase 6 (Docs)**: Depends on all above

### Parallel Opportunities

- Phase 4 (drplan-gen) 可与 Phase 2/3 并行
- Phase 2 内的 T005/T006 和 T007/T008 可并行（不同函数，不同文件区域）

### 每个 Phase 完成后的代码审查

- Phase 2 完成后（T005-T013）
- Phase 3 完成后（T014-T020）
- Phase 5 完成后（T024-T026）
