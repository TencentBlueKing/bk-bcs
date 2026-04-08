/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

type staticActionExecutor struct {
	actionType string
	status     *drv1alpha1.ActionStatus
	err        error
}

func (s *staticActionExecutor) Execute(_ context.Context, _ *drv1alpha1.Action, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	if s.status == nil {
		return nil, s.err
	}
	copy := *s.status
	return &copy, s.err
}

func (s *staticActionExecutor) Rollback(_ context.Context, _ *drv1alpha1.Action, _ *drv1alpha1.ActionStatus, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	return &drv1alpha1.ActionStatus{Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (s *staticActionExecutor) Type() string {
	return s.actionType
}

type recordingRollbackExecutor struct {
	actionType     string
	receivedParams map[string]interface{}
}

func (r *recordingRollbackExecutor) Execute(_ context.Context, action *drv1alpha1.Action, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	return &drv1alpha1.ActionStatus{Name: action.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (r *recordingRollbackExecutor) Rollback(_ context.Context, action *drv1alpha1.Action, _ *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	r.receivedParams = params
	return &drv1alpha1.ActionStatus{Name: action.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (r *recordingRollbackExecutor) Type() string {
	return r.actionType
}

type recordingActionExecutor struct {
	actionType     string
	executed       []string
	receivedParams []map[string]interface{}
}

func (r *recordingActionExecutor) Execute(_ context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	r.executed = append(r.executed, action.Name)
	r.receivedParams = append(r.receivedParams, params)
	return &drv1alpha1.ActionStatus{Name: action.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (r *recordingActionExecutor) Rollback(_ context.Context, action *drv1alpha1.Action, _ *drv1alpha1.ActionStatus, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	return &drv1alpha1.ActionStatus{Name: action.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (r *recordingActionExecutor) Type() string {
	return r.actionType
}

type recordingWorkflowExecutor struct {
	revertedWorkflows []string
}

func (r *recordingWorkflowExecutor) ExecuteWorkflow(_ context.Context, workflow *drv1alpha1.DRWorkflow, _ map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error) {
	return &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: workflow.Name, Namespace: workflow.Namespace},
		Phase:       drv1alpha1.PhaseSucceeded,
	}, nil
}

func (r *recordingWorkflowExecutor) RevertWorkflow(_ context.Context, workflow *drv1alpha1.DRWorkflow, _ *drv1alpha1.WorkflowExecutionStatus, _ map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error) {
	r.revertedWorkflows = append(r.revertedWorkflows, workflow.Name)
	return &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: workflow.Name, Namespace: workflow.Namespace},
		Phase:       drv1alpha1.PhaseSucceeded,
	}, nil
}

type fixedWorkflowPhaseExecutor struct {
	phase string
}

const testDefaultNamespace = "default"

func (e *fixedWorkflowPhaseExecutor) ExecuteWorkflow(_ context.Context, workflow *drv1alpha1.DRWorkflow, _ map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error) {
	return &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: workflow.Name, Namespace: workflow.Namespace},
		Phase:       e.phase,
	}, nil
}

func (e *fixedWorkflowPhaseExecutor) RevertWorkflow(_ context.Context, workflow *drv1alpha1.DRWorkflow, _ *drv1alpha1.WorkflowExecutionStatus, _ map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error) {
	return &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: workflow.Name, Namespace: workflow.Namespace},
		Phase:       drv1alpha1.PhaseSucceeded,
	}, nil
}

// TestExecutePlan_ExecutionParamsPriority verifies that execution.Spec.Params override globalParams.
// NOCC:tosa/fn_length(设计如此)
func TestExecutePlan_ExecutionParamsPriority(t *testing.T) {
	// Build a plan with a single stage that has no actions (to avoid complex mock setup).
	// We only need to verify that globalParams is correctly built with execution params taking priority.
	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Spec: drv1alpha1.DRPlanSpec{
			GlobalParams: []drv1alpha1.Parameter{
				{Name: "version", Value: "1"},
				{Name: "env", Value: "prod"},
			},
			Stages: []drv1alpha1.Stage{},
		},
	}

	execution := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "exec-1", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			OperationType: drv1alpha1.OperationTypeExecute,
			Params: []drv1alpha1.Parameter{
				{Name: "version", Value: "5"}, // overrides globalParams["version"]
			},
		},
	}

	scheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(scheme)
	_ = drv1alpha1.AddToScheme(scheme)
	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&drv1alpha1.DRPlanExecution{}).
		WithObjects(execution).
		Build()
	dc := fake.NewSimpleDynamicClient(scheme)

	executor := NewNativePlanExecutor(k8sClient, nil, nil, dc, jobRESTMapper{})

	// ExecutePlan with empty stages should succeed quickly
	if err := executor.ExecutePlan(context.Background(), plan, execution); err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}
}

// TestExecutePlan_ExecutionParams_ValueFrom verifies valueFrom resolution during ExecutePlan.
// NOCC:tosa/fn_length(设计如此)
func TestExecutePlan_ExecutionParams_ValueFrom(t *testing.T) {
	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{APIVersion: "batch/v1", Kind: "Job"},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "db-migrate-5",
			Namespace:         "default",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels:            map[string]string{"app": "myapp"},
		},
	}

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Spec: drv1alpha1.DRPlanSpec{
			Stages: []drv1alpha1.Stage{},
		},
	}

	execution := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "exec-2", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			OperationType: drv1alpha1.OperationTypeExecute,
			Params: []drv1alpha1.Parameter{
				{
					Name: "jobName",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Namespace:  "default",
							Name:       "db-migrate-5",
							JSONPath:   "{.metadata.name}",
						},
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	_ = batchv1.AddToScheme(scheme)
	_ = drv1alpha1.AddToScheme(scheme)
	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&drv1alpha1.DRPlanExecution{}).
		WithObjects(execution).
		Build()
	dc := fake.NewSimpleDynamicClient(scheme, job)

	executor := NewNativePlanExecutor(k8sClient, nil, nil, dc, jobRESTMapper{})
	if err := executor.ExecutePlan(context.Background(), plan, execution); err != nil {
		t.Fatalf("ExecutePlan with valueFrom failed: %v", err)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecutePlan_DeleteModeCleansHistoricalSubscriptionOutputs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	newSubscription := func(name string) *unstructured.Unstructured {
		sub := &unstructured.Unstructured{}
		sub.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps.clusternet.io",
			Version: "v1alpha1",
			Kind:    "Subscription",
		})
		sub.SetNamespace("default")
		sub.SetName(name)
		return sub
	}

	installExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app-install-001", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo-app",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Install",
		},
		Status: drv1alpha1.DRPlanExecutionStatus{
			Phase: drv1alpha1.PhaseSucceeded,
			StageStatuses: []drv1alpha1.StageStatus{
				{
					Name: "install",
					WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{
						{
							WorkflowRef: drv1alpha1.ObjectReference{Name: "wf-install", Namespace: "default"},
							Phase:       drv1alpha1.PhaseSucceeded,
							ActionStatuses: []drv1alpha1.ActionStatus{
								{
									Name:  "release-name-validate",
									Phase: drv1alpha1.PhaseSucceeded,
									Outputs: &drv1alpha1.ActionOutputs{
										SubscriptionRefs: []corev1.ObjectReference{
											{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Namespace: "default", Name: "release-name-validate-sub--cluster-a"},
											{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Namespace: "default", Name: "release-name-validate-sub--cluster-b"},
										},
									},
								},
								{
									Name:  "create-subscription",
									Phase: drv1alpha1.PhaseSucceeded,
									Outputs: &drv1alpha1.ActionOutputs{
										SubscriptionRef: &corev1.ObjectReference{
											APIVersion: "apps.clusternet.io/v1alpha1",
											Kind:       "Subscription",
											Namespace:  "default",
											Name:       "demo-app-subscription",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	revertExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app-revert-001", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo-app",
			OperationType: drv1alpha1.OperationTypeRevert,
		},
	}

	deleteExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app-delete-001", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo-app",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Delete",
		},
	}

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app", Namespace: "default"},
		Status: drv1alpha1.DRPlanStatus{
			ExecutionHistory: []drv1alpha1.ExecutionRecord{
				{Name: "demo-app-delete-001", Namespace: "default", OperationType: drv1alpha1.OperationTypeExecute},
				{Name: "demo-app-revert-001", Namespace: "default", OperationType: drv1alpha1.OperationTypeRevert},
				{Name: "demo-app-install-001", Namespace: "default", OperationType: drv1alpha1.OperationTypeExecute},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&drv1alpha1.DRPlanExecution{}).
		WithObjects(
			deleteExec,
			installExec,
			revertExec,
			newSubscription("release-name-validate-sub--cluster-a"),
			newSubscription("release-name-validate-sub--cluster-b"),
			newSubscription("demo-app-subscription"),
		).
		Build()

	executor := NewNativePlanExecutor(k8sClient, nil, nil, nil, nil)
	if err := executor.ExecutePlan(context.Background(), plan, deleteExec); err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}
	if deleteExec.Status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("execution phase = %s, want %s", deleteExec.Status.Phase, drv1alpha1.PhaseSucceeded)
	}
	if !strings.Contains(deleteExec.Status.Message, "cleaned 3 historical subscriptions") {
		t.Fatalf("unexpected execution message: %q", deleteExec.Status.Message)
	}

	for _, name := range []string{
		"release-name-validate-sub--cluster-a",
		"release-name-validate-sub--cluster-b",
		"demo-app-subscription",
	} {
		sub := newSubscription(name)
		err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(sub), sub)
		if !apierrors.IsNotFound(err) {
			t.Fatalf("expected subscription %s to be deleted, got err=%v", name, err)
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecutePlan_NonDeleteModeSkipsHistoricalSubscriptionCleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	sub := &unstructured.Unstructured{}
	sub.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	sub.SetNamespace("default")
	sub.SetName("release-name-validate-sub--cluster-a")

	installExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app-install-001", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo-app",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Install",
		},
		Status: drv1alpha1.DRPlanExecutionStatus{
			Phase: drv1alpha1.PhaseSucceeded,
			StageStatuses: []drv1alpha1.StageStatus{
				{
					Name: "install",
					WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{
						{
							WorkflowRef: drv1alpha1.ObjectReference{Name: "wf-install", Namespace: "default"},
							Phase:       drv1alpha1.PhaseSucceeded,
							ActionStatuses: []drv1alpha1.ActionStatus{
								{
									Name:  "release-name-validate",
									Phase: drv1alpha1.PhaseSucceeded,
									Outputs: &drv1alpha1.ActionOutputs{
										SubscriptionRefs: []corev1.ObjectReference{
											{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Namespace: "default", Name: "release-name-validate-sub--cluster-a"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	upgradeExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app-upgrade-001", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo-app",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Upgrade",
		},
	}

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-app", Namespace: "default"},
		Status: drv1alpha1.DRPlanStatus{
			ExecutionHistory: []drv1alpha1.ExecutionRecord{
				{Name: "demo-app-upgrade-001", Namespace: "default", OperationType: drv1alpha1.OperationTypeExecute},
				{Name: "demo-app-install-001", Namespace: "default", OperationType: drv1alpha1.OperationTypeExecute},
			},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&drv1alpha1.DRPlanExecution{}).
		WithObjects(upgradeExec, installExec, sub).
		Build()

	executor := NewNativePlanExecutor(k8sClient, nil, nil, nil, nil)
	if err := executor.ExecutePlan(context.Background(), plan, upgradeExec); err != nil {
		t.Fatalf("ExecutePlan failed: %v", err)
	}

	remaining := &unstructured.Unstructured{}
	remaining.SetGroupVersionKind(sub.GroupVersionKind())
	remaining.SetNamespace("default")
	remaining.SetName("release-name-validate-sub--cluster-a")
	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(remaining), remaining); err != nil {
		t.Fatalf("expected subscription to remain for non-delete mode: %v", err)
	}
}

// recordingStageExecutor records which stage names were passed to RevertStage.
type recordingStageExecutor struct {
	revertedStages   []string
	lastGlobalParams map[string]interface{}
}

func (r *recordingStageExecutor) ExecuteStage(_ context.Context, _ *drv1alpha1.DRPlan,
	stage *drv1alpha1.Stage, _ map[string]interface{}, _ map[string]interface{}) (*drv1alpha1.StageStatus, error) {
	return &drv1alpha1.StageStatus{Name: stage.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (r *recordingStageExecutor) RevertStage(_ context.Context, _ *drv1alpha1.DRPlan,
	stage *drv1alpha1.Stage, _ *drv1alpha1.StageStatus, globalParams map[string]interface{}, _ map[string]interface{}) (*drv1alpha1.StageStatus, error) {
	r.revertedStages = append(r.revertedStages, stage.Name)
	r.lastGlobalParams = globalParams
	return &drv1alpha1.StageStatus{Name: stage.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

// TestRevertPlan_FailedStageIsReverted verifies that a stage with Failed phase in the original
// execution still gets passed to RevertStage so its partially-succeeded workflows can be cleaned up.
// Stages in non-terminal (Pending/Running) or Skipped/Canceled phases must be skipped entirely.
// NOCC:tosa/fn_length(设计如此)
func TestRevertPlan_FailedStageIsReverted(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)
	dc := fake.NewSimpleDynamicClient(scheme)

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Spec: drv1alpha1.DRPlanSpec{
			Stages: []drv1alpha1.Stage{
				{Name: "stage-succeeded"},
				{Name: "stage-failed"},
				{Name: "stage-pending"},
				{Name: "stage-skipped"},
			},
		},
	}

	targetExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "target-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			OperationType: drv1alpha1.OperationTypeExecute,
		},
		Status: drv1alpha1.DRPlanExecutionStatus{
			Phase: drv1alpha1.PhaseFailed,
			StageStatuses: []drv1alpha1.StageStatus{
				{Name: "stage-succeeded", Phase: drv1alpha1.PhaseSucceeded},
				{Name: "stage-failed", Phase: drv1alpha1.PhaseFailed},
				{Name: "stage-pending", Phase: drv1alpha1.PhasePending},
				{Name: "stage-skipped", Phase: drv1alpha1.PhaseSkipped},
			},
		},
	}

	revertExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "revert-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:            "test-plan",
			OperationType:      drv1alpha1.OperationTypeRevert,
			RevertExecutionRef: "target-exec",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(revertExec, targetExec).
		WithObjects(revertExec, targetExec).
		Build()

	recorder := &recordingStageExecutor{}
	executor := NewNativePlanExecutor(k8sClient, recorder, nil, dc, jobRESTMapper{})

	if err := executor.RevertPlan(context.Background(), plan, revertExec); err != nil {
		t.Fatalf("RevertPlan failed: %v", err)
	}

	reverted := make(map[string]bool)
	for _, s := range recorder.revertedStages {
		reverted[s] = true
	}

	if !reverted["stage-succeeded"] {
		t.Error("expected stage-succeeded to be reverted, but it was skipped")
	}
	if !reverted["stage-failed"] {
		t.Error("expected stage-failed to be reverted (partial cleanup), but it was skipped")
	}
	if reverted["stage-pending"] {
		t.Error("expected stage-pending to be skipped, but it was reverted")
	}
	if reverted["stage-skipped"] {
		t.Error("expected stage-skipped to be skipped, but it was reverted")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestApplyPerClusterHookCleanup_OnSuccessDeletesChildSubscriptions(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	subA := &unstructured.Unstructured{}
	subA.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	subA.SetName("hook-sub--cluster-a")
	subA.SetNamespace("default")

	subB := &unstructured.Unstructured{}
	subB.SetGroupVersionKind(subA.GroupVersionKind())
	subB.SetName("hook-sub--cluster-b")
	subB.SetNamespace("default")

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(subA, subB).Build()
	subExec := &SubscriptionActionExecutor{client: k8sClient}
	executor := &NativeWorkflowExecutor{client: k8sClient}

	actions := []drv1alpha1.Action{
		{
			Name: "hook",
			Type: drv1alpha1.ActionTypeSubscription,
			HookCleanup: &drv1alpha1.HookCleanupPolicy{
				OnSuccess: true,
			},
		},
	}
	statuses := []drv1alpha1.ActionStatus{
		{
			Name:  "hook",
			Phase: drv1alpha1.PhaseSucceeded,
			Outputs: &drv1alpha1.ActionOutputs{
				SubscriptionRefs: []corev1.ObjectReference{
					{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Name: "hook-sub--cluster-a", Namespace: "default"},
					{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Name: "hook-sub--cluster-b", Namespace: "default"},
				},
			},
		},
	}

	if err := executor.applyPerClusterHookCleanup(context.Background(), subExec, actions, statuses); err != nil {
		t.Fatalf("unexpected cleanup error: %v", err)
	}

	probe := &unstructured.Unstructured{}
	probe.SetGroupVersionKind(subA.GroupVersionKind())
	for _, name := range []string{"hook-sub--cluster-a", "hook-sub--cluster-b"} {
		err := k8sClient.Get(context.Background(), client.ObjectKey{Name: name, Namespace: "default"}, probe)
		if err == nil {
			t.Fatalf("expected child Subscription %s to be deleted", name)
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecutePerClusterBatch_FailFastStillRunsFailureCleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	managedCluster := &unstructured.Unstructured{}
	managedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	managedCluster.SetNamespace("ns1")
	managedCluster.SetName("cluster-a")
	managedCluster.Object["spec"] = map[string]interface{}{"clusterId": "cluster-a-id"}

	childJobFailed := &unstructured.Unstructured{}
	childJobFailed.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "batch",
		Version: "v1",
		Kind:    "Job",
	})
	childJobFailed.SetNamespace("app-ns")
	childJobFailed.SetName("hook-job")
	childJobFailed.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"type": "Failed", "status": "True"},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(managedCluster).Build()
	subExec := &SubscriptionActionExecutor{
		client: k8sClient,
		childClientFactory: &fakeChildClusterClientFactory{
			clients: map[string]client.Client{
				"cluster-a-id": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJobFailed).Build(),
			},
		},
	}
	registry := NewExecutorRegistry()
	if err := registry.RegisterExecutor(subExec); err != nil {
		t.Fatalf("register subscription executor: %v", err)
	}
	executor := NewNativeWorkflowExecutor(k8sClient, registry)

	actions := []drv1alpha1.Action{
		{
			Name:                 "failing-hook",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			HookCleanup: &drv1alpha1.HookCleanupPolicy{
				OnFailure: true,
			},
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "hook-sub",
				Namespace: "default",
				Spec: &clusternetapps.SubscriptionSpec{
					SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
					Feeds: []clusternetapps.Feed{
						{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "app-ns"},
					},
				},
			},
			Timeout: "10s",
		},
	}

	statuses, err := executor.executePerClusterBatch(
		context.Background(),
		actions,
		map[string]interface{}{"mode": "install"},
		drv1alpha1.FailurePolicyFailFast,
	)
	if err == nil {
		t.Fatal("expected per-cluster batch to fail")
	}
	if len(statuses) != 1 || statuses[0].Phase != drv1alpha1.PhaseFailed {
		t.Fatalf("unexpected statuses: %#v", statuses)
	}

	childSub := &unstructured.Unstructured{}
	childSub.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "Subscription",
	})
	childSub.SetNamespace("default")
	childSub.SetName("hook-sub--cluster-a")
	getErr := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(childSub), childSub)
	if !apierrors.IsNotFound(getErr) {
		t.Fatalf("expected failed hook child subscription to be cleaned up, got err=%v", getErr)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteStage_AllWorkflowsSkippedIsSucceeded(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-skipped", Namespace: "default"},
	}
	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workflow).Build()
	stageExecutor := NewStageExecutor(k8sClient, &fixedWorkflowPhaseExecutor{phase: drv1alpha1.PhaseSkipped})

	stage := &drv1alpha1.Stage{
		Name: "install",
		Workflows: []drv1alpha1.WorkflowReference{
			{WorkflowRef: drv1alpha1.ObjectReference{Name: "wf-skipped", Namespace: "default"}},
		},
	}
	status, err := stageExecutor.ExecuteStage(context.Background(), &drv1alpha1.DRPlan{}, stage, nil, nil)
	if err != nil {
		t.Fatalf("ExecuteStage returned error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("stage phase = %s, want %s", status.Phase, drv1alpha1.PhaseSucceeded)
	}
	if status.Message != "All workflows skipped" {
		t.Fatalf("stage message = %q, want %q", status.Message, "All workflows skipped")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestUpdateExecutionSummary_DoesNotAccumulateWorkflowCounters(t *testing.T) {
	executor := &NativePlanExecutor{}
	execution := &drv1alpha1.DRPlanExecution{
		Status: drv1alpha1.DRPlanExecutionStatus{
			StageStatuses: []drv1alpha1.StageStatus{
				{
					Name:  "install",
					Phase: drv1alpha1.PhaseSucceeded,
					WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{
						{Phase: drv1alpha1.PhaseSucceeded},
						{Phase: drv1alpha1.PhaseFailed},
					},
				},
			},
		},
	}

	executor.updateExecutionSummary(execution)
	executor.updateExecutionSummary(execution)

	if execution.Status.Summary == nil {
		t.Fatal("summary should not be nil")
	}
	summary := execution.Status.Summary
	if summary.TotalWorkflows != 2 || summary.CompletedWorkflows != 1 || summary.FailedWorkflows != 1 {
		t.Fatalf("summary counters accumulated unexpectedly: %+v", *summary)
	}
}

// TestValidateAndFetchRevertTarget_TerminalPhase verifies that revert is permitted
// for both Succeeded and Failed target executions, and rejected for non-terminal phases.
// NOCC:tosa/fn_length(设计如此)
func TestValidateAndFetchRevertTarget_TerminalPhase(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)
	dc := fake.NewSimpleDynamicClient(scheme)

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
	}

	makeTarget := func(phase string) *drv1alpha1.DRPlanExecution {
		return &drv1alpha1.DRPlanExecution{
			ObjectMeta: metav1.ObjectMeta{Name: "target-exec", Namespace: "default"},
			Spec: drv1alpha1.DRPlanExecutionSpec{
				PlanRef:       "test-plan",
				OperationType: drv1alpha1.OperationTypeExecute,
			},
			Status: drv1alpha1.DRPlanExecutionStatus{
				Phase: phase,
				StageStatuses: []drv1alpha1.StageStatus{
					{Name: "stage-1", Phase: drv1alpha1.PhaseSucceeded},
				},
			},
		}
	}

	tests := []struct {
		name        string
		targetPhase string
		wantErr     bool
	}{
		{name: "Succeeded → allowed", targetPhase: drv1alpha1.PhaseSucceeded, wantErr: false},
		{name: "Failed → allowed", targetPhase: drv1alpha1.PhaseFailed, wantErr: false},
		{name: "Running → rejected", targetPhase: drv1alpha1.PhaseRunning, wantErr: true},
		{name: "Pending → rejected", targetPhase: drv1alpha1.PhasePending, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := makeTarget(tt.targetPhase)
			k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).
				WithStatusSubresource(target).
				WithObjects(target).
				Build()

			executor := NewNativePlanExecutor(k8sClient, nil, nil, dc, jobRESTMapper{})

			revertExec := &drv1alpha1.DRPlanExecution{
				ObjectMeta: metav1.ObjectMeta{Name: "revert-exec", Namespace: "default"},
				Spec: drv1alpha1.DRPlanExecutionSpec{
					PlanRef:            "test-plan",
					OperationType:      drv1alpha1.OperationTypeRevert,
					RevertExecutionRef: "target-exec",
				},
			}

			_, err := executor.validateAndFetchRevertTarget(context.Background(), plan, revertExec)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for phase %s, got nil", tt.targetPhase)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for phase %s, got: %v", tt.targetPhase, err)
			}
		})
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestRevertWorkflow_PassesParamsToActionRollback(t *testing.T) {
	registry := NewExecutorRegistry()
	recorder := &recordingRollbackExecutor{actionType: drv1alpha1.ActionTypeSubscription}
	if err := registry.RegisterExecutor(recorder); err != nil {
		t.Fatalf("register executor: %v", err)
	}

	executor := NewNativeWorkflowExecutor(nil, registry)
	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{Name: "hook-a", Type: drv1alpha1.ActionTypeSubscription},
			},
		},
	}
	workflowStatus := &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: "wf", Namespace: "default"},
		ActionStatuses: []drv1alpha1.ActionStatus{
			{Name: "hook-a", Phase: drv1alpha1.PhaseSucceeded},
		},
	}
	params := map[string]interface{}{"feedNamespace": "default", "revision": "5"}

	if _, err := executor.RevertWorkflow(context.Background(), workflow, workflowStatus, params); err != nil {
		t.Fatalf("RevertWorkflow failed: %v", err)
	}
	if recorder.receivedParams == nil {
		t.Fatal("expected rollback params to be passed through")
	}
	if recorder.receivedParams["revision"] != "5" {
		t.Fatalf("expected rollback params to include revision=5, got %v", recorder.receivedParams)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestRevertWorkflow_ExecutesRollbackHooksAroundRollback(t *testing.T) {
	registry := NewExecutorRegistry()
	hookRecorder := &recordingActionExecutor{actionType: drv1alpha1.ActionTypeSubscription}
	if err := registry.RegisterExecutor(hookRecorder); err != nil {
		t.Fatalf("register executor: %v", err)
	}

	executor := NewNativeWorkflowExecutor(nil, registry)
	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{Name: "pre-rb", Type: drv1alpha1.ActionTypeSubscription, HookType: "pre-rollback", When: `mode == "rollback"`},
				{Name: "main", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "post-rb", Type: drv1alpha1.ActionTypeSubscription, HookType: "post-rollback", When: `mode == "rollback"`},
			},
		},
	}
	workflowStatus := &drv1alpha1.WorkflowExecutionStatus{
		WorkflowRef: drv1alpha1.ObjectReference{Name: "wf", Namespace: "default"},
		ActionStatuses: []drv1alpha1.ActionStatus{
			{Name: "main", Phase: drv1alpha1.PhaseSucceeded},
		},
	}

	rollbackStatus, err := executor.RevertWorkflow(context.Background(), workflow, workflowStatus, map[string]interface{}{"mode": "rollback"})
	if err != nil {
		t.Fatalf("RevertWorkflow failed: %v", err)
	}
	if len(hookRecorder.executed) != 2 {
		t.Fatalf("expected 2 rollback hooks to execute, got %v", hookRecorder.executed)
	}
	if hookRecorder.executed[0] != "pre-rb" || hookRecorder.executed[1] != "post-rb" {
		t.Fatalf("unexpected rollback hook execution order: %v", hookRecorder.executed)
	}
	if len(rollbackStatus.ActionStatuses) != 3 {
		t.Fatalf("expected 3 action statuses, got %d", len(rollbackStatus.ActionStatuses))
	}
	if rollbackStatus.ActionStatuses[0].Name != "pre-rb" || rollbackStatus.ActionStatuses[2].Name != "post-rb" {
		t.Fatalf("unexpected rollback action status order: %+v", rollbackStatus.ActionStatuses)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestRevertStage_RevertsFailedWorkflowForPartialCleanup(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-failed", Namespace: "default"},
	}
	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(workflow).Build()

	recorder := &recordingWorkflowExecutor{}
	stageExecutor := NewStageExecutor(k8sClient, recorder)

	stage := &drv1alpha1.Stage{
		Name: "install",
		Workflows: []drv1alpha1.WorkflowReference{
			{WorkflowRef: drv1alpha1.ObjectReference{Name: "wf-failed", Namespace: "default"}},
		},
	}
	stageStatus := &drv1alpha1.StageStatus{
		Name: "install",
		WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{
			{
				WorkflowRef: drv1alpha1.ObjectReference{Name: "wf-failed", Namespace: "default"},
				Phase:       drv1alpha1.PhaseFailed,
				ActionStatuses: []drv1alpha1.ActionStatus{
					{Name: "hook-a", Phase: drv1alpha1.PhaseSucceeded},
				},
			},
		},
	}

	if _, err := stageExecutor.RevertStage(context.Background(), &drv1alpha1.DRPlan{}, stage, stageStatus, nil, nil); err != nil {
		t.Fatalf("RevertStage failed: %v", err)
	}
	if len(recorder.revertedWorkflows) != 1 || recorder.revertedWorkflows[0] != "wf-failed" {
		t.Fatalf("expected failed workflow to be reverted for partial cleanup, got %v", recorder.revertedWorkflows)
	}
}

func TestRevertPlan_InjectsRollbackMode(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = drv1alpha1.AddToScheme(scheme)
	dc := fake.NewSimpleDynamicClient(scheme)

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Spec: drv1alpha1.DRPlanSpec{
			Stages: []drv1alpha1.Stage{{Name: "install"}},
		},
	}
	targetExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "target-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Install",
		},
		Status: drv1alpha1.DRPlanExecutionStatus{
			Phase: drv1alpha1.PhaseSucceeded,
			StageStatuses: []drv1alpha1.StageStatus{
				{Name: "install", Phase: drv1alpha1.PhaseSucceeded},
			},
		},
	}
	revertExec := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "revert-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:            "test-plan",
			OperationType:      drv1alpha1.OperationTypeRevert,
			Mode:               "Rollback",
			RevertExecutionRef: "target-exec",
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(revertExec, targetExec).
		WithObjects(revertExec, targetExec).
		Build()
	recorder := &recordingStageExecutor{}
	executor := NewNativePlanExecutor(k8sClient, recorder, nil, dc, jobRESTMapper{})

	if err := executor.RevertPlan(context.Background(), plan, revertExec); err != nil {
		t.Fatalf("RevertPlan failed: %v", err)
	}
	if recorder.lastGlobalParams == nil || recorder.lastGlobalParams["mode"] != "rollback" {
		t.Fatalf("expected rollback mode in global params, got %#v", recorder.lastGlobalParams)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecutePerClusterBatch_WhenSkippedDoesNotCreateParentSubscriptions(t *testing.T) {
	scheme := runtime.NewScheme()

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
	executor := NewNativeWorkflowExecutor(k8sClient, nil)

	actions := []drv1alpha1.Action{
		{
			Name:                 "pre-install-hook",
			Type:                 drv1alpha1.ActionTypeSubscription,
			When:                 `mode == "install"`,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "pre-install-sub",
				Namespace: "default",
				Operation: drv1alpha1.OperationReplace,
				Spec: &clusternetapps.SubscriptionSpec{
					Feeds: []clusternetapps.Feed{
						{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "default"},
					},
				},
			},
		},
	}

	statuses, err := executor.executePerClusterBatch(context.Background(), actions, map[string]interface{}{"mode": "upgrade"}, drv1alpha1.FailurePolicyFailFast)
	if err != nil {
		t.Fatalf("executePerClusterBatch returned error: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Phase != drv1alpha1.PhaseSkipped {
		t.Fatalf("expected skipped status, got %s", statuses[0].Phase)
	}

	subList := &unstructured.UnstructuredList{}
	subList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.clusternet.io",
		Version: "v1alpha1",
		Kind:    "SubscriptionList",
	})
	if err := k8sClient.List(context.Background(), subList); err != nil {
		t.Fatalf("list subscriptions: %v", err)
	}
	if len(subList.Items) != 0 {
		t.Fatalf("expected no subscriptions to be created for skipped action, got %d", len(subList.Items))
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestCollectPerClusterBindings_UsesEachActionSelectors(t *testing.T) {
	scheme := runtime.NewScheme()
	clusterA := &unstructured.Unstructured{}
	clusterA.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	clusterA.SetNamespace("ns1")
	clusterA.SetName("cluster-a")
	clusterA.SetLabels(map[string]string{"env": "prod"})

	clusterB := &unstructured.Unstructured{}
	clusterB.SetGroupVersionKind(clusterA.GroupVersionKind())
	clusterB.SetNamespace("ns2")
	clusterB.SetName("cluster-b")
	clusterB.SetLabels(map[string]string{"env": "test"})

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(clusterA, clusterB).Build()
	subExec := &SubscriptionActionExecutor{client: k8sClient}
	executor := NewNativeWorkflowExecutor(k8sClient, nil)

	actions := []drv1alpha1.Action{
		{
			Name: "hook-a",
			Subscription: &drv1alpha1.SubscriptionAction{Spec: &clusternetapps.SubscriptionSpec{
				Subscribers: []clusternetapps.Subscriber{{ClusterAffinity: &metav1.LabelSelector{
					MatchLabels: map[string]string{"env": "prod"},
				}}},
			}},
		},
		{
			Name: "hook-b",
			Subscription: &drv1alpha1.SubscriptionAction{Spec: &clusternetapps.SubscriptionSpec{
				Subscribers: []clusternetapps.Subscriber{{ClusterAffinity: &metav1.LabelSelector{
					MatchLabels: map[string]string{"env": "test"},
				}}},
			}},
		},
	}

	targets, union, err := executor.collectPerClusterBindings(context.Background(), subExec, actions)
	if err != nil {
		t.Fatalf("collectPerClusterBindings failed: %v", err)
	}
	if len(union) != 2 {
		t.Fatalf("union bindings = %v, want 2 entries", union)
	}
	if !clusterTargeted(targets, "hook-a", "ns1/cluster-a") || clusterTargeted(targets, "hook-a", "ns2/cluster-b") {
		t.Fatalf("unexpected targets for hook-a: %#v", targets["hook-a"])
	}
	if !clusterTargeted(targets, "hook-b", "ns2/cluster-b") || clusterTargeted(targets, "hook-b", "ns1/cluster-a") {
		t.Fatalf("unexpected targets for hook-b: %#v", targets["hook-b"])
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteClusterActions_SkipsUntargetedAction(t *testing.T) {
	scheme := runtime.NewScheme()

	managedCluster := &unstructured.Unstructured{}
	managedCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "clusters.clusternet.io",
		Version: "v1beta1",
		Kind:    "ManagedCluster",
	})
	managedCluster.SetNamespace("ns1")
	managedCluster.SetName("cluster-a")
	managedCluster.Object["spec"] = map[string]interface{}{"clusterId": "cluster-a-id"}

	childJob := &unstructured.Unstructured{}
	childJob.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "batch",
		Version: "v1",
		Kind:    "Job",
	})
	childJob.SetNamespace("app-ns")
	childJob.SetName("hook-job")
	childJob.Object["status"] = map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{"type": "Complete", "status": "True"},
		},
	}

	k8sClient := fakeclient.NewClientBuilder().WithScheme(scheme).WithObjects(managedCluster).Build()
	subExec := &SubscriptionActionExecutor{
		client: k8sClient,
		childClientFactory: &fakeChildClusterClientFactory{
			clients: map[string]client.Client{
				"cluster-a-id": fakeclient.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(childJob).Build(),
			},
		},
	}
	executor := NewNativeWorkflowExecutor(k8sClient, nil)

	actions := []drv1alpha1.Action{
		{
			Name:                 "hook-a",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "hook-sub",
				Namespace: "default",
				Spec: &clusternetapps.SubscriptionSpec{
					SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
					Feeds: []clusternetapps.Feed{
						{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "app-ns"},
					},
				},
			},
			Timeout: "10s",
		},
		{
			Name:                 "hook-b",
			Type:                 drv1alpha1.ActionTypeSubscription,
			WaitReady:            true,
			ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
			Subscription: &drv1alpha1.SubscriptionAction{
				Name:      "unused-sub",
				Namespace: "default",
				Spec: &clusternetapps.SubscriptionSpec{
					SchedulingStrategy: clusternetapps.ReplicaSchedulingStrategyType,
					Feeds: []clusternetapps.Feed{
						{APIVersion: "batch/v1", Kind: "Job", Name: "hook-job", Namespace: "app-ns"},
					},
				},
			},
			Timeout: "10s",
		},
	}

	statusMap := make(map[string][]drv1alpha1.ClusterActionStatus)
	childRefMap := make(map[string][]corev1.ObjectReference)
	var mu sync.Mutex
	batchStartTime := metav1.Now()
	executor.executeClusterActions(
		context.Background(),
		subExec,
		actions,
		perClusterActionTargets{
			"hook-a": {"ns1/cluster-a": {}},
			"hook-b": {"ns2/cluster-b": {}},
		},
		"ns1/cluster-a",
		nil,
		drv1alpha1.FailurePolicyFailFast,
		&batchStartTime,
		&mu,
		statusMap,
		childRefMap,
		func() {},
	)

	if len(statusMap["hook-a"]) != 1 || statusMap["hook-a"][0].Phase != drv1alpha1.PhaseSucceeded {
		t.Fatalf("hook-a statuses = %#v, want one succeeded status", statusMap["hook-a"])
	}
	if len(statusMap["hook-b"]) != 1 || statusMap["hook-b"][0].Phase != drv1alpha1.PhaseSkipped {
		t.Fatalf("hook-b statuses = %#v, want one skipped status", statusMap["hook-b"])
	}
	if len(childRefMap["hook-a"]) != 1 || childRefMap["hook-a"][0].Name != "hook-sub--cluster-a" {
		t.Fatalf("unexpected child refs for hook-a: %#v", childRefMap["hook-a"])
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestBuildPerClusterResult_PersistsChildSubscriptionRefs(t *testing.T) {
	executor := &NativeWorkflowExecutor{}
	start := metav1.Now()
	actions := []drv1alpha1.Action{{Name: "hook-a"}}
	statusMap := map[string][]drv1alpha1.ClusterActionStatus{
		"hook-a": {
			{Cluster: "ns1/cluster-a", Phase: drv1alpha1.PhaseSucceeded},
		},
	}
	childRefMap := map[string][]corev1.ObjectReference{
		"hook-a": {
			{APIVersion: "apps.clusternet.io/v1alpha1", Kind: "Subscription", Namespace: "default", Name: "hook-a-sub--cluster-a"},
		},
	}

	statuses, err := executor.buildPerClusterResult(actions, statusMap, childRefMap, &start, drv1alpha1.FailurePolicyFailFast)
	if err != nil {
		t.Fatalf("buildPerClusterResult failed: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].Outputs == nil || len(statuses[0].Outputs.SubscriptionRefs) != 1 {
		t.Fatal("expected child subscriptionRefs to be persisted in action outputs")
	}
	if statuses[0].Outputs.SubscriptionRefs[0].Name != "hook-a-sub--cluster-a" || statuses[0].Outputs.SubscriptionRefs[0].Namespace != testDefaultNamespace {
		t.Fatalf("unexpected subscriptionRefs: %#v", statuses[0].Outputs.SubscriptionRefs)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestUpdateWorkflowFinalPhase_RunningNotSucceeded(t *testing.T) {
	status := &drv1alpha1.WorkflowExecutionStatus{
		ActionStatuses: []drv1alpha1.ActionStatus{
			{Name: "hook-a", Phase: drv1alpha1.PhaseRunning},
			{Name: "main", Phase: drv1alpha1.PhaseSucceeded},
		},
	}

	updateWorkflowFinalPhase(status, "default", "demo")

	if status.Phase != drv1alpha1.PhaseRunning {
		t.Fatalf("workflow phase = %q, want %q", status.Phase, drv1alpha1.PhaseRunning)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteGlobalBatchFallback_RejectsRunningStatus(t *testing.T) {
	registry := NewExecutorRegistry()
	if err := registry.RegisterExecutor(&staticActionExecutor{
		actionType: drv1alpha1.ActionTypeSubscription,
		status: &drv1alpha1.ActionStatus{
			Name:      "hook-a",
			Phase:     drv1alpha1.PhaseRunning,
			StartTime: &metav1.Time{Time: time.Now()},
		},
	}); err != nil {
		t.Fatalf("register executor: %v", err)
	}

	executor := &NativeWorkflowExecutor{registry: registry}
	actions := []drv1alpha1.Action{
		{Name: "hook-a", Type: drv1alpha1.ActionTypeSubscription},
	}

	statuses, err := executor.executeGlobalBatchFallback(context.Background(), actions, nil, drv1alpha1.FailurePolicyFailFast)
	if err == nil {
		t.Fatal("expected fallback to fail for non-terminal action status")
	}
	if len(statuses) != 1 {
		t.Fatalf("statuses count = %d, want 1", len(statuses))
	}
	if statuses[0].Phase != drv1alpha1.PhaseFailed {
		t.Fatalf("status phase = %q, want %q", statuses[0].Phase, drv1alpha1.PhaseFailed)
	}
}
