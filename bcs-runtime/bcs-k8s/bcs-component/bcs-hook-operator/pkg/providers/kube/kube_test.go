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

package kube

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cached "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestRun(t *testing.T) {
	kubeClient := fake.NewSimpleClientset()
	kubeClient.Resources = []*metav1.APIResourceList{
		{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			GroupVersion: corev1.SchemeGroupVersion.String(),
			APIResources: []metav1.APIResource{
				{
					Name:         "pods",
					SingularName: "pod",
					Namespaced:   true,
					Kind:         "Pod",
					ShortNames:   []string{"pod"},
				},
			},
		},
	}
	cachedClient := cached.NewMemCacheClient(kubeClient.Discovery())
	p := NewKubeProvider(getFakeDyanmicClient(), cachedClient)
	hr := testutil.NewHookRun("m0")
	hr.Spec.Args = []hookv1alpha1.Argument{
		{
			Name:  "PodName",
			Value: func() *string { s := "test-pod"; return &s }(),
		},
		{
			Name:  "PodNamespace",
			Value: func() *string { s := "test-ns"; return &s }(),
		},
	}
	metric := hookv1alpha1.Metric{
		Provider: hookv1alpha1.MetricProvider{
			Kubernetes: &hookv1alpha1.KubernetesMetric{
				Fields: []hookv1alpha1.Field{
					{
						Path:  "metadata.annotations.clb-ready",
						Value: "false",
					},
				},
				Function: FunctionTypeGet,
			},
		},
	}
	mm := p.Run(hr, metric)
	assert.Equal(t, mm.Phase, hookv1alpha1.HookPhaseSuccessful)

	hr2 := testutil.NewHookRun("m0")
	hr2.Spec.Args = []hookv1alpha1.Argument{
		{
			Name:  "PodName",
			Value: func() *string { s := "test-pod2"; return &s }(),
		},
		{
			Name:  "PodNamespace",
			Value: func() *string { s := "test-ns2"; return &s }(),
		},
	}
	metric2 := hookv1alpha1.Metric{
		Provider: hookv1alpha1.MetricProvider{
			Kubernetes: &hookv1alpha1.KubernetesMetric{
				Fields: []hookv1alpha1.Field{
					{
						Path:  "metadata.annotations.clb-ready",
						Value: "true",
					},
				},
				Function: FunctionTypeGet,
			},
		},
	}
	mm2 := p.Run(hr2, metric2)
	assert.Equal(t, mm2.Phase, hookv1alpha1.HookPhaseError)

	hr3 := testutil.NewHookRun("m0")
	hr3.Spec.Args = []hookv1alpha1.Argument{
		{
			Name:  "PodName",
			Value: func() *string { s := "test-pod2"; return &s }(),
		},
		{
			Name:  "PodNamespace",
			Value: func() *string { s := "test-ns2"; return &s }(),
		},
		{
			Name:  "Group",
			Value: func() *string { s := ""; return &s }(),
		},
		{
			Name:  "Version",
			Value: func() *string { s := "v1"; return &s }(),
		},
		{
			Name:  "Kind",
			Value: func() *string { s := "Pod"; return &s }(),
		},
	}
	metric3 := hookv1alpha1.Metric{
		Provider: hookv1alpha1.MetricProvider{
			Kubernetes: &hookv1alpha1.KubernetesMetric{
				Fields: []hookv1alpha1.Field{
					{
						Path:  "metadata.annotations.clb-ready",
						Value: "true",
					},
				},
				Function: FunctionTypeGet,
			},
		},
	}
	mm3 := p.Run(hr3, metric3)
	assert.Equal(t, mm3.Phase, hookv1alpha1.HookPhaseError)
}
