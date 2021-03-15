/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package randhostport

import (
	"encoding/json"
	"reflect"
	"testing"

	jsonpatch "github.com/evanphx/json-patch"
	corev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// TestInjectPod test injectPod function
func TestInjectPod(t *testing.T) {
	testCases := []struct {
		Message     string
		Pod         *corev1.Pod
		PortsList   []*PortEntry
		InjectedPod *corev1.Pod
		HasErr      bool
	}{
		{
			Message: "simple1",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app": "testname",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "8080",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				&PortEntry{
					Port:     31000,
					Quantity: 0,
				},
			},
			InjectedPod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app":                            "testname",
						podHostportLabelFlagKey:          podHostportLabelFlagValue,
						"31000" + podHostportLabelSuffix: "31000",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "8080",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								corev1.PodAffinityTerm{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
			HasErr: false,
		},
		{
			Message: "multiple ports",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app": "testname",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "8080,http",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
								},
							},
						},
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "http",
									ContainerPort: 8081,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				&PortEntry{
					Port:     31000,
					Quantity: 0,
				},
				&PortEntry{
					Port:     31001,
					Quantity: 3,
				},
				&PortEntry{
					Port:     31002,
					Quantity: 4,
				},
			},
			InjectedPod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app":                            "testname",
						podHostportLabelFlagKey:          podHostportLabelFlagValue,
						"31000" + podHostportLabelSuffix: "31000",
						"31001" + podHostportLabelSuffix: "31001",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "8080,http",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								corev1.EnvVar{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
							},
						},
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "http",
									ContainerPort: 8081,
									HostPort:      31001,
								},
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								corev1.EnvVar{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								corev1.PodAffinityTerm{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								corev1.PodAffinityTerm{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31001" + podHostportLabelSuffix: "31001",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
			HasErr: false,
		},
		{
			Message: "err container port not specify",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Annotations: map[string]string{
						pluginAnnotationKey: pluginAnnotationValue,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			HasErr: true,
		},
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)
		portCache := NewPortCache()
		for _, entry := range test.PortsList {
			portCache.PushPortEntry(entry)
		}
		hpi := &HostPortInjector{
			portCache: portCache,
		}
		patches, err := hpi.injectToPod(test.Pod)
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
		podBytes, err := json.Marshal(test.Pod)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		patchBytes, err := json.Marshal(patches)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		t.Logf("%s", string(patchBytes))
		p, err := jsonpatch.DecodePatch(patchBytes)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		podBytes, err = p.Apply(podBytes)
		if err != nil {
			t.Error(err.Error())
			continue
		}
		tmpPod := &corev1.Pod{}
		json.Unmarshal(podBytes, tmpPod)
		if !reflect.DeepEqual(tmpPod, test.InjectedPod) {
			tmpPodJSON, _ := json.Marshal(tmpPod)
			injectPodJSON, _ := json.Marshal(test.InjectedPod)
			t.Errorf("expect %s, but get %s", injectPodJSON, tmpPodJSON)
		}
	}
}
