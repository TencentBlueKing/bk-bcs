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

package canary

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"reflect"
	"testing"
	"time"
)

func TestGetCurrentCanaryStep(t *testing.T) {
	tests := []struct {
		name                     string
		deploy                   *v1alpha1.GameDeployment
		expectedCanaryStep       *v1alpha1.CanaryStep
		expectedCurrentStepIndex *int32
	}{
		{
			name:                     "canary empty",
			deploy:                   &v1alpha1.GameDeployment{},
			expectedCanaryStep:       nil,
			expectedCurrentStepIndex: nil,
		},
		{
			name: "have current step",
			deploy: &v1alpha1.GameDeployment{
				Spec: v1alpha1.GameDeploymentSpec{
					UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{
						CanaryStrategy: &v1alpha1.CanaryStrategy{
							Steps: []v1alpha1.CanaryStep{{Pause: &v1alpha1.CanaryPause{}}},
						},
					},
				},
				Status: v1alpha1.GameDeploymentStatus{
					CurrentStepIndex: func() *int32 { a := int32(1); return &a }(),
				},
			},
			expectedCanaryStep:       nil,
			expectedCurrentStepIndex: func() *int32 { a := int32(1); return &a }(),
		},
		{
			name: "steps number greater than current step index",
			deploy: &v1alpha1.GameDeployment{
				Spec: v1alpha1.GameDeploymentSpec{
					UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{
						CanaryStrategy: &v1alpha1.CanaryStrategy{
							Steps: []v1alpha1.CanaryStep{
								{Pause: &v1alpha1.CanaryPause{}},
								{Pause: &v1alpha1.CanaryPause{}},
							},
						},
					},
				},
				Status: v1alpha1.GameDeploymentStatus{
					CurrentStepIndex: func() *int32 { a := int32(1); return &a }(),
				},
			},
			expectedCanaryStep:       &v1alpha1.CanaryStep{Pause: &v1alpha1.CanaryPause{}},
			expectedCurrentStepIndex: func() *int32 { a := int32(1); return &a }(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			cs, csi := GetCurrentCanaryStep(s.deploy)
			if !reflect.DeepEqual(cs, s.expectedCanaryStep) {
				t.Errorf("GetCurrentCanaryStep() got = %v, want %v", *cs, *s.expectedCanaryStep)
			}
			if !reflect.DeepEqual(csi, s.expectedCurrentStepIndex) {
				t.Errorf("GetCurrentCanaryStep() got = %v, want %v", *csi, *s.expectedCurrentStepIndex)
			}
		})
	}
}

func TestGetCurrentPartition(t *testing.T) {
	tests := []struct {
		name                     string
		deploy                   *v1alpha1.GameDeployment
		expectedCurrentPartition int32
	}{
		{
			name:                     "currentStep is empty",
			deploy:                   &v1alpha1.GameDeployment{},
			expectedCurrentPartition: 0,
		},
		{
			name: "updateStrategy partition is specified",
			deploy: &v1alpha1.GameDeployment{Spec: v1alpha1.GameDeploymentSpec{
				UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{Partition: func() *int32 { a := int32(1); return &a }()}}},
			expectedCurrentPartition: 1,
		},
		{
			name: "step's partition is specified",
			deploy: &v1alpha1.GameDeployment{
				Spec: v1alpha1.GameDeploymentSpec{
					UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{
						CanaryStrategy: &v1alpha1.CanaryStrategy{
							Steps: []v1alpha1.CanaryStep{
								{Pause: &v1alpha1.CanaryPause{}},
								{Partition: func() *int32 { a := int32(1); return &a }()},
							},
						},
					},
				},
				Status: v1alpha1.GameDeploymentStatus{
					CurrentStepIndex: func() *int32 { a := int32(1); return &a }(),
				},
			},
			expectedCurrentPartition: 1,
		},
		{
			name: "currentStepIndex is not specified",
			deploy: &v1alpha1.GameDeployment{
				Spec: v1alpha1.GameDeploymentSpec{
					Replicas: func() *int32 { a := int32(2); return &a }(),
					UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{
						CanaryStrategy: &v1alpha1.CanaryStrategy{
							Steps: []v1alpha1.CanaryStep{
								{Pause: &v1alpha1.CanaryPause{}},
								{Partition: func() *int32 { a := int32(1); return &a }()},
							},
						},
					},
				},
			},
			expectedCurrentPartition: 2,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := GetCurrentPartition(s.deploy); got != s.expectedCurrentPartition {
				t.Errorf("GetCurrentPartition() got = %v, want %v", got, s.expectedCurrentPartition)
			}
		})
	}
}

func TestComputeStepHash(t *testing.T) {
	tests := []struct {
		name         string
		deploy       *v1alpha1.GameDeployment
		expectedHash string
	}{
		{
			name:         "canaryStrategy is nil",
			deploy:       &v1alpha1.GameDeployment{},
			expectedHash: "65bb57b6b5",
		},
		{
			name: "canaryStrategy is specified",
			deploy: &v1alpha1.GameDeployment{
				Spec: v1alpha1.GameDeploymentSpec{UpdateStrategy: v1alpha1.GameDeploymentUpdateStrategy{
					CanaryStrategy: &v1alpha1.CanaryStrategy{
						Steps: []v1alpha1.CanaryStep{{Pause: &v1alpha1.CanaryPause{}}},
					},
				}},
			},
			expectedHash: "5d9755c8cc",
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := ComputeStepHash(s.deploy); got != s.expectedHash {
				t.Errorf("expected: %s, got: %s", s.expectedHash, got)
			}
		})
	}
}

func TestGetMinDuration(t *testing.T) {
	tests := []struct {
		name             string
		duration1        time.Duration
		duration2        time.Duration
		expectedDuration time.Duration
	}{
		{
			name:             "all equal zero",
			duration1:        time.Duration(0),
			duration2:        time.Duration(0),
			expectedDuration: time.Duration(0),
		},
		{
			name:             "all less than zero, got error log",
			duration1:        time.Duration(-1),
			duration2:        time.Duration(-1),
			expectedDuration: time.Duration(-1),
		},
		{
			name:             "one duration less than zero, got error log",
			duration1:        time.Duration(-1),
			duration2:        time.Duration(0),
			expectedDuration: time.Duration(-1),
		},
		{
			name:             "one duration is zero",
			duration1:        time.Duration(0),
			duration2:        time.Duration(1),
			expectedDuration: time.Duration(1),
		},
		{
			name:             "one duration is zero 2",
			duration1:        time.Duration(1),
			duration2:        time.Duration(0),
			expectedDuration: time.Duration(1),
		},
		{
			name:             "have min duration",
			duration1:        time.Duration(1),
			duration2:        time.Duration(2),
			expectedDuration: time.Duration(1),
		},
		{
			name:             "have min duration 2",
			duration1:        time.Duration(2),
			duration2:        time.Duration(1),
			expectedDuration: time.Duration(1),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if got := GetMinDuration(s.duration1, s.duration2); got != s.expectedDuration {
				t.Errorf("expected: %v, got: %v", s.expectedDuration, got)
			}
		})
	}
}
