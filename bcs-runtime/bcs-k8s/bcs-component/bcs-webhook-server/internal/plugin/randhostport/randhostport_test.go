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
func TestInjectPod(t *testing.T) { // nolint
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080",
						annotationsRandHostportPrefix + "8080": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
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
			Message: "simple1-container-port-rand",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app": "testname",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:               pluginAnnotationValue,
						pluginPortsAnnotationKey:          "8080",
						pluginContainerPortsAnnotationKey: pluginAnnotationValue,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                     pluginAnnotationValue,
						pluginPortsAnnotationKey:                "8080",
						pluginContainerPortsAnnotationKey:       pluginAnnotationValue,
						annotationsRandHostportPrefix + "31000": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 31000,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
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
						pluginPortsAnnotationKey: "http,8080",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
					Port:     31000,
					Quantity: 0,
				},
				{
					Port:     31001,
					Quantity: 3,
				},
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "http,8080",
						annotationsRandHostportPrefix + "8080": "31000",
						annotationsRandHostportPrefix + "8081": "31001",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
									HostPort:      31001,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
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
			Message: "multiple ports of names",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app": "testname",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "grpc,http",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 8080,
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
					Port:     31000,
					Quantity: 0,
				},
				{
					Port:     31001,
					Quantity: 3,
				},
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "grpc,http",
						annotationsRandHostportPrefix + "8080": "31000",
						annotationsRandHostportPrefix + "8081": "31001",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
									HostPort:      31001,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
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
			Message: "multiple ports with init containers",
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
					InitContainers: []corev1.Container{
						{
							Image: "test-image",
						},
					},
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
					Port:     31000,
					Quantity: 0,
				},
				{
					Port:     31001,
					Quantity: 3,
				},
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080,http",
						annotationsRandHostportPrefix + "8081": "31000",
						annotationsRandHostportPrefix + "8080": "31001",
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "test-image",
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31001,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			HasErr: true,
		},
		{
			Message: "Affinity == nil",
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
					Affinity: nil,
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080",
						annotationsRandHostportPrefix + "8080": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
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
			Message: "PodAntiAffinity == nil",
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: nil,
					},
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080",
						annotationsRandHostportPrefix + "8080": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									}),
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
			Message: "execution != nil",
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(map[string]string{
										"31001" + podHostportLabelSuffix: "31001",
									}),
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080",
						annotationsRandHostportPrefix + "8080": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(map[string]string{
										"31001" + podHostportLabelSuffix: "31001",
									}),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									}),
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
			Message: "execution == nil",
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
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: nil,
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
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
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080",
						annotationsRandHostportPrefix + "8080": "31000",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31000",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									}),
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
			Message: "multiple ports with sidecar init containers",
			Pod: &corev1.Pod{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "testname",
					Namespace: "testns",
					Labels: map[string]string{
						"app": "testname",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:      pluginAnnotationValue,
						pluginPortsAnnotationKey: "8080,http,8088",
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "test-image",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8088,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
						},
					},
				},
			},
			PortsList: []*PortEntry{
				{
					Port:     31000,
					Quantity: 0,
				},
				{
					Port:     31001,
					Quantity: 3,
				},
				{
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
						"31002" + podHostportLabelSuffix: "31002",
					},
					Annotations: map[string]string{
						pluginAnnotationKey:                    pluginAnnotationValue,
						pluginPortsAnnotationKey:               "8080,http,8088",
						annotationsRandHostportPrefix + "8081": "31000",
						annotationsRandHostportPrefix + "8080": "31001",
						annotationsRandHostportPrefix + "8088": "31002",
					},
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "test-image",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8088,
									HostPort:      31002,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name:  envRandHostportPrefix + "8088",
									Value: "31002",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8081,
									HostPort:      31000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name:  envRandHostportPrefix + "8088",
									Value: "31002",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
						{
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 8080,
									HostPort:      31001,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  envRandHostportPrefix + "8081",
									Value: "31000",
								},
								{
									Name:  envRandHostportPrefix + "8080",
									Value: "31001",
								},
								{
									Name:  envRandHostportPrefix + "8088",
									Value: "31002",
								},
								{
									Name: envRandHostportHostIP,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.hostIP",
										},
									},
								},
								{
									Name: envRandHostportPodName,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name: envRandHostportPodNamespace,
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
							},
						},
					},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31000" + podHostportLabelSuffix: "31000",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31001" + podHostportLabelSuffix: "31001",
									})),
									TopologyKey: "kubernetes.io/hostname",
								},
								{
									LabelSelector: k8smetav1.SetAsLabelSelector(labels.Set(map[string]string{
										"31002" + podHostportLabelSuffix: "31002",
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
	}

	for testIndex, test := range testCases {
		t.Logf("test index %d, message %s", testIndex, test.Message)
		portCache := NewPortCache()
		for _, entry := range test.PortsList {
			portCache.PushPortEntry(entry)
		}
		hpi := &HostPortInjector{
			portCache: portCache,
			conf:      &HostPortInjectorConfig{},
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
