# drplan-gen Helmfile 模式实施计划

> **供智能体执行时使用：** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按任务逐步落地。本计划使用复选框 `- [ ]` 跟踪执行状态。

**目标：** 为 `drplan-gen` 增加一个 `helmfile` 模式，将单个 helmfile release 解析为 `DRPlan + DRWorkflow + DRPlanExecution`，并新增一类一等公民的 `HelmChart` workflow action，让生成的 workflow 可以按 `HelmChart -> Globalization -> Subscription` 的顺序执行。

**架构：** 这次工作分成两条边界清晰的主线。控制器/API 主线负责新增 `HelmChart` action、executor、webhook 默认值与校验、以及 RBAC。生成器主线负责新增 `helmfile` 子命令和一套独立的 helmfile planner，用于解析单个 release 并生成内联 `HelmChart`、`Globalization`、`Subscription` action 的 workflow。现有的 rendered-YAML 生成路径保持不变。

**技术栈：** Go、Cobra、controller-runtime fake client、Kubebuilder/controller-gen、helmfile Go 模块、Ginkgo/Gomega

---

### 任务 1：先为 HelmChart Action 写失败测试

**涉及文件：**
- 新建：`internal/executor/helmchart_executor_test.go`
- 修改：`internal/webhook/drworkflow_webhook_test.go`

- [ ] **步骤 1：编写失败中的 executor 测试**

```go
package executor

func TestHelmChartExecutor_Create(t *testing.T) {
	sc := newTestScheme()
	fakeClient := fakeclient.NewClientBuilder().WithScheme(sc).Build()
	ex := &HelmChartActionExecutor{client: fakeClient}

	action := drv1alpha1.Action{
		Name: "test-helmchart",
		Type: drv1alpha1.ActionTypeHelmChart,
		HelmChart: &drv1alpha1.HelmChartAction{
			Operation: drv1alpha1.OperationApply,
			Name:      "demo-chart",
			Namespace: "default",
			Spec: &clusternetapps.HelmChartSpec{
				HelmOptions: clusternetapps.HelmOptions{
					Repository:   "oci://registry.example.com/charts",
					Chart:        "demo-app",
					ChartVersion: "1.2.3",
				},
				TargetNamespace: "default",
			},
		},
	}

	status, err := ex.Execute(context.Background(), &action, map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Outputs == nil || status.Outputs.HelmChartRef == nil {
		t.Fatalf("expected HelmChartRef to be recorded")
	}
}
```

- [ ] **步骤 2：补充 HelmChart 默认值与 rollback 校验的 webhook 测试**

```go
func TestDRWorkflowWebhook_DefaultSetsHelmChartDefaults(t *testing.T) {
	webhook := &DRWorkflowWebhook{}
	workflow := &drv1alpha1.DRWorkflow{
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{{
				Name: "chart",
				Type: drv1alpha1.ActionTypeHelmChart,
				HelmChart: &drv1alpha1.HelmChartAction{
					Name:      "demo-chart",
					Namespace: "default",
					Spec: &clusternetapps.HelmChartSpec{
						HelmOptions: clusternetapps.HelmOptions{
							Repository: "oci://registry.example.com/charts",
							Chart:      "demo-app",
						},
						TargetNamespace: "default",
					},
				},
			}},
		},
	}

	if err := webhook.Default(context.Background(), workflow); err != nil {
		t.Fatalf("default failed: %v", err)
	}
	if workflow.Spec.Actions[0].HelmChart.Operation != drv1alpha1.OperationCreate {
		t.Fatalf("expected Create default")
	}
}
```

- [ ] **步骤 3：运行测试并确认确实先失败**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor ./internal/webhook
```

预期：

- 因缺少 `ActionTypeHelmChart` 而出现编译错误
- 因缺少 `HelmChartActionExecutor` 而出现编译错误
- 因缺少 `HelmChart` webhook 默认值与校验逻辑而失败

### 任务 2：完整实现 HelmChart Action 链路

**涉及文件：**
- 修改：`api/v1alpha1/constants.go`
- 修改：`api/v1alpha1/common_types.go`
- 修改：`cmd/main.go`
- 新建：`internal/executor/helmchart_executor.go`
- 修改：`internal/webhook/drworkflow_webhook.go`
- 修改：`internal/controller/drworkflow_validator.go`
- 修改：`internal/controller/drplanexecution_controller.go`
- 修改：`install/helm/bcs-drplan-controller/templates/clusterrole.yaml`

- [ ] **步骤 1：补充 API 类型和输出字段**

```go
const (
	ActionTypeHelmChart = "HelmChart"
)

type HelmChartAction struct {
	Operation string `json:"operation,omitempty"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Spec      *clusternetapps.HelmChartSpec `json:"spec,omitempty"`
}

type Action struct {
	HelmChart *HelmChartAction `json:"helmChart,omitempty"`
}

type ActionOutputs struct {
	HelmChartRef *corev1.ObjectReference `json:"helmChartRef,omitempty"`
}
```

- [ ] **步骤 2：注册 executor 并补齐 RBAC**

```go
if err := registry.RegisterExecutor(executor.NewHelmChartActionExecutor(mgr.GetClient())); err != nil {
	return fmt.Errorf("register HelmChart executor: %w", err)
}
```

```go
// +kubebuilder:rbac:groups=apps.clusternet.io,resources=helmcharts;globalizations;localizations;subscriptions,verbs=get;list;watch;create;update;patch;delete
```

- [ ] **步骤 3：实现支持 Create/Apply/Patch/Delete 的 executor**

```go
func (e *HelmChartActionExecutor) Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	if err := e.validateHelmChartConfig(action); err != nil {
		return failHelmChartStatus(status, err.Error()), err
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "HelmChart",
	})
	obj.SetName(name)
	obj.SetNamespace(namespace)
	obj.Object["spec"] = specMap

	switch operation {
	case drv1alpha1.OperationDelete:
		err = client.IgnoreNotFound(e.client.Delete(ctx, obj))
	case drv1alpha1.OperationApply:
		err = e.client.Patch(ctx, obj, client.Apply, client.FieldOwner("drplan-controller"), client.ForceOwnership)
	case drv1alpha1.OperationPatch:
		err = e.patchHelmChart(ctx, obj)
	default:
		err = e.client.Create(ctx, obj)
	}
}
```

- [ ] **步骤 4：补充 webhook 默认值与校验**

```go
case "HelmChart":
	if action.HelmChart != nil && action.HelmChart.Operation == "" {
		action.HelmChart.Operation = drv1alpha1.OperationCreate
	}
```

```go
func (w *DRWorkflowWebhook) validateHelmChart(action *drv1alpha1.Action, index int) []string {
	if action.HelmChart == nil {
		return []string{fmt.Sprintf("action[%d] '%s': HelmChart configuration is required", index, action.Name)}
	}
	if action.HelmChart.Operation != drv1alpha1.OperationDelete && action.HelmChart.Spec == nil {
		return []string{fmt.Sprintf("action[%d] '%s': HelmChart.Spec is required when operation=%s", index, action.Name, effectiveOperation(action.HelmChart.Operation))}
	}
	return nil
}
```

- [ ] **步骤 5：运行测试并确认转绿**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor ./internal/webhook
```

预期：

- `ok .../internal/executor`
- `ok .../internal/webhook`

### 任务 3：新增 Helmfile 子命令和 CLI 测试

**涉及文件：**
- 修改：`cmd/drplan-gen/main.go`
- 新建：`cmd/drplan-gen/main_test.go`

- [ ] **步骤 1：先为子命令写失败测试**

```go
func TestHelmfileCommand_RequiresChartRepo(t *testing.T) {
	root := newRootCommand()
	root.SetArgs([]string{"helmfile", "-f", "helmfile.yaml", "-l", "name=demo"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "required flag \"chart-repo\" not set") {
		t.Fatalf("expected missing chart-repo error, got %v", err)
	}
}
```

- [ ] **步骤 2：重构 `main.go`，让命令构造函数可测试**

```go
func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{Use: "drplan-gen"}
	rootCmd.AddCommand(newHelmfileCommand())
	return rootCmd
}
```

- [ ] **步骤 3：新增 `helmfile` 子命令**

```go
func newHelmfileCommand() *cobra.Command {
	var (
		helmfileFile string
		selectors    []string
		namespace    string
		chartRepo    string
		plainHTTP    bool
		outputDir    string
	)

	cmd := &cobra.Command{
		Use: "helmfile",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(chartRepo) == "" {
				return fmt.Errorf("required flag \"chart-repo\" not set")
			}
			return runHelmfileMode(HelmfileGenerateConfig{
				File:      helmfileFile,
				Selectors: selectors,
				Namespace: namespace,
				ChartRepo: chartRepo,
				PlainHTTP: plainHTTP,
				OutputDir: outputDir,
			})
		},
	}
	return cmd
}
```

- [ ] **步骤 4：运行命令测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./cmd/drplan-gen
```

预期：

- 在子命令实现前失败
- 命令接线完成后通过

### 任务 4：实现 Helmfile Loader 和 Release 归一化

**涉及文件：**
- 修改：`go.mod`
- 新建：`internal/generator/helmfile_types.go`
- 新建：`internal/generator/helmfile_loader.go`
- 新建：`internal/generator/helmfile_loader_test.go`

- [ ] **步骤 1：定义解析后的 release 模型和 chart 名归一化 helper**

```go
type HelmfileResolvedRelease struct {
	ReleaseName     string
	Namespace       string
	Chart           string
	ChartVersion    string
	ChartRepo       string
	TargetNamespace string
	ValuesYAML      string
	Wait            *bool
	WaitForJob      *bool
	Atomic          *bool
	CreateNamespace *bool
	TimeoutSeconds  int32
	PlainHTTP       *bool
}

func normalizeChartName(chartRef, version string) string {
	base := filepath.Base(chartRef)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	suffix := "-" + version
	if strings.HasSuffix(base, suffix) {
		return strings.TrimSuffix(base, suffix)
	}
	return base
}
```

- [ ] **步骤 2：为单 release 选择和 chart 归一化编写失败测试**

```go
func TestNormalizeChartName_LocalTgz(t *testing.T) {
	got := normalizeChartName("./charts/bcs-services-stack-1.2.3.tgz", "1.2.3")
	if got != "bcs-services-stack" {
		t.Fatalf("got %q", got)
	}
}
```

```go
func TestSelectSingleRelease_ErrorsOnMultiple(t *testing.T) {
	_, err := selectSingleRelease([]HelmfileResolvedRelease{{ReleaseName: "a"}, {ReleaseName: "b"}})
	if err == nil || !strings.Contains(err.Error(), "expected exactly 1 release") {
		t.Fatalf("unexpected err: %v", err)
	}
}
```

- [ ] **步骤 3：实现 helmfile loader 封装**

```go
func LoadHelmfileRelease(input HelmfileLoadInput) (*HelmfileResolvedRelease, error) {
	// 1. 使用 helmfile 模块从 input.File 加载 state
	// 2. 应用 input.Selectors 和 input.Namespace
	// 3. 严格提取出唯一一个 release
	// 4. 将最终 merge 后 values 序列化为 YAML
	// 5. 组装 HelmfileResolvedRelease 返回
}
```

- [ ] **步骤 4：运行 loader 测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator -run 'TestNormalizeChartName|TestSelectSingleRelease'
```

预期：

- chart 名归一化测试通过
- 单 release 约束测试通过

### 任务 5：实现 Helmfile Planner

**涉及文件：**
- 修改：`internal/generator/types.go`
- 新建：`internal/generator/helmfile_planner.go`
- 新建：`internal/generator/helmfile_planner_test.go`
- 修改：`internal/generator/writer.go`
- 修改：`internal/generator/writer_test.go`

- [ ] **步骤 1：仅在需要模式元数据时扩展生成器类型，不引入额外资源文件输出**

```go
type HelmfileGenerateConfig struct {
	File      string
	Selectors []string
	Namespace string
	ChartRepo string
	PlainHTTP bool
	OutputDir string
}
```

- [ ] **步骤 2：为 planner 编写失败测试**

```go
func TestGenerateHelmfilePlan_WithValues_GeneratesHelmChartGlobalizationSubscription(t *testing.T) {
	release := HelmfileResolvedRelease{
		ReleaseName:     "nginx-demo",
		Namespace:       "default",
		TargetNamespace: "default",
		Chart:           "nginx",
		ChartVersion:    "0.1.1",
		ChartRepo:       "oci://registry.example.com/charts",
		ValuesYAML:      "replicaCount: 2\n",
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wf := string(result.WorkflowYAMLs["workflow-install.yaml"])
	if !strings.Contains(wf, "type: HelmChart") || !strings.Contains(wf, "type: Globalization") || !strings.Contains(wf, "type: Subscription") {
		t.Fatalf("workflow missing expected action types:\n%s", wf)
	}
}
```

- [ ] **步骤 3：实现 planner 输出**

```go
func GenerateHelmfilePlan(release HelmfileResolvedRelease) (*GenerateResult, error) {
	actions := []map[string]interface{}{
		buildHelmChartAction(release),
	}
	if strings.TrimSpace(release.ValuesYAML) != "" {
		actions = append(actions, buildGlobalizationAction(release))
	}
	actions = append(actions, buildHelmChartSubscriptionAction(release))

	return &GenerateResult{
		PlanYAML:       planBytes,
		WorkflowYAMLs:  map[string][]byte{"workflow-install.yaml": wfBytes},
		ExecutionYAMLs: executionYAMLs,
	}, nil
}
```

- [ ] **步骤 4：保持 writer 输出行为不变，并校验最终文件集合**

```go
Expect(filepath.Join(tmpDir, "drplan.yaml")).To(BeAnExistingFile())
Expect(filepath.Join(tmpDir, "workflow-install.yaml")).To(BeAnExistingFile())
Expect(filepath.Join(tmpDir, "drplanexecution-execute.yaml")).To(BeAnExistingFile())
Expect(filepath.Join(tmpDir, "drplanexecution-revert.yaml")).To(BeAnExistingFile())
```

- [ ] **步骤 5：运行 generator 测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator
```

预期：

- 现有 rendered-YAML 生成器测试继续通过
- 新增 helmfile planner 测试通过

### 任务 6：把子命令与 Loader、Planner 串起来

**涉及文件：**
- 修改：`cmd/drplan-gen/main.go`
- 修改：`internal/generator/writer.go`（仅在命令接线确实需要包装时修改）

- [ ] **步骤 1：串联 CLI -> loader -> planner -> writer**

```go
func runHelmfileMode(cfg generator.HelmfileGenerateConfig) error {
	release, err := generator.LoadHelmfileRelease(generator.HelmfileLoadInput{
		File:      cfg.File,
		Selectors: cfg.Selectors,
		Namespace: cfg.Namespace,
		ChartRepo: cfg.ChartRepo,
		PlainHTTP: cfg.PlainHTTP,
	})
	if err != nil {
		return fmt.Errorf("loading helmfile release: %w", err)
	}

	result, err := generator.GenerateHelmfilePlan(*release)
	if err != nil {
		return fmt.Errorf("generating helmfile plan: %w", err)
	}

	return generator.WriteOutput(result, cfg.OutputDir)
}
```

- [ ] **步骤 2：运行聚焦命令测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./cmd/drplan-gen
```

预期：

- command package passes

### 任务 7：生成代码、格式化并完成整体验证

**涉及文件：**
- 修改：`api/v1alpha1`、`config/crd/bases`、`config/rbac`、`install/helm/bcs-drplan-controller/crds` 下的生成文件

- [ ] **步骤 1：运行代码生成**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache make manifests generate
```

预期：

- CRDs and RBAC refreshed
- deepcopy regenerated

- [ ] **步骤 2：运行格式化**

运行：

```bash
gofmt -w api/v1alpha1/*.go cmd/drplan-gen/*.go internal/controller/*.go internal/executor/*.go internal/generator/*.go internal/webhook/*.go
```

预期：

- no output

- [ ] **步骤 3：运行针对性验证**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor ./internal/webhook ./internal/generator ./cmd/drplan-gen
```

预期：

- all listed packages pass

- [ ] **步骤 4：运行更大范围的编译验证**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./api/... ./cmd/... ./internal/controller -run TestDoesNotExist
```

预期：

- compile passes for API, commands, and controller packages

- [ ] **步骤 5：提交变更**

```bash
git add api/v1alpha1 cmd/drplan-gen internal/controller internal/executor internal/generator internal/webhook config/rbac config/crd install/helm/bcs-drplan-controller docs/superpowers/specs/2026-04-21-drplan-gen-helmfile-design.md docs/superpowers/plans/2026-04-21-drplan-gen-helmfile.md
git commit -m "feat: add helmfile mode for drplan-gen"
```
