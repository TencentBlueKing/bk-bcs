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

package kube

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	dyfake "k8s.io/client-go/dynamic/fake"
)

var (
	podGVR = schema.GroupVersionResource{
		Version:  "v1",
		Resource: "pods",
	}
	crdGVR = schema.GroupVersionResource{
		Group:    "tkex.tencent.com",
		Version:  "v1alpha1",
		Resource: "gamedeployments",
	}
)

func getFakeDyanmicClient() dynamic.Interface {
	obj := &unstructured.Unstructured{}
	obj.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      "test-pod",
			"namespace": "test-ns",
			"annotations": map[string]interface{}{
				"clb-weight": "0",
				"clb-ready":  "false",
			},
		},
	})
	obj2 := &unstructured.Unstructured{}
	obj2.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name":      "test-pod2",
			"namespace": "test-ns2",
			"annotations": map[string]interface{}{
				"clb-weight": "0",
				"clb-ready":  "false",
			},
		},
	})

	obj3 := &unstructured.Unstructured{}
	obj3.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "tkex.tencent.com/v1alpha1",
		"kind":       "GameDeployment",
		"metadata": map[string]interface{}{
			"name":      "test-gd",
			"namespace": "test-ns",
			"annotations": map[string]interface{}{
				"tkex.bkbcs.tencent.com/clb-weight": "0",
				"tkex.bkbcs.tencent.com/clb-ready":  "false",
			},
		},
	})
	obj4 := &unstructured.Unstructured{}
	obj4.SetUnstructuredContent(map[string]interface{}{
		"apiVersion": "tkex.tencent.com/v1alpha1",
		"kind":       "GameDeployment",
		"metadata": map[string]interface{}{
			"name":      "test-gd2",
			"namespace": "test-ns2",
			"annotations": map[string]interface{}{
				"tkex.bkbcs.tencent.com/clb-weight": "0",
				"tkex.bkbcs.tencent.com/clb-ready":  "false",
			},
		},
	})

	client := dyfake.NewSimpleDynamicClient(runtime.NewScheme(), obj, obj2, obj3, obj4)
	return client
}

func TestProvider_handleFunctionGet(t *testing.T) {
	type fields struct {
		dynamicClient dynamic.Interface
		cachedClient  discovery.CachedDiscoveryInterface
	}
	type args struct {
		dr     dynamic.ResourceInterface
		name   string
		metric v1alpha1.KubernetesMetric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:   "GetPodAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-pod",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.clb-weight",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.clb-ready",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: false,
		},
		{
			name:   "GetPodAnnotationErrorValue",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-pod",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.clb-weight",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.clb-ready",
							Value: "true",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
		{
			name:   "GetPodAnnotationNotExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-pod",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.xxx",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.yyy",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
		{
			name:   "GetNotExistPodAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-podxxx",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.clb-weight",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.clb-ready",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
		{
			name:   "GetCrdAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gd",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-weight",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-ready",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: false,
		},
		{
			name:   "GetCrdAnnotationErrorValue",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gd",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-weight",
							Value: "1",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-ready",
							Value: "true",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
		{
			name:   "GetCrdAnnotationNotExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gd",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/xxx",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/yyy",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
		{
			name:   "GetCrdAnnotationErrorName",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gdxxx",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-weight",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-ready",
							Value: "false",
						},
					},
					Function: "get",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				dynamicClient: tt.fields.dynamicClient,
				cachedClient:  tt.fields.cachedClient,
			}
			if err := p.handleFunction(tt.args.dr, tt.args.name, &tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Provider.handleFunction() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("test case: %s, error: %v", tt.name, err)
			}
		})
	}
}

func TestProvider_handleFunctionPatch(t *testing.T) {
	type fields struct {
		dynamicClient dynamic.Interface
		cachedClient  discovery.CachedDiscoveryInterface
	}
	type args struct {
		dr     dynamic.ResourceInterface
		name   string
		metric v1alpha1.KubernetesMetric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "PatchPodExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-pod",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.clb-weight",
							Value: "1",
						},
						{
							Path:  "metadata.annotations.clb-ready",
							Value: "true",
						},
					},
					Function: "patch",
				},
			},
			wantErr: false,
		},
		{
			name:   "PatchPodNotExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(podGVR).Namespace("test-ns"),
				name: "test-pod",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.xxx",
							Value: "1",
						},
						{
							Path:  "metadata.annotations.yyy",
							Value: "true",
						},
					},
					Function: "patch",
				},
			},
			wantErr: false,
		},
		{
			name:   "PatchCrdExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gd",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-weight",
							Value: "1",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/clb-ready",
							Value: "true",
						},
					},
					Function: "patch",
				},
			},
			wantErr: false,
		},
		{
			name:   "PatchCrdNotExistAnnotation",
			fields: fields{},
			args: args{
				dr:   getFakeDyanmicClient().Resource(crdGVR).Namespace("test-ns"),
				name: "test-gd",
				metric: v1alpha1.KubernetesMetric{
					Fields: []v1alpha1.Field{
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/xxx",
							Value: "0",
						},
						{
							Path:  "metadata.annotations.tkex.bkbcs.tencent.com/yyy",
							Value: "false",
						},
					},
					Function: "patch",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				dynamicClient: tt.fields.dynamicClient,
				cachedClient:  tt.fields.cachedClient,
			}
			if err := p.handleFunction(tt.args.dr, tt.args.name, &tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Provider.handleFunction() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("test case: %s, error: %v", tt.name, err)
			}
		})
	}
}
