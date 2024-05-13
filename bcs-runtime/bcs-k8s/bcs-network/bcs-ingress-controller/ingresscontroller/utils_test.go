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

package ingresscontroller

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

var testIngress1 = networkextensionv1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ingress1",
		Namespace: "test",
	},
	Spec: networkextensionv1.IngressSpec{
		Rules: []networkextensionv1.IngressRule{
			{
				Port:     80,
				Protocol: "tcp",
				Services: []networkextensionv1.ServiceRoute{
					{
						ServiceName:      "tcp-test",
						ServiceNamespace: "test",
						ServicePort:      8080,
					},
				},
			},
			{
				Port:     443,
				Protocol: "http",
				Routes: []networkextensionv1.Layer7Route{
					{
						Domain: "www.test.com",
						Path:   "/",
						Services: []networkextensionv1.ServiceRoute{
							{
								ServiceName:      "http-test",
								ServiceNamespace: "test",
								ServicePort:      8080,
							},
						},
					},
				},
			},
		},
		PortMappings: []networkextensionv1.IngressPortMapping{
			{
				WorkloadKind:      "StatefulSet",
				WorkloadName:      "test-sts",
				WorkloadNamespace: "test",
			},
		},
	},
}

var testIngress2 = networkextensionv1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ingress2",
		Namespace: "test",
	},
	Spec: networkextensionv1.IngressSpec{
		Rules: []networkextensionv1.IngressRule{
			{
				Port:     80,
				Protocol: "tcp",
				Services: []networkextensionv1.ServiceRoute{
					{
						ServiceName:      "tcp-test",
						ServiceNamespace: "test",
						ServicePort:      8080,
					},
				},
			},
		},
	},
}

var testIngress3 = networkextensionv1.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "test-ingress3",
		Namespace: "test",
	},
	Spec: networkextensionv1.IngressSpec{
		Rules: []networkextensionv1.IngressRule{
			{
				Port:     80,
				Protocol: "tcp",
				Services: []networkextensionv1.ServiceRoute{
					{
						ServiceName:      "https-test",
						ServiceNamespace: "test",
						ServicePort:      8080,
					},
				},
			},
		},
	},
}

// TestDeduplicateIngress test function that deduplicate ingress
func TestDeduplicateIngress(t *testing.T) {
	before := []ingresscache.IngressMeta{
		{
			Namespace: testIngress1.Namespace,
			Name:      testIngress1.Name,
		}, {
			Namespace: testIngress1.Namespace,
			Name:      testIngress1.Name,
		}, {
			Namespace: testIngress2.Namespace,
			Name:      testIngress2.Name,
		}, {
			Namespace: testIngress2.Namespace,
			Name:      testIngress2.Name,
		}, {
			Namespace: testIngress3.Namespace,
			Name:      testIngress3.Name,
		},
	}
	after := []ingresscache.IngressMeta{
		{
			Namespace: testIngress1.Namespace,
			Name:      testIngress1.Name,
		}, {
			Namespace: testIngress2.Namespace,
			Name:      testIngress2.Name,
		}, {
			Namespace: testIngress3.Namespace,
			Name:      testIngress3.Name,
		},
	}
	results := deduplicateIngresses(before)
	if !reflect.DeepEqual(results, after) {
		t.Errorf("test failed")
	}
}
