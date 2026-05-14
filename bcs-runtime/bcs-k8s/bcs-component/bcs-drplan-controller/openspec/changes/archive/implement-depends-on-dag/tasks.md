## 1. DAG 工具函数（internal/executor/dag.go）

- [x] 1.1 实现 `buildActionGraph(actions []Action) (map[string][]string, error)`：构建邻接表，检测引用不存在的 action 名
- [x] 1.2 实现 `topoSort(actions []Action, graph map[string][]string) ([][]Action, error)`：Kahn 算法，返回按层分组的 action 列表，检测循环依赖
- [x] 1.3 为 `buildActionGraph` 和 `topoSort` 编写单元测试，覆盖：正常 DAG、循环依赖、引用不存在名称、全无 DependsOn 退化

## 2. ExecuteWorkflow 改造（internal/executor/native_executor.go）

- [x] 2.1 在 `ExecuteWorkflow` 入口调用 `buildActionGraph` + `topoSort` 做校验，校验失败直接返回 Failed 状态
- [x] 2.2 实现 `executeDAGWorkflow`：按拓扑层循环，同层 action 用 goroutine + `sync.WaitGroup` 并发执行，FailFast 时通过 `context.WithCancel` 取消其他 goroutine
- [x] 2.3 同层并发执行结果写入独立 slice，层完成后按原始定义顺序合并到 `status.ActionStatuses`
- [x] 2.4 保留向后兼容路径：所有 action 均无 `DependsOn` 时，走原有 `executeSingleGlobalAction` 串行逻辑（零行为变化）
- [x] 2.5 `when` 条件过滤（`shouldExecuteActionByWhen`）在并发 goroutine 内执行，Skipped 的 action 不阻塞依赖它的后续 action 启动

## 3. 回滚适配（internal/executor/native_executor.go）

- [x] 3.1 确认 `RevertWorkflow` 按 `ActionStatuses` 逆序处理，DAG 并行组内按 status 记录顺序逐个逆序回滚，无需修改主逻辑（验证现有逻辑已满足）
- [x] 3.2 若现有逻辑不满足，补充按层逆序批次处理（不需要，现有逻辑已满足）

## 4. 集成测试

- [x] 4.1 新增 `TestExecuteWorkflow_DAG_Parallel`：两个无依赖 action 并行，验证 ActionStatuses 顺序与定义顺序一致
- [x] 4.2 新增 `TestExecuteWorkflow_DAG_Sequential`：B 依赖 A，验证 B 在 A 完成后启动
- [x] 4.3 新增 `TestExecuteWorkflow_DAG_CycleDetection`：循环依赖，验证 workflow 直接 Failed
- [x] 4.4 新增 `TestExecuteWorkflow_DAG_UnknownRef`：引用不存在 action，验证 workflow 直接 Failed
- [x] 4.5 回归测试：所有无 DependsOn 的现有测试用例通过，行为不变
