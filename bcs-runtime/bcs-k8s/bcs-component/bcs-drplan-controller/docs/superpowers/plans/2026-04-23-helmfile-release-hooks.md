# helmfile release hooks 实施计划

> **供智能体执行时使用：** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按任务逐步落地。本计划使用复选框 `- [ ]` 跟踪执行状态。

**目标：** 为 `drplan-gen helmfile` 增加 `releases[].hooks` 迁移能力，将 `preapply`、`presync`、`postsync` 转换为“hub 侧 Manifest(template=Job) + PerCluster Subscription(feed=Job)”的多 stage DRPlan。

**架构：** 这次实现分成四条主线。
第一条主线在 helmfile loader 中提取并归一化 release hook。
第二条主线在 helmfile planner 中切换到 hook-aware 生成模式，按实际 hook 事件输出
`preapply`、`presync`、`postsync` stage / workflow，并始终输出 `execute`。
第三条主线增强 `Subscription waitReady`，让生成器输出的真实 Job feed 能在子集群等待最终 Job 状态，并保留旧 Manifest feed 兼容。
第四条主线调整 hook 场景下的 `PerCluster` 聚合行为，保证所有目标子集群都执行结束后再聚合。

**技术栈：** Go、Cobra、controller-runtime fake client、helmfile Go 模块、Clusternet `Manifest` / `Subscription`、Kubebuilder CRD 类型、`go test`、`golangci-lint`

---

### 任务 1：在 loader 中提取并归一化 helmfile release hooks

**涉及文件：**
- 新建：`testdata/helmfile/hooks/helmfile.yaml.gotmpl`
- 修改：`internal/generator/helmfile_types.go`
- 修改：`internal/generator/helmfile_loader.go`
- 修改：`internal/generator/helmfile_loader_test.go`
- 修改：`cmd/drplan-gen/main.go`

- [ ] **步骤 1：先写失败测试，固定支持事件与顺序**

```go
func TestLoadHelmfileRelease_NormalizesSupportedHooks(t *testing.T) {
	release, err := LoadHelmfileRelease(HelmfileLoadInput{
		File:      filepath.Join("..", "..", "testdata", "helmfile", "hooks", "helmfile.yaml.gotmpl"),
		Selectors: []string{"name=demo-app"},
		ChartRepo: "oci://registry.example.com/charts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(release.Hooks) != 3 {
		t.Fatalf("expected 3 supported hooks, got %d", len(release.Hooks))
	}
	if release.Hooks[0].Event != "preapply" || release.Hooks[1].Event != "presync" || release.Hooks[2].Event != "postsync" {
		t.Fatalf("unexpected hook order: %#v", release.Hooks)
	}
	if release.Hooks[0].Command != "./scripts/pre.sh" {
		t.Fatalf("unexpected first hook command: %q", release.Hooks[0].Command)
	}
}
```

- [ ] **步骤 2：补一个忽略 `showlogs` 和不支持事件的测试**

```go
func TestLoadHelmfileRelease_IgnoresShowLogsAndUnsupportedEvents(t *testing.T) {
	release, err := LoadHelmfileRelease(HelmfileLoadInput{
		File:      filepath.Join("..", "..", "testdata", "helmfile", "hooks", "helmfile.yaml.gotmpl"),
		Selectors: []string{"name=demo-app"},
		ChartRepo: "oci://registry.example.com/charts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, hook := range release.Hooks {
		if hook.Event == "cleanup" {
			t.Fatal("cleanup should not be normalized in v1")
		}
	}
}
```

- [ ] **步骤 3：运行测试并确认先失败**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator -run 'TestLoadHelmfileRelease_(NormalizesSupportedHooks|IgnoresShowLogsAndUnsupportedEvents)'
```

预期：

- 编译失败，提示 `HelmfileResolvedRelease` 没有 `Hooks`
- 或测试失败，提示当前 loader 未提取 release hooks

- [ ] **步骤 4：在归一化模型中新增 hook 结构**

```go
type HelmfileResolvedHook struct {
	Event   string
	Command string
	Args    []string
	Order   int
}

type HelmfileGenerateConfig struct {
	File           string
	Selectors      []string
	Namespace      string
	ChartRepo      string
	HookImage      string
	PlainHTTP      bool
	KeepFullValues bool
	OutputDir      string
}

type HelmfileLoadInput struct {
	File           string
	Selectors      []string
	Namespace      string
	ChartRepo      string
	HookImage      string
	PlainHTTP      bool
	KeepFullValues bool
}

type HelmfileResolvedRelease struct {
	ReleaseName     string
	Namespace       string
	Chart           string
	ChartVersion    string
	ChartRepo       string
	HookImage       string
	TargetNamespace string
	ValuesYAML      string
	Hooks           []HelmfileResolvedHook
	Wait            *bool
	WaitForJob      *bool
	Atomic          *bool
	CreateNamespace *bool
	TimeoutSeconds  int32
	PlainHTTP       *bool
}
```

- [ ] **步骤 5：在 loader 中提取 release hooks，并只保留首版支持事件**

```go
func normalizeHelmfileHooks(release *state.ReleaseSpec) []HelmfileResolvedHook {
	var hooks []HelmfileResolvedHook
	order := 0
	for _, hook := range release.Hooks {
		for _, event := range hook.Events {
			switch strings.TrimSpace(event) {
			case "preapply", "presync", "postsync":
				hooks = append(hooks, HelmfileResolvedHook{
					Event:   event,
					Command: hook.Command,
					Args:    append([]string(nil), hook.Args...),
					Order:   order,
				})
				order++
			}
		}
	}
	return hooks
}
```

- [ ] **步骤 6：把提取出的 hooks 挂到最终 release 结果上**

```go
return &HelmfileResolvedRelease{
	ReleaseName:     release.Name,
	Namespace:       namespace,
	Chart:           chartName,
	ChartVersion:    chartVersion,
	ChartRepo:       input.ChartRepo,
	HookImage:       input.HookImage,
	TargetNamespace: targetNamespace,
	ValuesYAML:      valuesYAML,
	Hooks:           normalizeHelmfileHooks(release),
	Wait:            effectiveReleaseWait(st, release),
	WaitForJob:      effectiveReleaseWaitForJob(st, release),
	Atomic:          effectiveReleaseAtomic(st, release),
	CreateNamespace: effectiveReleaseCreateNamespace(release),
	TimeoutSeconds:  timeoutSeconds,
	PlainHTTP:       boolPtrForLoader(input.PlainHTTP),
}, nil
```

- [ ] **步骤 7：回跑 loader 测试确认转绿**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator -run 'TestLoadHelmfileRelease_(NormalizesSupportedHooks|IgnoresShowLogsAndUnsupportedEvents)'
```

预期：

- `ok .../internal/generator`

- [ ] **步骤 8：提交这一组变更**

```bash
git add testdata/helmfile/hooks/helmfile.yaml.gotmpl internal/generator/helmfile_types.go internal/generator/helmfile_loader.go internal/generator/helmfile_loader_test.go cmd/drplan-gen/main.go
git commit -m "feat: load helmfile release hooks"
```

### 任务 2：让 helmfile planner 在有 hooks 时生成多 stage / 多 workflow

**涉及文件：**
- 修改：`internal/generator/helmfile_planner.go`
- 修改：`internal/generator/helmfile_planner_test.go`
- 修改：`internal/generator/integration_test.go`
- 修改：`cmd/drplan-gen/main.go`

- [ ] **步骤 1：先写失败测试，固定 hook-aware 输出结构**

```go
func TestGenerateHelmfilePlan_WithHooks_UsesMultiStageWorkflows(t *testing.T) {
	release := HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "default",
		Chart:           "demo",
		ChartRepo:       "oci://registry.example.com/charts",
		Hooks: []HelmfileResolvedHook{
			{Event: "preapply", Command: "./scripts/pre.sh", Order: 0},
			{Event: "presync", Command: "./scripts/sync.sh", Order: 1},
			{Event: "postsync", Command: "./scripts/post.sh", Order: 2},
		},
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result.WorkflowYAMLs["workflow-preapply.yaml"]; !ok {
		t.Fatal("expected workflow-preapply.yaml")
	}
	if _, ok := result.WorkflowYAMLs["workflow-presync.yaml"]; !ok {
		t.Fatal("expected workflow-presync.yaml")
	}
	if _, ok := result.WorkflowYAMLs["workflow-postsync.yaml"]; !ok {
		t.Fatal("expected workflow-postsync.yaml")
	}
}
```

- [ ] **步骤 2：补一个 Manifest(Job) 与 hookCleanup 的测试**

```go
func TestGenerateHelmfilePlan_HookWorkflowUsesManifestAndPerClusterSubscription(t *testing.T) {
	release := HelmfileResolvedRelease{
		ReleaseName:     "demo-app",
		Namespace:       "default",
		TargetNamespace: "default",
		Chart:           "demo",
		ChartRepo:       "oci://registry.example.com/charts",
		Hooks: []HelmfileResolvedHook{
			{Event: "presync", Command: "./scripts/sync.sh", Args: []string{"a", "b"}, Order: 0},
		},
	}

	result, err := GenerateHelmfilePlan(release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf := string(result.WorkflowYAMLs["workflow-presync.yaml"])
	if !strings.Contains(wf, "kind: Manifest") {
		t.Fatal("expected Manifest template carrier")
	}
	if !strings.Contains(wf, "apiVersion: batch/v1") {
		t.Fatal("expected Job feed")
	}
	if !strings.Contains(wf, "clusterExecutionMode: PerCluster") {
		t.Fatal("expected PerCluster hook subscription")
	}
	if !strings.Contains(wf, "beforeCreate: true") {
		t.Fatal("expected hookCleanup.beforeCreate")
	}
	if !strings.Contains(wf, "ttlSecondsAfterFinished: 300") {
		t.Fatal("expected ttlSecondsAfterFinished")
	}
}
```

- [ ] **步骤 3：运行测试并确认先失败**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator -run 'TestGenerateHelmfilePlan_(WithHooks_UsesMultiStageWorkflows|HookWorkflowUsesManifestAndPerClusterSubscription)'
```

预期：

- 测试失败，提示当前只生成 `workflow-execute.yaml`

- [ ] **步骤 4：在 planner 中增加 hook-aware 分支**

```go
func GenerateHelmfilePlan(release HelmfileResolvedRelease) (*GenerateResult, error) {
	if err := validateHelmfileRelease(release); err != nil {
		return nil, err
	}
	if len(release.Hooks) == 0 {
		return generateSimpleHelmfilePlan(release)
	}
	return generateHelmfileHookAwarePlan(release)
}
```

- [ ] **步骤 5：实现多 stage 计划结构与 `postsync` 的 `Continue` 策略**

```go
func buildHelmfileHookAwarePlanYAML(release HelmfileResolvedRelease) map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "dr.bkbcs.tencent.com/v1alpha1",
		"kind":       "DRPlan",
		"metadata": map[string]interface{}{
			"name":      release.ReleaseName,
			"namespace": release.Namespace,
		},
		"spec": map[string]interface{}{
			"description":   fmt.Sprintf("Auto-generated from Helmfile release: %s", release.ReleaseName),
			"failurePolicy": "Continue",
			"globalParams": []map[string]interface{}{
				{"name": "feedNamespace", "value": release.Namespace},
				{"name": "targetNamespace", "value": release.TargetNamespace},
				{"name": "hookImage", "value": release.HookImage},
			},
			"stages": buildHelmfileHookStages(release),
		},
	}
}
```

- [ ] **步骤 6：为每个 hook 生成一组 `Manifest + Subscription` actions**

```go
func buildHelmfileHookActions(release HelmfileResolvedRelease, event string) []map[string]interface{} {
	var actions []map[string]interface{}
	for _, hook := range release.Hooks {
		if hook.Event != event {
			continue
		}
		actions = append(actions, buildHelmfileHookManifestAction(release, hook))
		actions = append(actions, buildHelmfileHookSubscriptionAction(release, hook))
	}
	return actions
}
```

```go
func buildHelmfileHookSubscriptionAction(release HelmfileResolvedRelease, hook HelmfileResolvedHook) map[string]interface{} {
	name := hookActionBaseName(release, hook)
	return map[string]interface{}{
		"name":                 name,
		"type":                 "Subscription",
		"waitReady":            true,
		"clusterExecutionMode": "PerCluster",
		"hookCleanup": map[string]interface{}{
			"beforeCreate": true,
		},
		"hookType": hook.Event,
		"timeout":  "10m",
		"subscription": map[string]interface{}{
			"operation": "Create",
			"name":      fmt.Sprintf("%s-sub", name),
			"namespace": feedNamespaceRef,
			"spec": map[string]interface{}{
				"schedulingStrategy": "Replication",
				"feeds": []map[string]interface{}{{
					"apiVersion": "apps.clusternet.io/v1alpha1",
					"kind":       "Manifest",
					"name":       name,
					"namespace":  feedNamespaceRef,
				}},
				"subscribers": []map[string]interface{}{
					{"clusterAffinity": map[string]interface{}{}},
				},
			},
		},
	}
}
```

- [ ] **步骤 7：为 hook Job 模板固定稳定命名和默认生命周期**

```go
func buildHookJobTemplate(release HelmfileResolvedRelease, hook HelmfileResolvedHook) map[string]interface{} {
	name := hookActionBaseName(release, hook)
	return map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": feedNamespaceRef,
		},
		"spec": map[string]interface{}{
			"backoffLimit":            0,
			"ttlSecondsAfterFinished": 300,
			"template": map[string]interface{}{
				"spec": map[string]interface{}{
					"restartPolicy": "Never",
					"containers": []map[string]interface{}{{
						"name":    "hook",
						"image":   "$(params.hookImage)",
						"command": []string{"/bin/bash", "-c", buildHookShellCommand(hook)},
					}},
				},
			},
		},
	}
}
```

- [ ] **步骤 8：给 helmfile 子命令增加 `--hook-image` 必填参数**

```go
cmd.Flags().StringVar(&cfg.HookImage, "hook-image", "", "Hook runner image used by helmfile release hooks")
```

```go
if len(release.Hooks) > 0 && strings.TrimSpace(cfg.HookImage) == "" {
	return fmt.Errorf("required flag \"hook-image\" not set when helmfile release hooks are present")
}
```

- [ ] **步骤 9：回跑 planner / integration 测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator -run 'TestGenerateHelmfilePlan_(WithHooks_UsesMultiStageWorkflows|HookWorkflowUsesManifestAndPerClusterSubscription)'
```

预期：

- `ok .../internal/generator`

- [ ] **步骤 10：提交这一组变更**

```bash
git add internal/generator/helmfile_planner.go internal/generator/helmfile_planner_test.go internal/generator/integration_test.go
git commit -m "feat: generate hook-aware helmfile workflows"
```

### 任务 3：增强 `Subscription waitReady`，支持 `Manifest(template=Job)`

**涉及文件：**
- 修改：`internal/executor/subscription_executor.go`
- 修改：`internal/executor/subscription_waitready_test.go`

- [ ] **步骤 1：先写失败测试，固定 `Manifest(template=Job)` 的等待行为**

```go
func TestIsFeedReadyInChildCluster_ManifestTemplateJobComplete(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clusternetapps.AddToScheme(scheme)

	manifest := &unstructured.Unstructured{}
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Manifest",
	})
	manifest.SetNamespace("default")
	manifest.SetName("hook-job")
	manifest.Object["template"] = map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      "hook-job",
			"namespace": "default",
		},
	}

	childJob := &unstructured.Unstructured{}
	childJob.SetGroupVersionKind(schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"})
	childJob.SetNamespace("default")
	childJob.SetName("hook-job")
	childJob.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"type": "Complete", "status": "True"},
		},
	}

	executor := &SubscriptionActionExecutor{
		client: fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(manifest).Build(),
	}
	childClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJob).Build()

	ready, _, err := executor.isFeedReadyInChildCluster(context.Background(), childClient, clusternetapps.Feed{
		APIVersion: "apps.clusternet.io/v1alpha1",
		Kind:       "Manifest",
		Name:       "hook-job",
		Namespace:  "default",
	})
	if err != nil || !ready {
		t.Fatalf("expected ready, got ready=%v err=%v", ready, err)
	}
}
```

- [ ] **步骤 2：补一个失败态测试**

```go
func TestIsFeedReadyInChildCluster_ManifestTemplateJobFailed(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clusternetapps.AddToScheme(scheme)

	manifest := &unstructured.Unstructured{}
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Manifest",
	})
	manifest.SetNamespace("default")
	manifest.SetName("hook-job")
	manifest.Object["template"] = map[string]interface{}{
		"apiVersion": "batch/v1",
		"kind":       "Job",
		"metadata": map[string]interface{}{
			"name":      "hook-job",
			"namespace": "default",
		},
	}

	childJob := &unstructured.Unstructured{}
	childJob.SetGroupVersionKind(schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"})
	childJob.SetNamespace("default")
	childJob.SetName("hook-job")
	childJob.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"type": "Failed", "status": "True"},
		},
	}

	executor := &SubscriptionActionExecutor{
		client: fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(manifest).Build(),
	}
	childClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJob).Build()

	_, _, err := executor.isFeedReadyInChildCluster(context.Background(), childClient, clusternetapps.Feed{
		APIVersion: "apps.clusternet.io/v1alpha1",
		Kind:       "Manifest",
		Name:       "hook-job",
		Namespace:  "default",
	})
	if err == nil {
		t.Fatal("expected failed job to return error")
	}
}
```

- [ ] **步骤 3：运行测试并确认先失败**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor -run 'TestIsFeedReadyInChildCluster_ManifestTemplateJob(Complete|Failed)'
```

预期：

- 当前实现会在子集群直接查 `Manifest`，因此失败

- [ ] **步骤 4：在 waitReady 链路里解析 Manifest 模板目标**

```go
func (e *SubscriptionActionExecutor) resolveFeedTarget(
	ctx context.Context,
	feed clusternetapps.Feed,
) (clusternetapps.Feed, error) {
	if feed.Kind != "Manifest" || feed.APIVersion != "apps.clusternet.io/v1alpha1" {
		return feed, nil
	}

	manifest := &unstructured.Unstructured{}
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Manifest",
	})
	if err := e.client.Get(ctx, client.ObjectKey{Namespace: feed.Namespace, Name: feed.Name}, manifest); err != nil {
		return clusternetapps.Feed{}, err
	}

	template, found, err := unstructured.NestedMap(manifest.Object, "template")
	if err != nil || !found {
		return clusternetapps.Feed{}, fmt.Errorf("manifest %s/%s has no template", feed.Namespace, feed.Name)
	}
	return feedFromManifestTemplate(template)
}
```

```go
func feedFromManifestTemplate(template map[string]interface{}) (clusternetapps.Feed, error) {
	apiVersion, _, _ := unstructured.NestedString(template, "apiVersion")
	kind, _, _ := unstructured.NestedString(template, "kind")
	name, _, _ := unstructured.NestedString(template, "metadata", "name")
	namespace, _, _ := unstructured.NestedString(template, "metadata", "namespace")
	if apiVersion == "" || kind == "" || name == "" {
		return clusternetapps.Feed{}, fmt.Errorf("manifest template must include apiVersion, kind, and metadata.name")
	}
	return clusternetapps.Feed{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Namespace:  namespace,
	}, nil
}
```

- [ ] **步骤 5：在子集群检查前统一先解析 feed 目标**

```go
func (e *SubscriptionActionExecutor) isFeedReadyInChildCluster(
	ctx context.Context,
	childClient client.Client,
	feed clusternetapps.Feed,
) (bool, string, error) {
	targetFeed, err := e.resolveFeedTarget(ctx, feed)
	if err != nil {
		return false, "", err
	}

	gv, err := schema.ParseGroupVersion(targetFeed.APIVersion)
	if err != nil {
		return false, "", fmt.Errorf("parse feed apiVersion %q: %w", targetFeed.APIVersion, err)
	}
	target := &unstructured.Unstructured{}
	target.SetGroupVersionKind(schema.GroupVersionKind{
		Group: gv.Group, Version: gv.Version, Kind: targetFeed.Kind,
	})
	if err := childClient.Get(ctx, client.ObjectKey{
		Namespace: targetFeed.Namespace, Name: targetFeed.Name,
	}, target); err != nil {
		// 保持现有 not found 行为
	}
	return evaluateResourceReadiness(target)
}
```

- [ ] **步骤 6：回跑 waitReady 测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor -run 'Test(IsFeedReadyInChildCluster_ManifestTemplateJob(Complete|Failed)|EvaluateResourceReadiness)'
```

预期：

- `ok .../internal/executor`

- [ ] **步骤 7：提交这一组变更**

```bash
git add internal/executor/subscription_executor.go internal/executor/subscription_waitready_test.go
git commit -m "feat: wait for manifest-backed hook jobs"
```

### 任务 4：调整 hook 场景下的 `PerCluster` 聚合语义

**涉及文件：**
- 修改：`internal/executor/native_executor.go`
- 修改：`internal/executor/native_executor_test.go`
- 修改：`internal/executor/percluster_test.go`

- [ ] **步骤 1：先写失败测试，固定“单集群失败不取消其他集群”**

```go
func makeManagedCluster(namespace, name, clusterID string) *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{}
	cluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	cluster.SetNamespace(namespace)
	cluster.SetName(name)
	cluster.Object["spec"] = map[string]interface{}{"clusterId": clusterID}
	return cluster
}

func makeChildJobWithCondition(namespace, name, condType string) *unstructured.Unstructured {
	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "batch",
		Version: "v1",
		Kind:    "Job",
	})
	job.SetNamespace(namespace)
	job.SetName(name)
	job.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"type": condType, "status": "True"},
		},
	}
	return job
}

func TestExecuteClusterActions_HookActionDoesNotCancelOtherClusters(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	clusterA := makeManagedCluster("ns1", "cluster-a", "cluster-a-id")
	clusterB := makeManagedCluster("ns2", "cluster-b", "cluster-b-id")
	childJobFailed := makeChildJobWithCondition("app-ns", "hook-job", "Failed")
	childJobComplete := makeChildJobWithCondition("app-ns", "hook-job", "Complete")

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(clusterA, clusterB).Build()
	subExec := &SubscriptionActionExecutor{
		client: k8sClient,
		childClientFactory: &fakeChildClusterClientFactory{
			clients: map[string]client.Client{
				"cluster-a-id": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJobFailed).Build(),
				"cluster-b-id": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJobComplete).Build(),
			},
		},
	}
	executor := NewNativeWorkflowExecutor(k8sClient, nil)

	actions := []drv1alpha1.Action{{
		Name:                 "presync-job",
		Type:                 drv1alpha1.ActionTypeSubscription,
		HookType:             "presync",
		WaitReady:            true,
		ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
		Subscription: &drv1alpha1.SubscriptionAction{
			Operation: drv1alpha1.OperationCreate,
			Name:      "presync-job-sub",
			Namespace: "default",
			Spec: &clusternetapps.SubscriptionSpec{
				SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
				Feeds: []clusternetapps.Feed{
					{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "app-ns"},
				},
			},
		},
		Timeout: "10s",
	}}

	statusMap := make(map[string][]drv1alpha1.ClusterActionStatus)
	childRefMap := make(map[string][]corev1.ObjectReference)
	var mu sync.Mutex

	executor.executeClusterActions(
		context.Background(),
		subExec,
		actions,
		perClusterActionTargets{"presync-job": {"ns1/cluster-a": {}, "ns2/cluster-b": {}}},
		"ns1/cluster-a",
		nil,
		statusMap,
		childRefMap,
		&mu,
		drv1alpha1.FailurePolicyFailFast,
	)
	executor.executeClusterActions(
		context.Background(),
		subExec,
		actions,
		perClusterActionTargets{"presync-job": {"ns1/cluster-a": {}, "ns2/cluster-b": {}}},
		"ns2/cluster-b",
		nil,
		statusMap,
		childRefMap,
		&mu,
		drv1alpha1.FailurePolicyFailFast,
	)

	statuses := statusMap["presync-job"]
	if len(statuses) != 2 {
		t.Fatalf("expected 2 cluster statuses, got %#v", statuses)
	}
	if aggregateClusterStatuses(statuses) != drv1alpha1.PhaseFailed {
		t.Fatalf("expected aggregated phase Failed, got %#v", statuses)
	}
}
```

- [ ] **步骤 2：补一个非 hook 场景回归测试**

```go
func TestExecuteClusterActions_NonHookPerClusterStillFailFast(t *testing.T) {
	actions := []drv1alpha1.Action{{
		Name:                 "ordinary-subscription",
		Type:                 drv1alpha1.ActionTypeSubscription,
		WaitReady:            true,
		ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
		Subscription: &drv1alpha1.SubscriptionAction{
			Operation: drv1alpha1.OperationCreate,
			Name:      "ordinary-subscription",
			Namespace: "default",
			Spec:      &clusternetapps.SubscriptionSpec{},
		},
	}}

	if !shouldFailFastPerCluster(actions, drv1alpha1.FailurePolicyFailFast) {
		t.Fatal("expected non-hook per-cluster action to remain fail-fast")
	}
}
```

- [ ] **步骤 3：运行测试并确认先失败**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor -run 'TestExecuteClusterActions_(HookActionDoesNotCancelOtherClusters|NonHookPerClusterStillFailFast)'
```

预期：

- 现状下 hook 场景会触发 `cancelFn()`，因此测试失败

- [ ] **步骤 4：把“是否提前取消其他 cluster”收敛成显式判断**

```go
func shouldFailFastPerCluster(actions []drv1alpha1.Action, failurePolicy string) bool {
	if !isFailFast(failurePolicy) {
		return false
	}
	for i := range actions {
		if actions[i].HookType != "" || actions[i].HookCleanup != nil {
			return false
		}
	}
	return true
}
```

- [ ] **步骤 5：在 cluster fan-out 执行路径中只对非 hook 场景调用 `cancelFn()`**

```go
failFastClusters := shouldFailFastPerCluster(actions, failurePolicy)
cs, childRef, _ := subExec.ExecuteForCluster(ctx, &action, clusterBinding, params)
mu.Lock()
statusMap[action.Name] = append(statusMap[action.Name], *cs)
if childRef != nil {
	childRefMap[action.Name] = append(childRefMap[action.Name], *childRef)
}
mu.Unlock()
if cs.Phase == drv1alpha1.PhaseFailed {
	if failFastClusters && cancelFn != nil {
		cancelFn()
	}
	break
}
```

- [ ] **步骤 6：保持最终聚合规则不变，但确保聚合发生在所有 cluster 状态都收集完成后**

```go
overall := aggregateClusterStatuses(clusterStatuses)
if overall == drv1alpha1.PhaseFailed {
	status.Message = fmt.Sprintf("%d/%d clusters failed", failedCount, len(clusterStatuses))
}
```

- [ ] **步骤 7：回跑 hook 聚合测试**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/executor -run 'TestExecuteClusterActions_(HookActionDoesNotCancelOtherClusters|NonHookPerClusterStillFailFast)'
```

预期：

- `ok .../internal/executor`

- [ ] **步骤 8：提交这一组变更**

```bash
git add internal/executor/native_executor.go internal/executor/native_executor_test.go internal/executor/percluster_test.go
git commit -m "feat: keep hook per-cluster execution isolated"
```

### 任务 5：同步 CLI 帮助、用户文档并做最终验证

**涉及文件：**
- 修改：`cmd/drplan-gen/main.go`
- 修改：`docs/drplan-gen-guide.md`
- 修改：`docs/superpowers/specs/2026-04-23-helmfile-release-hooks-design.md`

- [ ] **步骤 1：更新 CLI 帮助文案，说明 helmfile hooks 的生成模式**

```go
Long: "drplan-gen reads rendered Kubernetes YAML (from helm template, helmfile template, etc.)\n" +
	"and generates DRPlan, DRWorkflow, and DRPlanExecution YAML files.\n\n" +
	"Helmfile mode now supports release hooks:\n" +
	"  - preapply / presync / postsync are converted to hook-aware stages\n" +
	"  - each hook is generated as Manifest(template=Job) + PerCluster Subscription(feed=Job)\n" +
	"  - postsync relies on plan failurePolicy=Continue to run after execute failures\n" +
	"  - --hook-image is required when the selected helmfile release contains hooks\n"
```

- [ ] **步骤 2：在 `drplan-gen-guide.md` 中补上 hook 迁移章节**

```md
### Helmfile release hooks

当 helmfile release 包含 `preapply`、`presync`、`postsync` 时，`drplan-gen helmfile` 会切换到 hook-aware 模式：

- 存在 `preapply` hook 时生成 `workflow-preapply.yaml`
- 存在 `presync` hook 时生成 `workflow-presync.yaml`
- 保留 `workflow-execute.yaml`
- 存在 `postsync` hook 时生成 `workflow-postsync.yaml`

每个 hook 会展开成两步：

1. 在 hub 集群创建 `Manifest(template=Job)`
2. 通过 `Subscription + PerCluster + waitReady` 将 Job 下发到每个目标子集群执行
```

- [ ] **步骤 3：把最终行为和 spec 对齐，删除已经失效的描述**

```md
- 若源 helmfile hook 带有 `showlogs`，首版直接忽略
- hook Job 成功判定以子集群 `Job Complete/Failed` 条件为准
- hook 场景下的 `PerCluster` 必须等待所有目标子集群执行结束后再聚合
```

- [ ] **步骤 4：运行最终验证**

运行：

```bash
env GOCACHE=/tmp/gocache GOMODCACHE=/tmp/gomodcache go test ./internal/generator ./internal/executor ./cmd/drplan-gen
make lint
```

预期：

- `ok .../internal/generator`
- `ok .../internal/executor`
- `ok .../cmd/drplan-gen`
- `make lint` 输出 `0 issues`

- [ ] **步骤 5：提交最终文档与验证收口**

```bash
git add cmd/drplan-gen/main.go docs/drplan-gen-guide.md docs/superpowers/specs/2026-04-23-helmfile-release-hooks-design.md
git commit -m "docs: describe helmfile hook generation"
```

## 自检结论

- spec 覆盖情况：
  - helmfile hook 解析：任务 1
  - 多 stage / 多 workflow 生成：任务 2
  - `Manifest(template=Job)` waitReady：任务 3
  - PerCluster 子集群互不干扰：任务 4
  - `postsync` 与文档收口：任务 5
- 无占位项：
  - 各任务都给出了明确文件路径、测试入口、命令和核心实现骨架
- 类型一致性：
  - 统一使用 `HelmfileResolvedHook`
  - 统一以 `Manifest(template=Job) + Subscription(PerCluster)` 作为 hook 执行模型
  - 统一以 `HookType != "" || HookCleanup != nil` 区分 hook 场景下的特殊 PerCluster 聚合逻辑
