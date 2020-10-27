/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ingresscontroller

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
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

// TestIsServiceIngress test isServiceInIngress function
func TestIsServiceIngress(t *testing.T) {

	testSvcs := []struct {
		ServiceName      string
		ServiceNamespace string
		Found            bool
	}{
		{
			"tcp-test",
			"test",
			true,
		},
		{
			"tcp1",
			"test",
			false,
		},
		{
			"http-test",
			"test",
			true,
		},
	}
	for _, test := range testSvcs {
		result := isServiceInIngress(&testIngress1, test.ServiceName, test.ServiceNamespace)
		if result != test.Found {
			t.Errorf("test data %+v, expect %v, but get %v", test, test.Found, result)
		}
	}
}

// TestFindIngressesByService test that find ingress by service
func TestFindIngressesByService(t *testing.T) {
	testSvcs := []struct {
		ServiceName      string
		ServiceNamespace string
		Ingresses        []*networkextensionv1.Ingress
	}{
		{
			"tcp-test",
			"test",
			[]*networkextensionv1.Ingress{
				&testIngress1,
				&testIngress2,
			},
		},
		{
			"https-test",
			"test",
			[]*networkextensionv1.Ingress{
				&testIngress3,
			},
		},
	}
	for _, test := range testSvcs {
		ingresses := findIngressesByService(test.ServiceName, test.ServiceNamespace, &networkextensionv1.IngressList{
			Items: []networkextensionv1.Ingress{
				testIngress1,
				testIngress2,
				testIngress3,
			},
		})
		if !reflect.DeepEqual(ingresses, test.Ingresses) {
			t.Errorf("test data %+v, expect %+v, but get %+v", test, test.Ingresses, ingresses)
		}
	}
}

// TestFindIngressesByWorkload test that find ingress by workload info
func TestFindIngressesByWorkload(t *testing.T) {
	testWorkloads := []struct {
		WorkloadKind      string
		WorkloadName      string
		WorkloadNamespace string
		Ingresses         []*networkextensionv1.Ingress
	}{
		{
			"StatefulSet",
			"test-sts",
			"test",
			[]*networkextensionv1.Ingress{
				&testIngress1,
			},
		},
	}

	for _, wl := range testWorkloads {
		ingresses := findIngressesByWorkload(wl.WorkloadKind, wl.WorkloadName, wl.WorkloadNamespace, &networkextensionv1.IngressList{
			Items: []networkextensionv1.Ingress{
				testIngress1,
				testIngress2,
				testIngress3,
			},
		})
		if !reflect.DeepEqual(ingresses, wl.Ingresses) {
			t.Errorf("test data %+v, expect %+v, but get %+v", wl, wl.Ingresses, ingresses)
		}
	}
}

// TestDeduplicateIngress test function that deduplicate ingress
func TestDeduplicateIngress(t *testing.T) {
	before := []*networkextensionv1.Ingress{
		&testIngress1,
		&testIngress1,
		&testIngress2,
		&testIngress2,
		&testIngress3,
	}
	after := []*networkextensionv1.Ingress{
		&testIngress1,
		&testIngress2,
		&testIngress3,
	}

	results := deduplicateIngresses(before)
	if !reflect.DeepEqual(results, after) {
		t.Errorf("test failed")
	}
}
