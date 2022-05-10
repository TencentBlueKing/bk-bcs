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

package core

import (
	"errors"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	gdutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubecontroller "k8s.io/kubernetes/pkg/controller"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// getPodFromTemplate generate a pod from template for testing case
func getPodFromTemplate(cs *v1alpha1.GameDeployment, revision, id string, index int) *corev1.Pod {
	pod, _ := kubecontroller.GetPodFromTemplate(&cs.Spec.Template, cs, metav1.NewControllerRef(cs, gdutil.ControllerKind))
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[appsv1.ControllerRevisionHashLabelKey] = revision

	pod.Name = fmt.Sprintf("%s-%s", cs.Name, id)
	pod.Namespace = cs.Namespace
	pod.Labels[v1alpha1.GameDeploymentInstanceID] = id

	if index >= 0 {
		pod.Annotations[v1alpha1.GameDeploymentIndexID] = strconv.Itoa(index)
		injectDeploymentPodIndexToEnv(pod, strconv.Itoa(index))
	}

	inplaceupdate.InjectReadinessGate(pod)
	return pod
}

func TestNewVersionedPods(t *testing.T) {
	deploy1 := test.NewGameDeployment(1)
	deploy2 := test.NewGameDeployment(1)
	tests := []struct {
		name                     string
		currentGD                *v1alpha1.GameDeployment
		updateGD                 *v1alpha1.GameDeployment
		currentRevision          string
		updateRevision           string
		expectedCreations        int
		expectedCurrentCreations int
		availableIDs             []string
		availableIndex           []int

		expectedPods []*corev1.Pod
		expectedErr  error
	}{
		{
			name:                     "less than current creation",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        1,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2"},
			availableIndex:           []int{1, 2},
			expectedPods:             []*corev1.Pod{getPodFromTemplate(deploy1, "1", "1", 1)},
			expectedErr:              nil,
		},
		{
			name:                     "equal to current creation",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        2,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2"},
			availableIndex:           []int{1, 2},
			expectedPods: []*corev1.Pod{
				getPodFromTemplate(deploy1, "1", "1", 1),
				getPodFromTemplate(deploy1, "1", "2", 2),
			},
			expectedErr: nil,
		},
		{
			name:                     "more than current creation",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        3,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2", "3"},
			availableIndex:           []int{1, 2, 3},
			expectedPods: []*corev1.Pod{
				getPodFromTemplate(deploy1, "1", "1", 1),
				getPodFromTemplate(deploy1, "1", "2", 2),
				getPodFromTemplate(deploy1, "2", "3", 3),
			},
			expectedErr: nil,
		},
		{
			name:                     "more than current creation and available ids is empty",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        3,
			expectedCurrentCreations: 2,
			availableIndex:           []int{1, 2},
			expectedErr:              nil,
		},
		{
			name:                     "more than current creation and available ids is less than required",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        3,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2"},
			availableIndex:           []int{1, 2},
			expectedPods: []*corev1.Pod{
				getPodFromTemplate(deploy1, "1", "1", 1),
				getPodFromTemplate(deploy1, "1", "2", 2),
			},
			expectedErr: nil,
		},
		{
			name:                     "more than current creation and available index is empty",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        3,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2", "3"},
			expectedPods: []*corev1.Pod{
				getPodFromTemplate(deploy1, "1", "1", -1),
				getPodFromTemplate(deploy1, "1", "2", -1),
				getPodFromTemplate(deploy1, "2", "3", -1),
			},
			expectedErr: nil,
		},
		{
			name:                     "more than current creation and available index is less than required",
			currentGD:                deploy1,
			updateGD:                 deploy2,
			currentRevision:          "1",
			updateRevision:           "2",
			expectedCreations:        3,
			expectedCurrentCreations: 2,
			availableIDs:             []string{"1", "2", "3"},
			availableIndex:           []int{1, 2},
			expectedPods: []*corev1.Pod{
				getPodFromTemplate(deploy1, "1", "1", 1),
				getPodFromTemplate(deploy1, "1", "2", 2),
				getPodFromTemplate(deploy1, "2", "3", -1),
			},
			expectedErr: nil,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			control := New(s.updateGD)
			expectedPods, err := control.NewVersionedPods(s.currentGD, s.updateGD, s.currentRevision,
				s.updateRevision, s.expectedCreations, s.expectedCurrentCreations, s.availableIDs,
				s.availableIndex)
			if err != s.expectedErr {
				t.Error("not expected error")
			}
			if !reflect.DeepEqual(expectedPods, s.expectedPods) {
				t.Errorf("expected %v, but got %v", s.expectedPods, expectedPods)
			}
		})
	}
}

func newPod(phase corev1.PodPhase, hasInPlaceUpdateCondition, inPlaceUpdateConditionTrue, available bool) *corev1.Pod {
	readyConditionStatus := corev1.ConditionFalse
	if available {
		readyConditionStatus = corev1.ConditionTrue
	}
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			Phase: phase,
			Conditions: []corev1.PodCondition{
				{
					Type:               corev1.PodReady,
					LastTransitionTime: metav1.NewTime(metav1.Now().Time.Add(-1 * time.Duration(0) * time.Second)),
					Status:             readyConditionStatus,
				},
			},
		},
	}

	// add in-place update condition
	if hasInPlaceUpdateCondition {
		con := corev1.ConditionFalse
		if inPlaceUpdateConditionTrue {
			con = corev1.ConditionTrue
		}
		pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
			Type:   inplaceupdate.InPlaceUpdateReady,
			Status: con,
		})
	}
	return pod
}

func TestIsPodUpdateReady(t *testing.T) {
	deploy := test.NewGameDeployment(1)
	tests := []struct {
		name            string
		pod             *corev1.Pod
		minReadySeconds int32
		expected        bool
	}{
		{
			name:            "pod is running but not available",
			pod:             newPod(corev1.PodRunning, false, false, false),
			minReadySeconds: 0,
			expected:        false,
		},
		{
			name:            "pod isn't running but available",
			pod:             newPod(corev1.PodPending, false, false, true),
			minReadySeconds: 0,
			expected:        false,
		},
		{
			name:            "pod is running and available, inPlaceUpdate condition is empty",
			pod:             newPod(corev1.PodRunning, false, false, true),
			minReadySeconds: 0,
			expected:        true,
		},
		{
			name:            "pod is running and available, inPlaceUpdate condition is false",
			pod:             newPod(corev1.PodRunning, true, false, true),
			minReadySeconds: 0,
			expected:        false,
		},
		{
			name:            "pod is running and available, inPlaceUpdate condition is true",
			pod:             newPod(corev1.PodRunning, true, true, true),
			minReadySeconds: 0,
			expected:        true,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			control := New(deploy)
			if got := control.IsPodUpdateReady(s.pod, s.minReadySeconds); got != s.expected {
				t.Errorf("IsPodUpdateReady() = %v, want %v", got, s.expected)
			}
		})
	}
}

func TestValidateGameDeploymentUpdate(t *testing.T) {
	deploy1 := test.NewGameDeployment(1)
	deploy1.Spec.Template.Spec.Containers[0].Image = "test1"
	deploy1.Spec.UpdateStrategy.Type = v1alpha1.RollingGameDeploymentUpdateStrategyType
	deploy2 := test.NewGameDeployment(1)
	deploy2.Spec.Template.Spec.NodeName = "test2"
	deploy2.Spec.UpdateStrategy.Type = v1alpha1.InPlaceGameDeploymentUpdateStrategyType
	deploy3 := test.NewGameDeployment(1)
	deploy3.Spec.Template.Spec.Containers[0].Image = "test3"
	deploy3.Spec.UpdateStrategy.Type = v1alpha1.InPlaceGameDeploymentUpdateStrategyType
	tests := []struct {
		name          string
		oldGD         *v1alpha1.GameDeployment
		newGD         *v1alpha1.GameDeployment
		expectedError error
	}{
		{
			name:          "not inplace update strategy",
			newGD:         deploy1,
			expectedError: nil,
		},
		{
			name:          "not patches",
			oldGD:         deploy1,
			newGD:         deploy1,
			expectedError: nil,
		},
		{
			name:          "patch operation have add",
			oldGD:         deploy1,
			newGD:         deploy2,
			expectedError: errors.New("only allowed to update images in spec for InplaceUpdate, but found add /nodeName"),
		},
		{
			name:          "path match error",
			oldGD:         deploy1,
			newGD:         deploy2,
			expectedError: errors.New("only allowed to update images in spec for InplaceUpdate, but found add /nodeName"),
		},
		{
			name:          "patches are alright",
			oldGD:         deploy1,
			newGD:         deploy3,
			expectedError: nil,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			control := New(s.newGD)
			if err := control.ValidateGameDeploymentUpdate(s.oldGD, s.newGD); !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("got: %v, want: %v", err, s.expectedError)
			}
		})
	}
}
