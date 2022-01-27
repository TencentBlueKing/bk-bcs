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
 *
 */

package gamedeployment

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	deployFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/fake"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	v1alpha12 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	"github.com/stretchr/testify/assert"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
		deploy   *v1alpha1.GameDeployment
		ctx      *canaryContext
		expected bool
	}{
		{
			name: "pause step complete",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Pause: &v1alpha1.CanaryPause{Duration: &duration},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			expected: true,
		},
		{
			name: "GameDeployment has been unpaused",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Pause: &v1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			expected: true,
		},
		{
			name: "GameDeployment has reached the desired state",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Partition: func() *int32 { a := int32(1); return &a }(),
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdatedReadyReplicas: 2,
					AvailableReplicas:    2,
					ReadyReplicas:        2,
				},
			},
			expected: true,
		},
		{
			name: "hook run complete",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Hook: &v1alpha12.HookStep{
								TemplateName: "hook-template",
							},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx: &canaryContext{
				currentHrs: []*v1alpha12.HookRun{
					func() *v1alpha12.HookRun {
						hr := test.NewHookRunFromTemplate(test.NewHookTemplate(), test.NewGameDeployment(3))
						hr.Labels[hook.HookRunTypeLabel] = hook.HookRunTypeCanaryStepLabel
						hr.Status.Phase = v1alpha12.HookPhaseSuccessful
						return hr
					}(),
				},
			},
			expected: true,
		},
		{
			name: "hook run pause",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Hook: &v1alpha12.HookStep{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonStepBasedHook,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			ctx:      &canaryContext{},
			expected: true,
		},
		{
			name: "canary step isn't complete",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Hook: &v1alpha12.HookStep{},
						},
					},
				}
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			ctx:      &canaryContext{},
			expected: false,
		},
	}
	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := completeCurrentCanaryStep(s.deploy, s.ctx); got != s.expected {
				t.Errorf("completeCurrentCanaryStep() = %v, want %v", got, s.expected)
			}
		})
	}
}

func TestCalculateConditionStatus(t *testing.T) {
	tests := []struct {
		name                       string
		pauseConditions            []v1alpha12.PauseCondition
		canaryCtx                  *canaryContext
		expectedNewPauseConditions []v1alpha12.PauseCondition
		expectedPaused             bool
	}{
		{
			name: "paused",
			pauseConditions: []v1alpha12.PauseCondition{
				{
					Reason: v1alpha12.PauseReasonStepBasedHook,
				},
			},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{},
				pauseReasons: []v1alpha12.PauseReason{
					v1alpha12.PauseReasonStepBasedHook,
					v1alpha12.PauseReasonCanaryPauseStep,
				},
			},
			expectedNewPauseConditions: []v1alpha12.PauseCondition{
				{Reason: v1alpha12.PauseReasonStepBasedHook},
				{Reason: v1alpha12.PauseReasonCanaryPauseStep},
			},
			expectedPaused: true,
		},
		{
			name:            "not paused",
			pauseConditions: []v1alpha12.PauseCondition{},
			canaryCtx: &canaryContext{
				newStatus:    &v1alpha1.GameDeploymentStatus{},
				pauseReasons: []v1alpha12.PauseReason{},
			},
			expectedPaused: false,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			updater := &realGameDeploymentStatusUpdater{}
			deploy := test.NewGameDeployment(1)
			deploy.Status.PauseConditions = s.pauseConditions
			paused := updater.calculateConditionStatus(deploy, s.canaryCtx)
			if paused != s.expectedPaused {
				t.Errorf("got: %v, expected: %v", paused, s.expectedPaused)
			}
			if !comparePauseConditions(s.expectedNewPauseConditions, s.canaryCtx.newStatus.PauseConditions) {
				t.Errorf("got conditions: %v, expected: %v", s.canaryCtx.newStatus.PauseConditions, s.expectedNewPauseConditions)
			}
		})
	}
}

// comparePauseConditions compare pause conditions, cause of calculateConditionStatus function generate time itself,
// we need remove start time first, then compare the two conditions.
func comparePauseConditions(x, y []v1alpha12.PauseCondition) bool {
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

func TestCalculateBaseStatus(t *testing.T) {
	pods := []*corev1.Pod{
		{
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					apps.ControllerRevisionHashLabelKey: "1",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				Conditions: []corev1.PodCondition{
					{
						Type:   corev1.PodReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		},
	}
	canaryCtx := &canaryContext{
		newStatus: &v1alpha1.GameDeploymentStatus{UpdateRevision: "1"},
	}

	updater := &realGameDeploymentStatusUpdater{}
	deploy := test.NewGameDeployment(1)
	updater.calculateBaseStatus(deploy, canaryCtx, pods)

	if canaryCtx.newStatus.ReadyReplicas != 2 {
		t.Errorf("go ready replicas %d", canaryCtx.newStatus.ReadyReplicas)
	}
	if canaryCtx.newStatus.AvailableReplicas != 2 {
		t.Errorf("go available replicas %d", canaryCtx.newStatus.AvailableReplicas)
	}
	if canaryCtx.newStatus.UpdatedReplicas != 1 {
		t.Errorf("go updated replicas %d", canaryCtx.newStatus.UpdatedReplicas)
	}
	if canaryCtx.newStatus.UpdatedReadyReplicas != 1 {
		t.Errorf("go updated ready replicas %d", canaryCtx.newStatus.UpdatedReadyReplicas)
	}
}

func TestInconsistentStatus(t *testing.T) {
	deploy := test.NewGameDeployment(1)
	deploy.Status = v1alpha1.GameDeploymentStatus{
		ObservedGeneration:   1,
		Replicas:             1,
		ReadyReplicas:        2,
		AvailableReplicas:    2,
		UpdatedReadyReplicas: 1,
		UpdatedReplicas:      1,
		UpdateRevision:       "1",
		LabelSelector:        "foo=bar",
	}

	newStatus := v1alpha1.GameDeploymentStatus{
		ObservedGeneration:   1,
		Replicas:             1,
		ReadyReplicas:        2,
		AvailableReplicas:    2,
		UpdatedReadyReplicas: 1,
		UpdatedReplicas:      1,
		UpdateRevision:       "2",
		LabelSelector:        "foo=bar",
	}

	updater := &realGameDeploymentStatusUpdater{}
	if !updater.inconsistentStatus(deploy, &newStatus) {
		t.Errorf("consistent status")
	}
}

func TestUpdateGameDeploymentStatus(t *testing.T) {
	duration := int32(1)
	tests := []struct {
		name            string
		deploy          *v1alpha1.GameDeployment
		canaryCtx       *canaryContext
		pods            []*corev1.Pod
		expectedError   error
		expectedActions []testing2.Action
	}{
		{
			name: "step count 0",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{},
				}
				deploy.Status.Canary.Revision = "1"
				return deploy
			}(),
			pods: []*corev1.Pod{},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{},
			},
			expectedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "step count 0 with pause",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.Paused = true
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{},
				}
				deploy.Status.Canary.Revision = "1"
				return deploy
			}(),
			pods: []*corev1.Pod{},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{},
			},
			expectedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "every step has executed",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.Paused = true
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Pause: &v1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(1); return &a }()
				return deploy
			}(),
			pods: []*corev1.Pod{},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{},
			},
			expectedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "complete some step",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(1)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Pause: &v1alpha1.CanaryPause{Duration: &duration},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.Now().Add(-5 * time.Second)),
					},
				}
				return deploy
			}(),
			pods: []*corev1.Pod{},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{},
			},
			expectedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
		{
			name: "not complete step",
			deploy: func() *v1alpha1.GameDeployment {
				deploy := test.NewGameDeployment(3)
				deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{
							Hook:  &v1alpha12.HookStep{},
							Pause: &v1alpha1.CanaryPause{},
						},
					},
				}
				deploy.Status.Canary.Revision = "1"
				deploy.Status.UpdateRevision = "1"
				deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return deploy
			}(),
			pods: []*corev1.Pod{},
			canaryCtx: &canaryContext{
				newStatus: &v1alpha1.GameDeploymentStatus{UpdateRevision: "1"},
			},
			expectedActions: []testing2.Action{
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
				testing2.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
					Resource: v1alpha1.Plural}, "default", "foo", types.MergePatchType, nil),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			gdClient := deployFake.NewSimpleClientset(s.deploy)
			updater := &realGameDeploymentStatusUpdater{
				recorder: &record.FakeRecorder{},
				gdClient: gdClient,
				metrics:  gdmetrics.NewMetrics(),
			}
			err := updater.UpdateGameDeploymentStatus(s.deploy, s.canaryCtx, s.pods)
			if s.expectedError != nil {
				assert.EqualError(t, err, s.expectedError.Error())
			} else {
				assert.Equal(t, nil, err)
			}
			deployActions := test.FilterActions(gdClient.Actions(), test.FilterPatchAction)
			if !test.EqualActions(s.expectedActions, deployActions) {
				t.Errorf("expected actions:\t\n%v\ngot actions:\t\n%v", s.expectedActions, deployActions)
			}
		})
	}

}
