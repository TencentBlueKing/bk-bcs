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

package generator

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	k8sfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud/mock"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
)

func getNowTimeStamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// get fake existed listeners
func getExistedListeners() []networkextensionv1.Listener {
	return []networkextensionv1.Listener{
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      GetListenerName("lb1", 8000),
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           8000,
				EndPort:        0,
				Protocol:       "tcp",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      GetListenerName("lb1", 8001),
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           8001,
				EndPort:        0,
				Protocol:       "tcp",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      GetSegmentListenerName("lb1", 3000, 3002),
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           3000,
				EndPort:        3002,
				Protocol:       "tcp",
			},
		},
		{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      GetSegmentListenerName("lb1", 3003, 3005),
				Namespace: "ns1",
			},
			Spec: networkextensionv1.ListenerSpec{
				LoadbalancerID: "lb1",
				Port:           3003,
				EndPort:        3005,
				Protocol:       "tcp",
			},
		},
	}
}

// construct fake statefulset data
func constructStatefulsetData(cli k8sclient.Client) {
	podIPs := []string{
		"127.0.1.1",
		"127.0.1.2",
		"127.0.1.3",
		"127.0.1.4",
	}
	hostIPs := []string{
		"192.168.1.1", // nolint
		"192.168.1.2", // nolint
		"192.168.1.3", // nolint
		"192.168.1.4", // nolint
	}
	containerPort := 8080
	for i := 0; i < 4; i++ {
		cli.Create(context.TODO(), &k8scorev1.Pod{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "sts-" + strconv.Itoa(i),
				Namespace: "test",
				Labels: map[string]string{
					"app": "sts-1",
				},
			},
			Spec: k8scorev1.PodSpec{
				Containers: []k8scorev1.Container{
					{
						Ports: []k8scorev1.ContainerPort{
							{
								ContainerPort: int32(containerPort),
							},
						},
					},
				},
			},
			Status: k8scorev1.PodStatus{
				PodIP:  podIPs[i],
				HostIP: hostIPs[i],
			},
		})
	}
	cli.Create(context.TODO(), &k8sappsv1.StatefulSet{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      "sts-1",
			Namespace: "test",
		},
		Spec: k8sappsv1.StatefulSetSpec{
			Selector: k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
				"app": "sts-1",
			})),
		},
	})
}

// construct fake k8s data
func constructK8sData(cli k8sclient.Client) {
	podIPs := []string{
		"127.0.0.1", // nolint
		"127.0.0.2", // nolint
		"127.0.0.3", // nolint
		"127.0.0.4", // nolint
	}
	hostIPs := []string{
		"192.168.0.1", // nolint
		"192.168.0.2", // nolint
		"192.168.0.3", // nolint
		"192.168.0.4", // nolint
	}
	containerPort := 8080
	labelAppValue := []string{
		"app1",
		"app1",
		"app2",
		"app2",
	}
	labelExtraKey := []string{
		"k1",
		"k2",
		"k3",
		"k4",
	}
	labelExtraValue := []string{
		"v1",
		"v2",
		"v3",
		"v4",
	}

	for i := 0; i < 4; i++ {
		cli.Create(context.TODO(), &k8scorev1.Pod{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      "app-" + strconv.Itoa(i),
				Namespace: "test",
				Labels: map[string]string{
					"app":            labelAppValue[i],
					labelExtraKey[i]: labelExtraValue[i],
				},
			},
			Spec: k8scorev1.PodSpec{
				Containers: []k8scorev1.Container{
					{
						Ports: []k8scorev1.ContainerPort{
							{
								ContainerPort: int32(containerPort),
							},
						},
					},
				},
			},
			Status: k8scorev1.PodStatus{
				PodIP:  podIPs[i],
				HostIP: hostIPs[i],
			},
		})
	}

	svcNames := []string{
		"svc1",
		"svc2",
	}
	svcPorts := []int32{
		9000,
		9001,
	}
	nodePorts := []int32{
		30021,
		30022,
	}
	selectorValue := []string{
		"app1",
		"app2",
	}
	for i := 0; i < 2; i++ {
		cli.Create(context.TODO(), &k8scorev1.Service{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      svcNames[i],
				Namespace: "test",
			},
			Spec: k8scorev1.ServiceSpec{
				Selector: map[string]string{
					"app": selectorValue[i],
				},
				Ports: []k8scorev1.ServicePort{
					{
						Protocol:   "tcp",
						Port:       svcPorts[i],
						NodePort:   nodePorts[i],
						TargetPort: intstr.FromInt(containerPort),
					},
				},
			},
		})
		cli.Create(context.TODO(), &k8scorev1.Endpoints{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name:      svcNames[i],
				Namespace: "test",
			},
			Subsets: []k8scorev1.EndpointSubset{
				{
					Addresses: []k8scorev1.EndpointAddress{
						{
							IP: podIPs[i*2],
						},
						{
							IP: podIPs[i*2+1],
						},
					},
					Ports: []k8scorev1.EndpointPort{
						{
							Port: int32(containerPort),
						},
					},
				},
			},
		})
	}
}

// TestIngressConvert test converting ingress function
func TestIngressConvert(t *testing.T) {

	testCases := []struct {
		testTitle          string
		ingresses          []networkextensionv1.Ingress
		generatedListeners map[string]networkextensionv1.Listener
		isTCPUDPReuse      bool
		hasErr             bool
	}{
		{
			testTitle: "layer 4 listener to nodeport",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						Rules: []networkextensionv1.IngressRule{
							{
								Port:     8000,
								Protocol: "TCP",
								Services: []networkextensionv1.ServiceRoute{
									{
										ServiceName:      "svc1",
										ServiceNamespace: "test",
										ServicePort:      9000,
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetListenerName("lb1", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerName("lb1", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "192.168.0.1", // nolint
									Port:   30021,
									Weight: 10,
								},
								{
									IP:     "192.168.0.2", // nolint
									Port:   30021,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "layer 4 listener to pod",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						Rules: []networkextensionv1.IngressRule{
							{
								Port:     8000,
								Protocol: "TCP",
								Services: []networkextensionv1.ServiceRoute{
									{
										ServiceName:      "svc1",
										ServiceNamespace: "test",
										ServicePort:      9000,
										IsDirectConnect:  true,
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetListenerName("lb1", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerName("lb1", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.0.1",
									Port:   8080,
									Weight: 10,
								},
								{
									IP:     "127.0.0.2",
									Port:   8080,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "layer 7 listener to nodeport",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						Rules: []networkextensionv1.IngressRule{
							{
								Port:     8000,
								Protocol: "HTTP",
								Routes: []networkextensionv1.Layer7Route{
									{
										Domain: "www.qq.com",
										Path:   "/",
										Services: []networkextensionv1.ServiceRoute{
											{
												ServiceName:      "svc1",
												ServiceNamespace: "test",
												ServicePort:      9000,
											},
										},
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetListenerName("lb1", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerName("lb1", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "HTTP",
						Rules: []networkextensionv1.ListenerRule{
							{
								Domain: "www.qq.com",
								Path:   "/",
								TargetGroup: &networkextensionv1.ListenerTargetGroup{
									TargetGroupProtocol: "HTTP",
									Backends: []networkextensionv1.ListenerBackend{
										{
											IP:     "192.168.0.1", // nolint
											Port:   30021,
											Weight: 10,
										},
										{
											IP:     "192.168.0.2", // nolint
											Port:   30021,
											Weight: 10,
										},
									},
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "layer 4 to service subset",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						Rules: []networkextensionv1.IngressRule{
							{
								Port:     8000,
								Protocol: "TCP",
								Services: []networkextensionv1.ServiceRoute{
									{
										ServiceName:      "svc1",
										ServiceNamespace: "test",
										ServicePort:      9000,
										IsDirectConnect:  true,
										Subsets: []networkextensionv1.IngressSubset{
											{
												LabelSelector: map[string]string{
													"k1": "v1",
												},
												Weight: &networkextensionv1.IngressWeight{
													Value: 100,
												},
											},
											{
												LabelSelector: map[string]string{
													"k2": "v2",
												},
												Weight: &networkextensionv1.IngressWeight{
													Value: 2,
												},
											},
										},
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetListenerName("lb1", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerName("lb1", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.0.1",
									Port:   8080,
									Weight: 100,
								},
								{
									IP:     "127.0.0.2",
									Port:   8080,
									Weight: 2,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "mapping test 1",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						PortMappings: []networkextensionv1.IngressPortMapping{
							{
								WorkloadKind:      "StatefulSet",
								WorkloadName:      "sts-1",
								WorkloadNamespace: "test",
								StartPort:         18000,
								StartIndex:        1,
								EndIndex:          4,
								Protocol:          "TCP",
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetSegmentListenerName("lb1", 18001, 0): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18001, 0),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18001,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.2",
									Port:   18001,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
				GetSegmentListenerName("lb1", 18002, 0): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18002, 0),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18002,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.3",
									Port:   18002,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "mapping test 2",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						PortMappings: []networkextensionv1.IngressPortMapping{
							{
								WorkloadKind:      "StatefulSet",
								WorkloadName:      "sts-1",
								WorkloadNamespace: "test",
								StartPort:         18000,
								StartIndex:        1,
								SegmentLength:     10,
								EndIndex:          4,
								Protocol:          "TCP",
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetSegmentListenerName("lb1", 18010, 18019): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18010, 18019),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18010,
						EndPort:        18019,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.2",
									Port:   18010,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
				GetSegmentListenerName("lb1", 18020, 18029): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18020, 18029),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18020,
						EndPort:        18029,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.3",
									Port:   18020,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "mapping test 3",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						PortMappings: []networkextensionv1.IngressPortMapping{
							{
								WorkloadKind:      "StatefulSet",
								WorkloadName:      "sts-1",
								WorkloadNamespace: "test",
								StartPort:         18000,
								StartIndex:        1,
								RsStartPort:       28000,
								SegmentLength:     10,
								EndIndex:          4,
								Protocol:          "TCP",
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetSegmentListenerName("lb1", 18010, 18019): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18010, 18019),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18010,
						EndPort:        18019,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.2",
									Port:   28010,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
				GetSegmentListenerName("lb1", 18020, 18029): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18020, 18029),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueTrue,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18020,
						EndPort:        18029,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "127.0.1.3",
									Port:   28020,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "mapping test for http",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "test-ingress-for-http-mapping",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						PortMappings: []networkextensionv1.IngressPortMapping{
							{
								WorkloadKind:      "StatefulSet",
								WorkloadName:      "sts-1",
								WorkloadNamespace: "test",
								StartPort:         18000,
								StartIndex:        1,
								EndIndex:          4,
								Protocol:          "HTTP",
								Routes: []networkextensionv1.IngressPortMappingLayer7Route{
									networkextensionv1.IngressPortMappingLayer7Route{
										Domain: "www.testdomain1.com",
										Path:   "/url1",
									},
									networkextensionv1.IngressPortMappingLayer7Route{
										Domain: "www.testdomain2.com",
										Path:   "/url2",
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetSegmentListenerName("lb1", 18001, 0): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18001, 0),
						Namespace: "test",
						Labels: map[string]string{
							"test-ingress-for-http-mapping": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18001,
						Protocol:       "HTTP",
						Rules: []networkextensionv1.ListenerRule{
							networkextensionv1.ListenerRule{
								Domain: "www.testdomain1.com",
								Path:   "/url1",
								TargetGroup: &networkextensionv1.ListenerTargetGroup{
									TargetGroupProtocol: "HTTP",
									Backends: []networkextensionv1.ListenerBackend{
										{
											IP:     "127.0.1.2",
											Port:   18001,
											Weight: 10,
										},
									},
								},
							},
							networkextensionv1.ListenerRule{
								Domain: "www.testdomain2.com",
								Path:   "/url2",
								TargetGroup: &networkextensionv1.ListenerTargetGroup{
									TargetGroupProtocol: "HTTP",
									Backends: []networkextensionv1.ListenerBackend{
										{
											IP:     "127.0.1.2",
											Port:   18001,
											Weight: 10,
										},
									},
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
				GetSegmentListenerName("lb1", 18002, 0): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetSegmentListenerName("lb1", 18002, 0),
						Namespace: "test",
						Labels: map[string]string{
							"test-ingress-for-http-mapping": networkextensionv1.LabelValueForIngressName,
							// if segment length is 1, don't use segment feature
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           18002,
						Protocol:       "HTTP",
						Rules: []networkextensionv1.ListenerRule{
							networkextensionv1.ListenerRule{
								Domain: "www.testdomain1.com",
								Path:   "/url1",
								TargetGroup: &networkextensionv1.ListenerTargetGroup{
									TargetGroupProtocol: "HTTP",
									Backends: []networkextensionv1.ListenerBackend{
										{
											IP:     "127.0.1.3",
											Port:   18002,
											Weight: 10,
										},
									},
								},
							},
							networkextensionv1.ListenerRule{
								Domain: "www.testdomain2.com",
								Path:   "/url2",
								TargetGroup: &networkextensionv1.ListenerTargetGroup{
									TargetGroupProtocol: "HTTP",
									Backends: []networkextensionv1.ListenerBackend{
										{
											IP:     "127.0.1.3",
											Port:   18002,
											Weight: 10,
										},
									},
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			hasErr: false,
		},
		{
			testTitle: "rule for layer 4 listener to nodeport, with tcp udp port reuse",
			ingresses: []networkextensionv1.Ingress{
				{
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "ingress1",
						Namespace: "test",
						Annotations: map[string]string{
							networkextensionv1.AnnotationKeyForLoadbalanceIDs: "lb1",
						},
					},
					Spec: networkextensionv1.IngressSpec{
						Rules: []networkextensionv1.IngressRule{
							{
								Port:     8000,
								Protocol: "TCP",
								Services: []networkextensionv1.ServiceRoute{
									{
										ServiceName:      "svc1",
										ServiceNamespace: "test",
										ServicePort:      9000,
									},
								},
							},
							{
								Port:     8000,
								Protocol: "UDP",
								Services: []networkextensionv1.ServiceRoute{
									{
										ServiceName:      "svc1",
										ServiceNamespace: "test",
										ServicePort:      9000,
									},
								},
							},
						},
					},
					Status: networkextensionv1.IngressStatus{},
				},
			},
			generatedListeners: map[string]networkextensionv1.Listener{
				GetListenerNameWithProtocol("lb1", "tcp", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerNameWithProtocol("lb1", "tcp", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "TCP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "TCP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "192.168.0.1", // nolint
									Port:   30021,
									Weight: 10,
								},
								{
									IP:     "192.168.0.2", // nolint
									Port:   30021,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
				GetListenerNameWithProtocol("lb1", "udp", 8000): {
					TypeMeta: k8smetav1.TypeMeta{},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      GetListenerNameWithProtocol("lb1", "udp", 8000),
						Namespace: "test",
						Labels: map[string]string{
							"ingress1": networkextensionv1.LabelValueForIngressName,
							networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
							networkextensionv1.LabelKeyForLoadbalanceID:     "lb1",
							networkextensionv1.LabelKeyForLoadbalanceRegion: "testregion",
						},
						ResourceVersion: "1",
						Finalizers:      []string{"ingresscontroller.bkbcs.tencent.com"},
					},
					Spec: networkextensionv1.ListenerSpec{
						LoadbalancerID: "lb1",
						Port:           8000,
						Protocol:       "UDP",
						TargetGroup: &networkextensionv1.ListenerTargetGroup{
							TargetGroupProtocol: "UDP",
							Backends: []networkextensionv1.ListenerBackend{
								{
									IP:     "192.168.0.1", // nolint
									Port:   30021,
									Weight: 10,
								},
								{
									IP:     "192.168.0.2", // nolint
									Port:   30021,
									Weight: 10,
								},
							},
						},
					},
					Status: networkextensionv1.ListenerStatus{},
				},
			},
			isTCPUDPReuse: true,
			hasErr:        false,
		},
	}

	ctrl := gomock.NewController(t)
	mockCloud := mock.NewMockLoadBalance(ctrl)
	mockCloud.EXPECT().
		IsNamespaced().
		Return(false).
		AnyTimes()
	mockCloud.
		EXPECT().
		DescribeLoadBalancer("testregion", "lb1", "").
		Return(&cloud.LoadBalanceObject{
			LbID:   "lb1",
			Name:   "lbname1",
			Region: "testregion",
		}, nil).
		AnyTimes()
	mockCloud.
		EXPECT().
		DescribeLoadBalancer("testregion", "lb2", "").
		Return(&cloud.LoadBalanceObject{
			LbID:   "lb2",
			Name:   "lbname2",
			Region: "testregion",
		}, nil).
		AnyTimes()

	for _, test := range testCases {
		t.Logf("test content: %s", test.testTitle)
		newScheme := runtime.NewScheme()
		existedListeners := &networkextensionv1.ListenerList{}
		newScheme.AddKnownTypes(
			networkextensionv1.GroupVersion,
			&networkextensionv1.Listener{},
			&networkextensionv1.Ingress{},
			existedListeners)
		newScheme.AddKnownTypes(
			k8scorev1.SchemeGroupVersion,
			&k8scorev1.Service{},
			&k8scorev1.ServiceList{},
			&k8scorev1.Endpoints{},
			&k8scorev1.Pod{},
			&k8scorev1.PodList{})
		newScheme.AddKnownTypes(
			k8sappsv1.SchemeGroupVersion,
			&k8sappsv1.StatefulSet{},
		)
		cli := k8sfake.NewFakeClientWithScheme(newScheme)

		constructK8sData(cli)
		constructStatefulsetData(cli)

		ic, err := NewIngressConverter(
			&IngressConverterOpt{
				DefaultRegion:     "testregion",
				IsTCPUDPPortReuse: test.isTCPUDPReuse,
			},
			cli,
			tencentcloud.NewClbValidater(),
			mockCloud,
		)
		if err != nil {
			t.Errorf("create new ingress converter failed, err %s", err.Error())
		}

		for _, ingress := range test.ingresses {
			cli.Create(context.TODO(), &ingress)
			err := ic.ProcessUpdateIngress(&ingress)
			if (err != nil && !test.hasErr) || (err == nil && test.hasErr) {
				t.Errorf("expect %v, but err is %v", test.hasErr, err)
			}
		}

		tmpListenerMap := make(map[string]networkextensionv1.Listener)
		cli.List(context.TODO(), existedListeners, &k8sclient.ListOptions{})
		for index := range existedListeners.Items {
			tmpListenerMap[existedListeners.Items[index].GetName()] = existedListeners.Items[index]
		}
		for key, lis := range test.generatedListeners {
			tmpLis := tmpListenerMap[key]
			if !reflect.DeepEqual(lis, tmpLis) {
				lisData, _ := json.Marshal(lis)
				tmpData, _ := json.Marshal(tmpLis)
				t.Errorf("expected %s but get %s", string(lisData), string(tmpData))
			}
		}
	}
}
