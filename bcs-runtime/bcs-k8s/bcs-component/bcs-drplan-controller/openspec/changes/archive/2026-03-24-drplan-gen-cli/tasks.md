## 1. 数据结构与 YAML 解析器

- [x] 1.1 定义内部数据结构（types.go）  <!-- 非 TDD 任务 -->
  - [x] 1.1.1 执行变更：`internal/generator/types.go`（ChartAnalysis、HookResource、MainResource 等结构体）
  - [x] 1.1.2 验证无回归（运行：`go build ./...`，确认编译通过）
  - [x] 1.1.3 检查：确认结构体字段覆盖所有 Helm Hook annotation 信息

- [x] 1.2 实现 YAML 多文档解析器  <!-- TDD 任务 -->
  - [x] 1.2.1 写失败测试：`internal/generator/parser_test.go`（测试多文档分割、空文档跳过、无效 YAML 报错）
  - [x] 1.2.2 验证测试失败（运行：`go test ./internal/generator/ -run TestParser -v`，确认失败原因是缺少功能）
  - [x] 1.2.3 写最小实现：`internal/generator/parser.go`（ParseYAML 函数，按 `---` 分割 YAML 文档，反序列化为 []unstructured.Unstructured）
  - [x] 1.2.4 验证测试通过（运行：`go test ./internal/generator/ -run TestParser -v`，确认所有测试通过）
  - [x] 1.2.5 重构：整理代码、改善命名、消除重复（保持测试通过）

- [x] 1.3 代码审查（跳过：superpowers 工具不可用，人工确认通过）

## 2. 资源分类器

- [x] 2.1 实现资源分类器  <!-- TDD 任务 -->
  - [x] 2.1.1 写失败测试：`internal/generator/classifier_test.go`（测试 Hook 识别、hook-weight 排序、hook-delete-policy 映射、主资源归类、无 hook 场景、混合场景）
  - [x] 2.1.2 验证测试失败（运行：`go test ./internal/generator/ -run TestClassifier -v`，确认失败原因是缺少功能）
  - [x] 2.1.3 写最小实现：`internal/generator/classifier.go`（Classify 函数，遍历 []Unstructured 按 annotation 分类为 ChartAnalysis）
  - [x] 2.1.4 验证测试通过（运行：`go test ./internal/generator/ -run TestClassifier -v`，确认所有测试通过）
  - [x] 2.1.5 重构：整理代码、改善命名、消除重复（保持测试通过）

- [x] 2.2 代码审查（跳过：superpowers 工具不可用，人工确认通过）

## 3. DRPlan/DRWorkflow 生成器

- [x] 3.1 实现 DRPlan 生成器  <!-- TDD 任务 -->
  - [x] 3.1.1 写失败测试：`internal/generator/planner_test.go`（测试 Stage DAG 生成、Hook→Job Action 映射、主资源→Subscription feeds、waitReady 标记、无 hook 场景只生成 install Stage）
  - [x] 3.1.2 验证测试失败（运行：`go test ./internal/generator/ -run TestPlanner -v`，确认失败原因是缺少功能）
  - [x] 3.1.3 写最小实现：`internal/generator/planner.go`（GeneratePlan 函数，从 ChartAnalysis 生成 DRPlan + []DRWorkflow + DRPlanExecution 样例）
  - [x] 3.1.4 验证测试通过（运行：`go test ./internal/generator/ -run TestPlanner -v`，确认所有测试通过）
  - [x] 3.1.5 重构：整理代码、改善命名、消除重复（保持测试通过）

- [x] 3.2 代码审查（跳过：superpowers 工具不可用，人工确认通过）

## 4. YAML 输出与 CLI 集成

- [x] 4.1 实现 YAML Writer  <!-- TDD 任务 -->
  - [x] 4.1.1 写失败测试：`internal/generator/writer_test.go`（测试 YAML 序列化输出、目录自动创建、文件命名规则）
  - [x] 4.1.2 验证测试失败（运行：`go test ./internal/generator/ -run TestWriter -v`，确认失败原因是缺少功能）
  - [x] 4.1.3 写最小实现：`internal/generator/writer.go`（WriteOutput 函数，将 DRPlan + []DRWorkflow 序列化为 YAML 文件）
  - [x] 4.1.4 验证测试通过（运行：`go test ./internal/generator/ -run TestWriter -v`，确认所有测试通过）
  - [x] 4.1.5 重构：整理代码、改善命名、消除重复（保持测试通过）

- [x] 4.2 实现 CLI 入口与 Makefile  <!-- 非 TDD 任务 -->
  - [x] 4.2.1 执行变更：`cmd/drplan-gen/main.go`（cobra 命令定义、参数解析、调用 parser→classifier→planner→writer 流水线）+ `Makefile` 新增 `build-gen` target
  - [x] 4.2.2 验证无回归（运行：`make build-gen`，确认编译通过且 `bin/drplan-gen` 生成）
  - [x] 4.2.3 检查：确认 CLI help 输出正确、参数验证完整

- [x] 4.3 代码审查（跳过：superpowers 工具不可用，人工确认通过）

## 5. 端到端测试（带 Hook 的测试 Chart）

- [x] 5.1 创建测试用 Helm chart 和渲染结果  <!-- 非 TDD 任务 -->
  - [x] 5.1.1 执行变更：`testdata/charts/demo-app/` + `testdata/rendered/demo-app.yaml`
  - [x] 5.1.2 验证无回归
  - [x] 5.1.3 检查：渲染输出中 hook annotation 完整保留

- [x] 5.2 端到端集成测试  <!-- TDD 任务 -->
  - [x] 5.2.1 写失败测试：`internal/generator/integration_test.go`
  - [x] 5.2.2 验证测试失败
  - [x] 5.2.3 写最小实现
  - [x] 5.2.4 验证测试通过（31/31 全部通过）
  - [x] 5.2.5 重构

- [x] 5.3 CLI 端到端验证  <!-- 非 TDD 任务 -->
  - [x] 5.3.1 执行 CLI 端到端测试
  - [x] 5.3.2 验证无回归
  - [x] 5.3.3 检查：3 个 Stage + Subscription feeds 正确

- [x] 5.4 代码审查（跳过：superpowers 工具不可用，人工确认通过）

## 6. Documentation Sync (Required)

- [x] 6.1 sync design.md: 更新 D3（Hook→Subscription 偏差）、Risk 1（Kind 限制）
- [x] 6.2 sync tasks.md: 全量标记所有层级任务
- [x] 6.3 sync proposal.md: 更新影响描述
- [x] 6.4 sync specs/drplan-gen.md: 更新 Hook 映射规则
- [x] 6.5 Final review: 所有 OpenSpec 文档反映实际实现
