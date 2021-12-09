/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gamestatefulset

import (
	"errors"
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	stsfake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/clientset/internalclientset/fake"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"reflect"
	"testing"
	"time"
)

func TestCompleteCurrentCanaryStep(t *testing.T) {
	duration := int32(1)
	tests := []struct {
		name     string
		sts      *gstsv1alpha1.GameStatefulSet
		ctx      *canaryContext
		excepted bool
	}{
		{
			name: "pause step complete",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Pause: &gstsv1alpha1.CanaryPause{Duration: &duration},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []hookv1alpha1.PauseCondition{
					{
						Reason:    hookv1alpha1.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			excepted: true,
		},
		{
			name: "GameDeployment has been unpaused",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Pause: &gstsv1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []hookv1alpha1.PauseCondition{
					{
						Reason:    hookv1alpha1.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			excepted: true,
		},
		{
			name: "GameDeployment has reached the desired state",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Partition: func() *int32 { a := int32(1); return &a }(),
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdatedReadyReplicas: 2,
					ReadyReplicas:        2,
				},
			},
			excepted: true,
		},
		{
			name: "hook run complete",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Hook: &hookv1alpha1.HookStep{
								TemplateName: "hook-template",
							},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx: &canaryContext{
				currentHrs: []*hookv1alpha1.HookRun{
					func() *hookv1alpha1.HookRun {
						hr := testutil.NewHookRunFromTemplate(testutil.NewHookTemplate(), testutil.NewGameStatefulSet(3))
						hr.Labels[hook.HookRunTypeLabel] = hook.HookRunTypeCanaryStepLabel
						hr.Status.Phase = hookv1alpha1.HookPhaseSuccessful
						return hr
					}(),
				},
			},
			excepted: true,
		},
		{
			name: "hook run pause",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Hook: &hookv1alpha1.HookStep{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []hookv1alpha1.PauseCondition{
					{
						Reason:    hookv1alpha1.PauseReasonStepBasedHook,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			excepted: true,
		},
		{
			name: "canary step isn't complete",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Hook: &hookv1alpha1.HookStep{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx:      &canaryContext{},
			excepted: false,
		},
	}
	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := completeCurrentCanaryStep(s.sts, s.ctx); got != s.excepted {
				t.Errorf("completeCurrentCanaryStep() = %v, want %v", got, s.excepted)
			}
		})
	}
}

func TestCalculateConditionStatus(t *testing.T) {
	tests := []struct {
		name                       string
		pauseConditions            []hookv1alpha1.PauseCondition
		canaryCtx                  *canaryContext
		exceptedNewPauseConditions []hookv1alpha1.PauseCondition
		exceptedPaused             bool
	}{
		{
			name: "paused",
			pauseConditions: []hookv1alpha1.PauseCondition{
				{
					Reason: hookv1alpha1.PauseReasonStepBasedHook,
				},
			},
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{},
				pauseReasons: []hookv1alpha1.PauseReason{
					hookv1alpha1.PauseReasonStepBasedHook,
					hookv1alpha1.PauseReasonCanaryPauseStep,
				},
			},
			exceptedNewPauseConditions: []hookv1alpha1.PauseCondition{
				{Reason: hookv1alpha1.PauseReasonStepBasedHook},
				{Reason: hookv1alpha1.PauseReasonCanaryPauseStep},
			},
			exceptedPaused: true,
		},
		{
			name:            "not paused",
			pauseConditions: []hookv1alpha1.PauseCondition{},
			canaryCtx: &canaryContext{
				newStatus:    &gstsv1alpha1.GameStatefulSetStatus{},
				pauseReasons: []hookv1alpha1.PauseReason{},
			},
			exceptedPaused: false,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			updater := &realGameStatefulSetStatusUpdater{}
			deploy := testutil.NewGameStatefulSet(1)
			deploy.Status.PauseConditions = s.pauseConditions
			paused := updater.calculateConditionStatus(deploy, s.canaryCtx)
			if paused != s.exceptedPaused {
				t.Errorf("got: %v, excepted: %v", paused, s.exceptedPaused)
			}
			if !comparePauseConditions(s.exceptedNewPauseConditions, s.canaryCtx.newStatus.PauseConditions) {
				t.Errorf("got conditions: %v, excepted: %v", s.canaryCtx.newStatus.PauseConditions, s.exceptedNewPauseConditions)
			}
		})
	}
}

// comparePauseConditions compare pause conditions, cause of calculateConditionStatus function generate time itself,
// we need remove start time first, then compare the two conditions.
func comparePauseConditions(x, y []hookv1alpha1.PauseCondition) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		x[i].StartTime = metav1.Time{}
	}
	for i := range y {
		y[i].StartTime = metav1.Time{}
	}
	return reflect.DeepEqual(x, y)
}

func TestUpdateGameDeploymentStatus(t *testing.T) {
	duration := int32(1)
	tests := []struct {
		name            string
		sts             *gstsv1alpha1.GameStatefulSet
		canaryCtx       *canaryContext
		exceptedError   error
		exceptedActions []testing2.Action
	}{
		{
			name: "step count 0",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{},
				}
				deploy.Status.Canary.Revision = "1"
				return deploy
			}(),
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{},
			},
			exceptedError: errors.New("PatchType is not supported"),
			exceptedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
					Resource: gstsv1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "step count 0 with pause",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.Paused = true
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{},
				}
				deploy.Status.Canary.Revision = "1"
				return deploy
			}(),
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{},
			},
			exceptedError: errors.New("PatchType is not supported"),
			exceptedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
					Resource: gstsv1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "every step has executed",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.Paused = true
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Pause: &gstsv1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(1); return &a }()
				return deploy
			}(),
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{},
			},
			exceptedError: errors.New("PatchType is not supported"),
			exceptedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
					Resource: gstsv1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "complete some step",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Pause: &gstsv1alpha1.CanaryPause{Duration: &duration},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []hookv1alpha1.PauseCondition{
					{
						Reason:    hookv1alpha1.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{},
			},
			exceptedError: errors.New("PatchType is not supported"),
			exceptedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
					Resource: gstsv1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "not complete step",
			sts: func() *gstsv1alpha1.GameStatefulSet {
				deploy := testutil.NewGameStatefulSet(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
					Steps: []gstsv1alpha1.CanaryStep{
						{
							Hook:  &hookv1alpha1.HookStep{},
							Pause: &gstsv1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.UpdateRevision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			canaryCtx: &canaryContext{
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{UpdateRevision: "1"},
			},
			exceptedError: errors.New("PatchType is not supported"),
			exceptedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
					Resource: gstsv1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			gstsClient := stsfake.NewSimpleClientset(s.sts)
			updater := &realGameStatefulSetStatusUpdater{
				recorder:   &record.FakeRecorder{},
				gstsClient: gstsClient,
			}
			if err := updater.UpdateGameStatefulSetStatus(s.sts, s.canaryCtx); !reflect.DeepEqual(s.exceptedError, err) {
				t.Errorf("excepted error: %v, got error: %v", s.exceptedError, err)
			}
			if !testutil.EqualActions(s.exceptedActions, testutil.FilterPatchActionsObject(gstsClient.Actions())) {
				t.Errorf("excepted actions: %v, got actions: %v", s.exceptedActions, gstsClient.Actions())
			}
		})
	}
}
