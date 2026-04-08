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

package webhook

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestValidateExecutionParams(t *testing.T) {
	tests := []struct {
		name      string
		params    []drv1alpha1.Parameter
		wantErrs  int
		wantMatch string
	}{
		{
			name:     "empty params is valid",
			params:   nil,
			wantErrs: 0,
		},
		{
			name: "valid static value param",
			params: []drv1alpha1.Parameter{
				{Name: "foo", Value: "bar"},
			},
			wantErrs: 0,
		},
		{
			name: "valid valueFrom manifestRef",
			params: []drv1alpha1.Parameter{
				{
					Name: "jobName",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							Kind:       "Job",
							Namespace:  "default",
							JSONPath:   "{.metadata.name}",
						},
					},
				},
			},
			wantErrs: 0,
		},
		{
			name: "empty name is invalid",
			params: []drv1alpha1.Parameter{
				{Name: "", Value: "bar"},
			},
			wantErrs:  1,
			wantMatch: "must not be empty",
		},
		{
			name: "duplicate name is invalid",
			params: []drv1alpha1.Parameter{
				{Name: "foo", Value: "a"},
				{Name: "foo", Value: "b"},
			},
			wantErrs:  1,
			wantMatch: "duplicate",
		},
		{
			name: "value and valueFrom mutually exclusive",
			params: []drv1alpha1.Parameter{
				{
					Name:  "foo",
					Value: "bar",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							Kind:       "Job",
							JSONPath:   "{.metadata.name}",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "mutually exclusive",
		},
		{
			name: "manifestRef apiVersion required",
			params: []drv1alpha1.Parameter{
				{
					Name: "foo",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							Kind:     "Job",
							JSONPath: "{.metadata.name}",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "apiVersion",
		},
		{
			name: "manifestRef kind required",
			params: []drv1alpha1.Parameter{
				{
					Name: "foo",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							JSONPath:   "{.metadata.name}",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "kind",
		},
		{
			name: "manifestRef jsonPath required",
			params: []drv1alpha1.Parameter{
				{
					Name: "foo",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							Kind:       "Job",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "jsonPath",
		},
		{
			name: "name and labelSelector mutually exclusive",
			params: []drv1alpha1.Parameter{
				{
					Name: "foo",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion:    "batch/v1",
							Kind:          "Job",
							JSONPath:      "{.metadata.name}",
							Name:          "my-job",
							LabelSelector: "app=foo",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "mutually exclusive",
		},
		{
			name: "name starting with digit is invalid",
			params: []drv1alpha1.Parameter{
				{Name: "1foo", Value: "bar"},
			},
			wantErrs:  1,
			wantMatch: "must match",
		},
		{
			name: "name with invalid chars is invalid",
			params: []drv1alpha1.Parameter{
				{Name: "foo.bar", Value: "baz"},
			},
			wantErrs:  1,
			wantMatch: "must match",
		},
		{
			name: "name with underscore and dash is valid",
			params: []drv1alpha1.Parameter{
				{Name: "my_param-1", Value: "val"},
			},
			wantErrs: 0,
		},
		{
			name: "reserved name mode is rejected",
			params: []drv1alpha1.Parameter{
				{Name: "mode", Value: "upgrade"},
			},
			wantErrs:  1,
			wantMatch: "reserved",
		},
		{
			name: "invalid select value",
			params: []drv1alpha1.Parameter{
				{
					Name: "foo",
					ValueFrom: &drv1alpha1.ParameterValueFrom{
						ManifestRef: &drv1alpha1.ManifestRef{
							APIVersion: "batch/v1",
							Kind:       "Job",
							JSONPath:   "{.metadata.name}",
							Select:     "Invalid",
						},
					},
				},
			},
			wantErrs:  1,
			wantMatch: "select",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateExecutionParams(tt.params)
			if len(errs) != tt.wantErrs {
				t.Errorf("got %d errors, want %d: %v", len(errs), tt.wantErrs, errs)
				return
			}
			if tt.wantMatch != "" && len(errs) > 0 {
				found := false
				for _, e := range errs {
					if contains(e, tt.wantMatch) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tt.wantMatch, errs)
				}
			}
		})
	}
}

func TestValidateExecutionSpecParams(t *testing.T) {
	// Tests that DRPlanExecutionSpec.Params is validated end-to-end via validateExecutionParams
	params := []drv1alpha1.Parameter{
		{Name: "releaseRevision", Value: "5"},
		{
			Name: "jobName",
			ValueFrom: &drv1alpha1.ParameterValueFrom{
				ManifestRef: &drv1alpha1.ManifestRef{
					APIVersion: "batch/v1",
					Kind:       "Job",
					JSONPath:   "{.metadata.name}",
					Select:     "Last",
				},
			},
		},
	}
	errs := validateExecutionParams(params)
	if len(errs) != 0 {
		t.Errorf("expected no errors for valid spec params, got: %v", errs)
	}
}

// TestValidateRevertExecutionPhase verifies that a Revert execution is allowed
// when the referenced Execute execution is in a terminal phase (Succeeded or Failed),
// and rejected for non-terminal phases (Running, Pending, etc.).
func TestValidateRevertExecutionPhase(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := drv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}

	makePlan := func(phase string) *drv1alpha1.DRPlan {
		return &drv1alpha1.DRPlan{
			ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
			Status:     drv1alpha1.DRPlanStatus{Phase: phase},
		}
	}

	makeExecRef := func(name, phase string) *drv1alpha1.DRPlanExecution {
		return &drv1alpha1.DRPlanExecution{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
			Spec: drv1alpha1.DRPlanExecutionSpec{
				PlanRef:       "test-plan",
				OperationType: drv1alpha1.OperationTypeExecute,
			},
			Status: drv1alpha1.DRPlanExecutionStatus{Phase: phase},
		}
	}

	tests := []struct {
		name          string
		refPhase      string
		wantErrSubstr string // empty means expect no error about phase
	}{
		{
			name:     "revert allowed when referenced execution is Succeeded",
			refPhase: drv1alpha1.PhaseSucceeded,
		},
		{
			name:     "revert allowed when referenced execution is Failed",
			refPhase: drv1alpha1.PhaseFailed,
		},
		{
			name:          "revert rejected when referenced execution is Running",
			refPhase:      drv1alpha1.PhaseRunning,
			wantErrSubstr: "terminal phase",
		},
		{
			name:          "revert rejected when referenced execution is Pending",
			refPhase:      drv1alpha1.PhasePending,
			wantErrSubstr: "terminal phase",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refExec := makeExecRef("ref-exec", tt.refPhase)
			plan := makePlan(drv1alpha1.PlanPhaseExecuted)

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(plan, refExec).
				WithStatusSubresource(refExec).
				Build()

			w := &DRPlanExecutionWebhook{Client: fakeClient}

			revertExec := &drv1alpha1.DRPlanExecution{
				ObjectMeta: metav1.ObjectMeta{Name: "revert-exec", Namespace: "default"},
				Spec: drv1alpha1.DRPlanExecutionSpec{
					PlanRef:            "test-plan",
					OperationType:      drv1alpha1.OperationTypeRevert,
					RevertExecutionRef: "ref-exec",
				},
			}

			errs := w.validateExecution(context.Background(), revertExec)

			phaseErr := ""
			for _, e := range errs {
				if contains(e, "terminal phase") {
					phaseErr = e
					break
				}
			}

			if tt.wantErrSubstr == "" {
				if phaseErr != "" {
					t.Errorf("expected no phase error, got: %s", phaseErr)
				}
			} else {
				if phaseErr == "" {
					t.Errorf("expected error containing %q, got no phase error (all errors: %v)", tt.wantErrSubstr, errs)
				}
			}
		})
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestValidateExecutionAllowsExecuteWhenPlanExecuted(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := drv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Status:     drv1alpha1.DRPlanStatus{Phase: drv1alpha1.PlanPhaseExecuted},
	}

	execution := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "upgrade-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			Mode:          "Upgrade",
			OperationType: drv1alpha1.OperationTypeExecute,
		},
	}

	w := &DRPlanExecutionWebhook{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(plan).Build(),
	}

	errs := w.validateExecution(context.Background(), execution)
	for _, err := range errs {
		if contains(err, "not ready") {
			t.Fatalf("expected Executed plan to allow Execute, got errors: %v", errs)
		}
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestValidateExecutionAllowsDeleteMode(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := drv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add scheme: %v", err)
	}

	plan := &drv1alpha1.DRPlan{
		ObjectMeta: metav1.ObjectMeta{Name: "test-plan", Namespace: "default"},
		Status:     drv1alpha1.DRPlanStatus{Phase: drv1alpha1.PlanPhaseReady},
	}

	execution := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "delete-exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "test-plan",
			Mode:          "Delete",
			OperationType: drv1alpha1.OperationTypeExecute,
		},
	}

	w := &DRPlanExecutionWebhook{
		Client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(plan).Build(),
	}

	if errs := w.validateExecution(context.Background(), execution); len(errs) != 0 {
		t.Fatalf("expected delete mode execute to be allowed, got errors: %v", errs)
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestValidateUpdateRejectsAnySpecMutation(t *testing.T) {
	w := &DRPlanExecutionWebhook{}

	oldObj := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Install",
			Params: []drv1alpha1.Parameter{
				{Name: "revision", Value: "1"},
			},
		},
	}
	newObj := oldObj.DeepCopy()
	newObj.Spec.Mode = "Upgrade"

	if _, err := w.ValidateUpdate(context.Background(), oldObj, newObj); err == nil {
		t.Fatal("expected spec mutation to be rejected")
	}
}

// NOCC:tosa/fn_length(设计如此)
func TestValidateUpdateAllowsMetadataOrStatusOnlyChanges(t *testing.T) {
	w := &DRPlanExecutionWebhook{}

	oldObj := &drv1alpha1.DRPlanExecution{
		ObjectMeta: metav1.ObjectMeta{Name: "exec", Namespace: "default"},
		Spec: drv1alpha1.DRPlanExecutionSpec{
			PlanRef:       "demo",
			OperationType: drv1alpha1.OperationTypeExecute,
			Mode:          "Install",
		},
	}
	newObj := oldObj.DeepCopy()
	newObj.Labels = map[string]string{"k": "v"}
	newObj.Status.Phase = drv1alpha1.PhaseRunning

	if _, err := w.ValidateUpdate(context.Background(), oldObj, newObj); err != nil {
		t.Fatalf("expected metadata/status-only update to be allowed, got: %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
