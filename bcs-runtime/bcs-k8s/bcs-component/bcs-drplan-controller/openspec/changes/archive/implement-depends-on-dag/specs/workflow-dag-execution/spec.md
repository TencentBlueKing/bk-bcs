## ADDED Requirements

### Requirement: DependsOn 字段合法性校验
执行器 SHALL 在 workflow 开始执行前校验所有 action 的 `DependsOn` 字段：
- 引用的 action 名必须存在于同一 workflow 的 actions 列表中
- action 之间不得存在循环依赖
- 校验失败时 workflow 直接进入 Failed 状态，不执行任何 action

#### Scenario: 引用不存在的 action 名
- **WHEN** workflow 中 action A 的 `DependsOn` 包含 "nonexistent"，而 "nonexistent" 不在 actions 列表中
- **THEN** workflow 状态为 Failed，message 包含 "unknown action in dependsOn"，无任何 action 被执行

#### Scenario: 存在循环依赖
- **WHEN** action A 依赖 B，action B 依赖 A（或更长的环）
- **THEN** workflow 状态为 Failed，message 包含 "cycle detected"，无任何 action 被执行

#### Scenario: 合法的 DependsOn
- **WHEN** action B 的 `DependsOn` 包含 "A"，A 存在且无循环
- **THEN** 校验通过，workflow 正常进入执行阶段

---

### Requirement: 无依赖关系的 action 并行执行
当同一拓扑层内存在多个 action（互相之间无依赖关系）时，执行器 SHALL 并发执行这些 action，不等待其中任一完成后再启动其他。

#### Scenario: 两个无依赖的 post-upgrade action 并行
- **WHEN** workflow 中 post-a 和 post-b 均 `dependsOn: [main]`，且互相之间无依赖
- **THEN** main 完成后，post-a 和 post-b 同时启动执行，两者执行时间重叠

#### Scenario: 三个完全独立的 action 并行
- **WHEN** workflow 中 A、B、C 均无 `DependsOn`
- **THEN** A、B、C 同时启动，执行时间重叠

---

### Requirement: 有依赖的 action 等待所有依赖完成
执行器 SHALL 等待 action 的所有 `DependsOn` 中列出的 action 达到 Succeeded 或 Skipped 状态后，再启动该 action。
若任一依赖 action 为 Failed 且 FailurePolicy=FailFast，则该 action 不再启动。

#### Scenario: 单依赖顺序执行
- **WHEN** action B 的 `DependsOn: [A]`，A 执行完成（Succeeded）
- **THEN** B 在 A 完成后才开始执行

#### Scenario: 多依赖全部完成才启动
- **WHEN** action C 的 `DependsOn: [A, B]`，A 和 B 并行执行
- **THEN** C 等待 A 和 B 均完成（Succeeded）后才启动

#### Scenario: 依赖失败时 FailFast 阻止后续
- **WHEN** action B 依赖 A，A 执行失败，workflow FailurePolicy=FailFast
- **THEN** B 不被执行，状态为 Skipped 或 Failed（标注依赖失败原因）

---

### Requirement: 向后兼容——无 DependsOn 时行为不变
当 workflow 中所有 action 均无 `DependsOn` 字段（或字段为空）时，执行器 SHALL 按 actions 数组顺序逐个顺序执行，行为与引入 DAG 调度前完全一致。

#### Scenario: 全部无 DependsOn 的 workflow
- **WHEN** workflow 中所有 action 的 `dependsOn` 均为空或未设置
- **THEN** action 按定义顺序逐个执行，执行时间不重叠，status.ActionStatuses 顺序与 actions 顺序一致

---

### Requirement: ActionStatus 顺序与原始定义顺序一致
无论 action 实际完成顺序如何，最终 `WorkflowExecutionStatus.ActionStatuses` SHALL 按 `workflow.Spec.Actions` 的原始定义顺序排列。

#### Scenario: 并行 action 完成顺序不同
- **WHEN** action A 和 B 并行执行，B 先于 A 完成
- **THEN** status.ActionStatuses 中 A 的条目仍在 B 之前（按原始定义顺序）

---

### Requirement: DAG 感知的回滚顺序
RevertWorkflow SHALL 按执行记录（ActionStatuses）的逆序处理回滚，DAG 并行组内的 action 以顺序逐个逆序回滚（不并行回滚）。

#### Scenario: 并行组内逆序回滚
- **WHEN** action A 和 B 并行执行均 Succeeded，C 依赖 A、B 后执行
- **THEN** 回滚顺序为：先回滚 C，再按逆序回滚 B、A（顺序逐个，保证回滚的确定性）
