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

package filterclb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestHandle(t *testing.T) {
	var testCases = []struct {
		name                string
		kind                metav1.GroupVersionKind
		ingress             runtime.Object
		supportIngressClass bool
		objectData          []runtime.Object
		expectedAllowed     bool
	}{
		{
			name: "test lb service",
			kind: serviceKind,
			ingress: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			expectedAllowed: false,
		},
		{
			name: "test lb service with subnetid",
			kind: serviceKind,
			ingress: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						annotationServiceInternalSubnetID: "subnet-xxxx",
					},
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeLoadBalancer,
				},
			},
			expectedAllowed: true,
		},
		{
			name: "test v1 ingress",
			kind: ingressV1Kind,
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			supportIngressClass: true,
			expectedAllowed:     false,
		},
		{
			name: "test multi ingress controller",
			kind: ingressV1Kind,
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			supportIngressClass: true,
			objectData: []runtime.Object{
				&networkingv1.IngressClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nginx",
						Annotations: map[string]string{
							annotationIngressClassDefaultKey: "true",
						},
					},
				},
				&networkingv1.IngressClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "qcloud",
					},
				},
			},
			expectedAllowed: true,
		},
		{
			name: "test v1 ingress with subnetid",
			kind: ingressV1Kind,
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						annotationIngressSubnetID: "subnet-xxxx",
					},
				},
			},
			expectedAllowed: true,
		},
		{
			name: "test v1beta1 ingress with subnetid",
			kind: ingressV1Kind,
			ingress: &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						annotationIngressSubnetID: "subnet-xxxx",
					},
				},
			},
			supportIngressClass: false,
			expectedAllowed:     true,
		},
		{
			name: "test v1beta1 ingress with l7-lb-controller",
			kind: ingressV1Kind,
			ingress: &networkingv1beta1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			},
			supportIngressClass: false,
			objectData: []runtime.Object{
				&extensionsv1beta1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      l7LbControllerName,
						Namespace: "kube-system",
					},
				},
			},
			expectedAllowed: false,
		},
		{
			name: "test qcloud ingress with nginx ingress class",
			kind: ingressV1Kind,
			ingress: &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						annotationsIngressClass: "nginx",
					},
				},
			},
			expectedAllowed: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			objectRaw, err := json.Marshal(testCase.ingress)
			assert.Nil(t, err)
			plugin := &Handler{}
			review := v1beta1.AdmissionReview{
				Request: &v1beta1.AdmissionRequest{
					Kind:      testCase.kind,
					Operation: v1beta1.Create,
					Object: runtime.RawExtension{
						Raw:    objectRaw,
						Object: testCase.ingress,
					},
				},
			}
			//err = plugin.Init("")
			gvkArrayString := make([]string, 0, 4)
			gvkArrayString = append(gvkArrayString, serviceKind.String(), ingressV1Kind.String(),
				ingressExtensionsV1Kind.String(), ingressV1Beta1Kind.String())
			plugin.gvkArrayString = gvkArrayString
			plugin.kubeClient = fake.NewSimpleClientset(testCase.objectData...)
			plugin.supportIngressClass = testCase.supportIngressClass
			assert.Nil(t, err)
			response := plugin.Handle(review)
			if response.Result != nil {
				t.Log(response.Result.Message)
			}
			assert.Equal(t, response.Allowed, testCase.expectedAllowed)
		})
	}
}
