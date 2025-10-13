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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// 测试正常GPU资源分配
func TestGetPerGPUResource_Normal(t *testing.T) {
	injector := &Injector{}

	testCases := []struct {
		resourceName    corev1.ResourceName
		pod             *corev1.Pod
		coefficients    []ResourceCoefficient
		expectedCPU     resource.Quantity
		expectedMEM     resource.Quantity
		expectedStorage resource.Quantity
		expectedMap     map[corev1.ResourceName]resource.Quantity
		expectedExtMap  map[corev1.ResourceName]ResourceCoefficient
	}{
		{
			resourceName: "nvidia.com/gpu",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "gpu-container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"nvidia.com/gpu":      resource.MustParse("2"),
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
								},
							},
						},
						{
							Name: "cpu-container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
									"others-resources":    resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			coefficients: []ResourceCoefficient{
				{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "m"},
				{Name: corev1.ResourceMemory, Coefficient: 2000, Unit: "Mi"},
				{Name: corev1.ResourceEphemeralStorage, Coefficient: 30, Unit: "Gi"},
			},
			expectedMap: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:              resource.MustParse("750m"),
				corev1.ResourceMemory:           resource.MustParse("1750Mi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("30Gi"),
			},
		},
		{
			resourceName: "tke.cloud.tencent.com/qgpu-memory",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "gpu-container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"tke.cloud.tencent.com/qgpu-memory": resource.MustParse("14"),
								},
							},
						},
					},
				},
			},
			coefficients: []ResourceCoefficient{
				{Name: "tke.cloud.tencent.com/qgpu-core", Coefficient: 33 / 14, Unit: ""},
				{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "m"},
				{Name: corev1.ResourceMemory, Coefficient: 2000, Unit: "Mi"},
				{Name: corev1.ResourceEphemeralStorage, Coefficient: 30, Unit: "Gi"},
			},
			expectedMap: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:              resource.MustParse("1000m"),
				corev1.ResourceMemory:           resource.MustParse("2000Mi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("30Gi"),
			},
			expectedExtMap: map[corev1.ResourceName]ResourceCoefficient{
				"tke.cloud.tencent.com/qgpu-core": {Name: "tke.cloud.tencent.com/qgpu-core", Coefficient: 33 / 14, Unit: ""},
			},
		},
		{
			resourceName: "tke.cloud.tencent.com/qgpu-memory",
			pod: &corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "gpu-container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"tke.cloud.tencent.com/qgpu-memory": resource.MustParse("14"),
								},
							},
						},
						{
							Name: "cpu-container",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("500Mi"),
									"others-resources":    resource.MustParse("1"),
								},
							},
						},
					},
				},
			},
			coefficients: []ResourceCoefficient{
				{Name: "tke.cloud.tencent.com/qgpu-core", Coefficient: 33 / 14, Unit: ""},
				{Name: corev1.ResourceCPU, Coefficient: 1.0 / 14, Unit: ""},
				{Name: corev1.ResourceMemory, Coefficient: 2000.0 / 14, Unit: "Mi"},
				{Name: corev1.ResourceEphemeralStorage, Coefficient: 30.0 / 14, Unit: "Gi"},
			},
			expectedMap: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceCPU:              resource.MustParse("35m"),
				corev1.ResourceMemory:           resource.MustParse("112347428571m"),
				corev1.ResourceEphemeralStorage: resource.MustParse("2300875337142m"),
			},
			expectedExtMap: map[corev1.ResourceName]ResourceCoefficient{
				"tke.cloud.tencent.com/qgpu-core": {Name: "tke.cloud.tencent.com/qgpu-core", Coefficient: 33 / 14, Unit: ""},
			},
		},
	}

	for _, testCase := range testCases {
		defaultRes, extRes, err := injector.getPerGPUResource(
			testCase.pod, testCase.resourceName, InjectInfo{ResourceList: testCase.coefficients})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for k, v := range testCase.expectedMap {
			retV := defaultRes[k]
			t.Logf("log: resouce %s expected %s, got %s", k, v.String(), retV.String())
			if !v.IsZero() && !v.Equal(defaultRes[k]) {
				t.Errorf("resouce %s expected %v, got %v", k, v, defaultRes[k])
			}
		}
		for k, v := range testCase.expectedExtMap {
			realV, ok := extRes[k]
			if !ok {
				t.Errorf("resource %s not found", k)
			}
			if !reflect.DeepEqual(realV, v) {
				t.Errorf("resource %s expected %v, got %v", k, v, realV)
			}
		}
	}

}

func Test_OtherContainerTooLarge(t *testing.T) {
	injector := &Injector{}

	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "gpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							"nvidia.com/gpu": resource.MustParse("2"),
						},
					},
				},
				{
					Name: "cpu-container",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("2500m"),
						},
					},
				},
			},
		},
	}

	coefficients := []ResourceCoefficient{
		{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "m"},
		{Name: corev1.ResourceMemory, Coefficient: 2000, Unit: "Mi"},
		{Name: corev1.ResourceEphemeralStorage, Coefficient: 30, Unit: "Gi"},
	}

	_, _, err := injector.getPerGPUResource(pod, "nvidia.com/gpu", InjectInfo{ResourceList: coefficients})
	if err == nil {
		t.Error("Expected error when other container resource too large")
	}

}

// 测试GPU数量为0的情况
func TestGetPerGPUResource_ZeroGPU(t *testing.T) {
	injector := &Injector{}

	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "non-gpu-container",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU: resource.MustParse("1"),
					},
				},
			}},
		},
	}

	_, _, err := injector.getPerGPUResource(pod, "nvidia.com/gpu", InjectInfo{ResourceList: []ResourceCoefficient{}})
	if err == nil {
		t.Error("Expected error when GPU count is zero")
	}
}

// 测试资源解析错误
func TestGetPerGPUResource_ParseError(t *testing.T) {
	injector := &Injector{}

	pod := &corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "gpu-container",
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"nvidia.com/gpu": resource.MustParse("1"),
					},
				},
			}},
		},
	}

	// 使用无效单位触发解析错误
	coefficients := []ResourceCoefficient{
		{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "invalid"},
	}

	_, _, err := injector.getPerGPUResource(pod, "nvidia.com/gpu", InjectInfo{ResourceList: coefficients})
	if err == nil {
		t.Error("Expected parse error")
	}
}

func TestGetGPUTypeAndResourceName(t *testing.T) {
	// 创建测试用的GPUInjector配置
	conf := &InjectorConfig{
		GPUResourceMap: map[string]map[corev1.ResourceName]InjectInfo{
			"L20": {
				corev1.ResourceName("nvidia.com/gpu"): InjectInfo{
					ResourceList: []ResourceCoefficient{
						{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "m"},
					},
				},
				corev1.ResourceName("amd.com/gpu"): InjectInfo{
					ResourceList: []ResourceCoefficient{
						{Name: corev1.ResourceCPU, Coefficient: 1000, Unit: "m"},
					},
				},
			},
		},
	}
	gi := &Injector{conf: conf}

	t.Run("支持的标准GPU类型", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{pluginAnnotationKey: "L20"},
				Namespace:   "test",
				Name:        "pod1",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"nvidia.com/gpu": resource.MustParse("1"),
							},
						},
					},
				},
			},
		}

		gpuType, resourceName, err := gi.getGPUTypeAndResourceName(pod)
		assert.NoError(t, err)
		assert.Equal(t, "L20", gpuType)
		assert.Equal(t, corev1.ResourceName("nvidia.com/gpu"), resourceName)
	})

	t.Run("不支持的GPU类型", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{pluginAnnotationKey: "T4"},
				Namespace:   "test",
				Name:        "pod2",
			},
		}

		gpuT, r, err := gi.getGPUTypeAndResourceName(pod)
		assert.NoError(t, err)
		assert.Equal(t, "", gpuT)
		assert.Equal(t, corev1.ResourceName(""), r)
	})

	t.Run("无GPU资源", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{pluginAnnotationKey: "L20"},
				Namespace:   "test",
				Name:        "pod3",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu": resource.MustParse("1"),
							},
						},
					},
				},
			},
		}

		gpuT, r, err := gi.getGPUTypeAndResourceName(pod)
		assert.NoError(t, err)
		assert.Equal(t, "", gpuT)
		assert.Equal(t, corev1.ResourceName(""), r)
	})

	t.Run("多个GPU资源", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{pluginAnnotationKey: "L20"},
				Namespace:   "test",
				Name:        "pod4",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"nvidia.com/gpu": resource.MustParse("1"),
								"amd.com/gpu":    resource.MustParse("1"),
							},
						},
					},
					{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"amd.com/gpu": resource.MustParse("1"),
							},
						},
					},
				},
			},
		}

		_, _, err := gi.getGPUTypeAndResourceName(pod)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pod test/pod4 has different gpu resource")
	})
}
