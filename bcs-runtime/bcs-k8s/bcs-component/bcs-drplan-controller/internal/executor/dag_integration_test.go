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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// fakeActionExecutor records execution timing for concurrency assertions.
type fakeActionExecutor struct {
	delay     time.Duration
	failNames map[string]bool

	mu          sync.Mutex
	execOrder   []string
	concurrency int32
	maxConc     int32
}

func (f *fakeActionExecutor) Execute(_ context.Context, action *drv1alpha1.Action, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	cur := atomic.AddInt32(&f.concurrency, 1)
	f.mu.Lock()
	if cur > f.maxConc {
		f.maxConc = cur
	}
	f.mu.Unlock()

	if f.delay > 0 {
		time.Sleep(f.delay)
	}

	atomic.AddInt32(&f.concurrency, -1)

	f.mu.Lock()
	f.execOrder = append(f.execOrder, action.Name)
	f.mu.Unlock()

	phase := drv1alpha1.PhaseSucceeded
	if f.failNames != nil && f.failNames[action.Name] {
		phase = drv1alpha1.PhaseFailed
	}

	return &drv1alpha1.ActionStatus{
		Name:           action.Name,
		Phase:          phase,
		StartTime:      &metav1.Time{Time: time.Now()},
		CompletionTime: &metav1.Time{Time: time.Now()},
	}, nil
}

func (f *fakeActionExecutor) Rollback(_ context.Context, action *drv1alpha1.Action, _ *drv1alpha1.ActionStatus, _ map[string]interface{}) (*drv1alpha1.ActionStatus, error) {
	return &drv1alpha1.ActionStatus{Name: action.Name, Phase: drv1alpha1.PhaseSucceeded}, nil
}

func (f *fakeActionExecutor) Type() string { return drv1alpha1.ActionTypeSubscription }

func newTestExecutor(delay time.Duration, failNames map[string]bool) (*NativeWorkflowExecutor, *fakeActionExecutor) {
	fake := &fakeActionExecutor{delay: delay, failNames: failNames}
	registry := NewExecutorRegistry()
	_ = registry.RegisterExecutor(fake)
	return NewNativeWorkflowExecutor(nil, registry), fake
}

func TestExecuteWorkflow_DAG_Parallel(t *testing.T) {
	exec, fake := newTestExecutor(50*time.Millisecond, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "main", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "post-a", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"main"}},
				{Name: "post-b", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"main"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}

	// ActionStatuses must follow original definition order regardless of finish order
	if len(status.ActionStatuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(status.ActionStatuses))
	}
	expected := []string{"main", "post-a", "post-b"}
	for i, name := range expected {
		if status.ActionStatuses[i].Name != name {
			t.Errorf("status[%d] expected %s, got %s", i, name, status.ActionStatuses[i].Name)
		}
	}

	// post-a and post-b should have run concurrently (maxConc >= 2)
	if fake.maxConc < 2 {
		t.Errorf("expected max concurrency >= 2 for parallel post actions, got %d", fake.maxConc)
	}
}

func TestExecuteWorkflow_DAG_Sequential(t *testing.T) {
	exec, fake := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "a", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "b", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"a"}},
				{Name: "c", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"b"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s", status.Phase)
	}

	// Linear chain: each should have run sequentially
	fake.mu.Lock()
	order := make([]string, len(fake.execOrder))
	copy(order, fake.execOrder)
	fake.mu.Unlock()

	expectedOrder := []string{"a", "b", "c"}
	for i, name := range expectedOrder {
		if i >= len(order) || order[i] != name {
			t.Errorf("execution order mismatch at %d: expected %s, got %v", i, name, order)
			break
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_DAG_CycleDetection(t *testing.T) {
	exec, _ := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{Name: "a", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"c"}},
				{Name: "b", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"a"}},
				{Name: "c", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"b"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err == nil {
		t.Fatal("expected error for cycle, got nil")
	}
	if status.Phase != drv1alpha1.PhaseFailed {
		t.Errorf("expected Failed phase, got %s", status.Phase)
	}
	if len(status.ActionStatuses) != 0 {
		t.Errorf("expected 0 action statuses (no action should execute), got %d", len(status.ActionStatuses))
	}
}

func TestExecuteWorkflow_DAG_UnknownRef(t *testing.T) {
	exec, _ := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			Actions: []drv1alpha1.Action{
				{Name: "a", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"nonexistent"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err == nil {
		t.Fatal("expected error for unknown ref, got nil")
	}
	if status.Phase != drv1alpha1.PhaseFailed {
		t.Errorf("expected Failed phase, got %s", status.Phase)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_DAG_FailFastCancelsParallel(t *testing.T) {
	exec, _ := newTestExecutor(100*time.Millisecond, map[string]bool{"post-a": true})

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "main", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "post-a", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"main"}},
				{Name: "post-b", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"main"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err == nil {
		t.Fatal("expected error when action fails in FailFast mode")
	}
	if status.Phase != drv1alpha1.PhaseFailed {
		t.Errorf("expected Failed, got %s", status.Phase)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_DAG_WhenSkipDoesNotBlock(t *testing.T) {
	exec, _ := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "pre-install", Type: drv1alpha1.ActionTypeSubscription, When: `mode == "install"`},
				{Name: "main", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"pre-install"}},
			},
		},
	}

	params := map[string]interface{}{"mode": "upgrade"}
	status, err := exec.ExecuteWorkflow(context.Background(), workflow, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if len(status.ActionStatuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(status.ActionStatuses))
	}
	if status.ActionStatuses[0].Phase != drv1alpha1.PhaseSkipped {
		t.Errorf("pre-install should be Skipped, got %s", status.ActionStatuses[0].Phase)
	}
	if status.ActionStatuses[1].Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("main should be Succeeded, got %s", status.ActionStatuses[1].Phase)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_NoDependsOn_BackwardCompatible(t *testing.T) {
	exec, fake := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "a", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "b", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "c", Type: drv1alpha1.ActionTypeSubscription},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s", status.Phase)
	}

	// Should run sequentially (backward compatible), never concurrent
	if fake.maxConc > 1 {
		t.Errorf("backward-compat path should be sequential, but maxConc=%d", fake.maxConc)
	}

	// Status order matches definition order
	for i, name := range []string{"a", "b", "c"} {
		if status.ActionStatuses[i].Name != name {
			t.Errorf("status[%d] expected %s, got %s", i, name, status.ActionStatuses[i].Name)
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_DAG_PerClusterInLayer(t *testing.T) {
	// PerCluster actions in DAG require a real SubscriptionActionExecutor with a k8s client.
	// Without one, executePerClusterBatch falls back to executeGlobalBatchFallback
	// (which calls actionExecutor.Execute like Global). Verify the fallback path works
	// and PerCluster actions are not silently skipped.
	exec, _ := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-mixed", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "main", Type: drv1alpha1.ActionTypeSubscription},
				{
					Name: "pc-hook", Type: drv1alpha1.ActionTypeSubscription,
					DependsOn:            []string{"main"},
					WaitReady:            true,
					ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster,
					Subscription:         &drv1alpha1.SubscriptionAction{Name: "pc-sub", Namespace: "default"},
				},
				{
					Name: "global-hook", Type: drv1alpha1.ActionTypeSubscription,
					DependsOn: []string{"main"},
				},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
	if len(status.ActionStatuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(status.ActionStatuses))
	}
	// Original definition order preserved
	expected := []string{"main", "pc-hook", "global-hook"}
	for i, name := range expected {
		if status.ActionStatuses[i].Name != name {
			t.Errorf("status[%d] expected %s, got %s", i, name, status.ActionStatuses[i].Name)
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestExecuteWorkflow_DAG_DuplicateDependsOn(t *testing.T) {
	exec, _ := newTestExecutor(0, nil)

	workflow := &drv1alpha1.DRWorkflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-dup", Namespace: "default"},
		Spec: drv1alpha1.DRWorkflowSpec{
			FailurePolicy: drv1alpha1.FailurePolicyFailFast,
			Actions: []drv1alpha1.Action{
				{Name: "a", Type: drv1alpha1.ActionTypeSubscription},
				{Name: "b", Type: drv1alpha1.ActionTypeSubscription, DependsOn: []string{"a", "a"}},
			},
		},
	}

	status, err := exec.ExecuteWorkflow(context.Background(), workflow, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Phase != drv1alpha1.PhaseSucceeded {
		t.Errorf("expected Succeeded, got %s: %s", status.Phase, status.Message)
	}
}
