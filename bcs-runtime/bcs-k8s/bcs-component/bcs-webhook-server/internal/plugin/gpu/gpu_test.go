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

package gpu

import (
	"encoding/json"
	"reflect"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

func getPatchedPod(pod *corev1.Pod, patches []types.PatchOperation) (*corev1.Pod, error) {
	podBytes, err := json.Marshal(pod)
	if err != nil {
		return nil, err
	}
	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, err
	}
	realP, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		return nil, err
	}
	realPodPatchedBytes, err := realP.Apply(podBytes)
	if err != nil {
		return nil, err
	}
	realPatchedPod := &corev1.Pod{}
	json.Unmarshal(realPodPatchedBytes, realPatchedPod)
	return realPatchedPod, nil
}

// TestDoInject test doInject function
func TestDoInject(t *testing.T) {
	testCases := []struct {
		Message         string
		Pod             *corev1.Pod
		GPUConfig       *InjectorConfig
		ExpectedPatches []types.PatchOperation
		HasErr          bool
	}{
		{
			Message: "simple GPU injection with V100",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
							Annotations: map[string]string{
								"tke.cloud.tencent.com/networks": "tke-route-eni",
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/metadata/annotations/tke.cloud.tencent.com~1networks",
					Value: "tke-route-eni",
				},
			},
			HasErr: false,
		},
		{
			Message: "no GPU annotation",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: "testns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "regular-container",
							Image: "regular-image",
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: nil,
			HasErr:          false,
		},
		{
			Message: "unsupported GPU type",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "T4",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: nil,
			HasErr:          false,
		},
		{
			Message: "multiple containers with GPU",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container-1",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
						{
							Name:  "regular-container",
							Image: "regular-image",
						},
						{
							Name:  "gpu-container-2",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/2/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/2/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/2/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/2/resources/limits/memory",
					Value: "16Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "extended resource injection",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
								{
									Name:        "nvidia.com/shm",
									Coefficient: 1,
									Unit:        "Gi",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/nvidia.com~1shm",
					Value: "1Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/nvidia.com~1shm",
					Value: "1Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "no GPU resource in containers",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "regular-container",
							Image: "regular-image",
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: nil,
			HasErr:          false,
		},
		{
			Message: "different GPU resources in containers",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container-1",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
						{
							Name:  "gpu-container-2",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"amd.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"amd.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
						"amd.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: nil,
			HasErr:          true,
		},
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)

		gi := &Injector{
			conf: test.GPUConfig,
		}

		patches, err := gi.doInject(test.Pod)

		if err == nil {
			if test.HasErr {
				t.Errorf("expect err but get no err")
				continue
			}
		} else {
			if !test.HasErr {
				t.Errorf("expect no err but get err %s", err.Error())
				continue
			}
			continue
		}

		realPatchedPod, err := getPatchedPod(test.Pod, patches)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		expectedPatchedPod, err := getPatchedPod(test.Pod, test.ExpectedPatches)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		if !reflect.DeepEqual(realPatchedPod, expectedPatchedPod) {
			t.Errorf("expect pod %v, but get %v", expectedPatchedPod, realPatchedPod)
		}
	}
}

// TestDoInjectWithExistingResources test doInject with existing resource requests
func TestDoInjectWithExistingResources(t *testing.T) {
	testCases := []struct {
		Message         string
		Pod             *corev1.Pod
		GPUConfig       *InjectorConfig
		ExpectedPatches []types.PatchOperation
		HasErr          bool
	}{
		{
			Message: "GPU injection with existing CPU and memory requests",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu":      resource.MustParse("2"),
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu":      resource.MustParse("2"),
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("2Gi"),
								},
							},
						},
						{
							Name:  "regular-container",
							Image: "regular-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "7500m", // After resource sharing calculation
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "15Gi", // After resource sharing calculation
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "7500m", // After resource sharing calculation
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "15Gi", // After resource sharing calculation
				},
			},
			HasErr: false,
		},
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)

		gi := &Injector{
			conf: test.GPUConfig,
		}

		patches, err := gi.doInject(test.Pod)

		if err == nil {
			if test.HasErr {
				t.Errorf("expect err but get no err")
				continue
			}
		} else {
			if !test.HasErr {
				t.Errorf("expect no err but get err %s", err.Error())
				continue
			}
			continue
		}

		realPatchedPod, err := getPatchedPod(test.Pod, patches)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		expectedPatchedPod, err := getPatchedPod(test.Pod, test.ExpectedPatches)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		if !reflect.DeepEqual(realPatchedPod, expectedPatchedPod) {
			realPatchedPodBytes, _ := json.Marshal(realPatchedPod)
			expectedPatchedPodBytes, _ := json.Marshal(expectedPatchedPod)
			t.Errorf("expect pod %s, but get %s", expectedPatchedPodBytes, realPatchedPodBytes)
		}
	}
}

// TestDoInjectWithSpecificRules test doInject with namespace specific rules
func TestDoInjectWithSpecificRules(t *testing.T) {
	testCases := []struct {
		Message         string
		Pod             *corev1.Pod
		GPUConfig       *InjectorConfig
		ExpectedPatches []types.PatchOperation
		HasErr          bool
	}{
		{
			Message: "namespace specific rule matched - should use namespace config",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "special-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				// 原有规则
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
							Annotations: map[string]string{
								"default.annotation": "default-value",
							},
						},
					},
				},
				// namespace 特定规则
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"special-ns": {
						"V100": {
							"nvidia.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 6000,
										Unit:        "m",
									},
									{
										Name:        corev1.ResourceMemory,
										Coefficient: 12,
										Unit:        "Gi",
									},
								},
								Annotations: map[string]string{
									"namespace.annotation": "namespace-value",
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "12",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "24Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "12",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "24Gi",
				},
				{
					Op:    "add",
					Path:  "/metadata/annotations/namespace.annotation",
					Value: "namespace-value",
				},
			},
			HasErr: false,
		},
		{
			Message: "namespace not matched - should fallback to default config",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "normal-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				// 原有规则
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
							Annotations: map[string]string{
								"default.annotation": "default-value",
							},
						},
					},
				},
				// namespace 特定规则（但 namespace 不匹配）
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"special-ns": {
						"V100": {
							"nvidia.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 6000,
										Unit:        "m",
									},
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "16Gi",
				},
				{
					Op:    "add",
					Path:  "/metadata/annotations/default.annotation",
					Value: "default-value",
				},
			},
			HasErr: false,
		},
		{
			Message: "namespace matched but GPU type not matched - should fallback to default config",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "special-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "T4",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				// 原有规则
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"T4": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
				// namespace 特定规则（但 GPU 类型不匹配）
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"special-ns": {
						"V100": {
							"nvidia.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 6000,
										Unit:        "m",
									},
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "16Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "namespace matched but resource name not matched - should fallback to default config",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "special-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				// 原有规则
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
				// namespace 特定规则（但 resource name 不匹配）
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"special-ns": {
						"V100": {
							"amd.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 6000,
										Unit:        "m",
									},
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "16Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "namespace specific rule with extended resources",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "special-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"special-ns": {
						"V100": {
							"nvidia.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 5000,
										Unit:        "m",
									},
									{
										Name:        "nvidia.com/shm",
										Coefficient: 2,
										Unit:        "Gi",
									},
								},
							},
						},
					},
				},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "5",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/nvidia.com~1shm",
					Value: "2Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "5",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/nvidia.com~1shm",
					Value: "2Gi",
				},
			},
			HasErr: false,
		},
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)

		gi := &Injector{
			conf: test.GPUConfig,
		}

		patches, err := gi.doInject(test.Pod)

		if err == nil {
			if test.HasErr {
				t.Errorf("expect err but get no err")
				continue
			}
		} else {
			if !test.HasErr {
				t.Errorf("expect no err but get err %s", err.Error())
				continue
			}
			continue
		}

		realPatchedPod, err := getPatchedPod(test.Pod, patches)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		expectedPatchedPod, err := getPatchedPod(test.Pod, test.ExpectedPatches)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		if !reflect.DeepEqual(realPatchedPod, expectedPatchedPod) {
			realPatchedPodBytes, _ := json.Marshal(realPatchedPod)
			expectedPatchedPodBytes, _ := json.Marshal(expectedPatchedPod)
			t.Errorf("expect pod %s, but get %s", expectedPatchedPodBytes, realPatchedPodBytes)
		}
	}
}

// TestDoInjectWithNamespaceWhiteList test doInject with namespace white list
func TestDoInjectWithNamespaceWhiteList(t *testing.T) {
	testCases := []struct {
		Message         string
		Pod             *corev1.Pod
		GPUConfig       *InjectorConfig
		ExpectedPatches []types.PatchOperation
		HasErr          bool
	}{
		{
			Message: "namespace in white list - should skip injection",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "whitelist-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
							Annotations: map[string]string{
								"test.annotation": "test-value",
							},
						},
					},
				},
				NamespaceWhiteList: []string{"whitelist-ns", "another-whitelist-ns"},
			},
			ExpectedPatches: nil,
			HasErr:          false,
		},
		{
			Message: "namespace not in white list - should inject normally",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "normal-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
							Annotations: map[string]string{
								"test.annotation": "test-value",
							},
						},
					},
				},
				NamespaceWhiteList: []string{"whitelist-ns", "another-whitelist-ns"},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "16Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "8",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "16Gi",
				},
				{
					Op:    "add",
					Path:  "/metadata/annotations/test.annotation",
					Value: "test-value",
				},
			},
			HasErr: false,
		},
		{
			Message: "empty white list - should inject normally",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "any-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
				NamespaceWhiteList: []string{},
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "8Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "nil white list - should inject normally",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "any-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
								{
									Name:        corev1.ResourceMemory,
									Coefficient: 8,
									Unit:        "Gi",
								},
							},
						},
					},
				},
				NamespaceWhiteList: nil,
			},
			ExpectedPatches: []types.PatchOperation{
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/requests/memory",
					Value: "8Gi",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/cpu",
					Value: "4",
				},
				{
					Op:    "replace",
					Path:  "/spec/containers/0/resources/limits/memory",
					Value: "8Gi",
				},
			},
			HasErr: false,
		},
		{
			Message: "namespace in white list with namespace specific rule - should skip injection",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "test-gpu-pod",
					Namespace: "whitelist-ns",
					Annotations: map[string]string{
						pluginAnnotationKey: "V100",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "gpu-container",
							Image: "gpu-image",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
								Limits: corev1.ResourceList{
									"nvidia.com/gpu": resource.MustParse("2"),
								},
							},
						},
					},
				},
			},
			GPUConfig: &InjectorConfig{
				ResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
					"V100": {
						"nvidia.com/gpu": InjectInfo{
							ResourceList: []ResourceCoefficient{
								{
									Name:        corev1.ResourceCPU,
									Coefficient: 4000,
									Unit:        "m",
								},
							},
						},
					},
				},
				NamespaceResourceMap: map[string]map[string]map[corev1.ResourceName]InjectInfo{
					"whitelist-ns": {
						"V100": {
							"nvidia.com/gpu": InjectInfo{
								ResourceList: []ResourceCoefficient{
									{
										Name:        corev1.ResourceCPU,
										Coefficient: 6000,
										Unit:        "m",
									},
								},
							},
						},
					},
				},
				NamespaceWhiteList: []string{"whitelist-ns"},
			},
			ExpectedPatches: nil,
			HasErr:          false,
		},
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)

		gi := &Injector{
			conf: test.GPUConfig,
		}

		patches, err := gi.doInject(test.Pod)

		if err == nil {
			if test.HasErr {
				t.Errorf("expect err but get no err")
				continue
			}
		} else {
			if !test.HasErr {
				t.Errorf("expect no err but get err %s", err.Error())
				continue
			}
			continue
		}

		if len(patches) == 0 && len(test.ExpectedPatches) == 0 {
			continue
		}

		if len(patches) == 0 && len(test.ExpectedPatches) > 0 {
			t.Errorf("expect patches but get no patches")
			continue
		}

		if len(patches) > 0 && len(test.ExpectedPatches) == 0 {
			t.Errorf("expect no patches but get patches: %v", patches)
			continue
		}

		realPatchedPod, err := getPatchedPod(test.Pod, patches)
		if err != nil {
			t.Error(err.Error())
			continue
		}

		expectedPatchedPod, err := getPatchedPod(test.Pod, test.ExpectedPatches)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		if !reflect.DeepEqual(realPatchedPod, expectedPatchedPod) {
			realPatchedPodBytes, _ := json.Marshal(realPatchedPod)
			expectedPatchedPodBytes, _ := json.Marshal(expectedPatchedPod)
			t.Errorf("expect pod %s, but get %s", expectedPatchedPodBytes, realPatchedPodBytes)
		}
	}
}
