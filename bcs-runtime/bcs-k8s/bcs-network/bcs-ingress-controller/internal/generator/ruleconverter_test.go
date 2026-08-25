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

package generator

import (
	"reflect"
	"testing"

	federationv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/federation/v1"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func protocolPtr(p k8scorev1.Protocol) *k8scorev1.Protocol { return &p }

func TestGenerateBackendsIPPortDedup(t *testing.T) {
	portName := "http"
	targetPort := int32(8080)
	hostPort := int32(30080)
	svcPort := &k8scorev1.ServicePort{
		Name:       portName,
		Protocol:   k8scorev1.ProtocolTCP,
		Port:       80,
		TargetPort: intstr.FromInt(int(targetPort)),
	}
	rc := &RuleConverter{}

	tests := []struct {
		name       string
		matchedEps []federationv1.MultiClusterEndpointSlice
		svcRoute   *networkextensionv1.ServiceRoute
		want       []networkextensionv1.ListenerBackend
	}{
		{
			name: "dedup same IP+Port across sub-clusters",
			matchedEps: []federationv1.MultiClusterEndpointSlice{
				{
					Spec: federationv1.MultiClusterEndpointSliceSpec{
						Endpoints: []federationv1.MultiClusterEndpointSliceEd{
							{
								Addresses: []string{"10.0.0.1"},
								Ports: []federationv1.EndpointPort{
									{
										Name:     &portName,
										Protocol: protocolPtr(k8scorev1.ProtocolTCP),
										Port:     &targetPort,
									},
								},
							},
						},
					},
				},
				{
					Spec: federationv1.MultiClusterEndpointSliceSpec{
						Endpoints: []federationv1.MultiClusterEndpointSliceEd{
							{
								Addresses: []string{"10.0.0.1"},
								Ports: []federationv1.EndpointPort{
									{
										Name:     &portName,
										Protocol: protocolPtr(k8scorev1.ProtocolTCP),
										Port:     &targetPort,
									},
								},
							},
							{
								Addresses: []string{"10.0.0.2"},
								Ports: []federationv1.EndpointPort{
									{
										Name:     &portName,
										Protocol: protocolPtr(k8scorev1.ProtocolTCP),
										Port:     &targetPort,
									},
								},
							},
						},
					},
				},
			},
			svcRoute: &networkextensionv1.ServiceRoute{},
			want: []networkextensionv1.ListenerBackend{
				{IP: "10.0.0.1", Port: 8080, Weight: 10},
				{IP: "10.0.0.2", Port: 8080, Weight: 10},
			},
		},
		{
			name: "dedup hostPort same NodeIP+HostPort",
			matchedEps: []federationv1.MultiClusterEndpointSlice{
				{
					Spec: federationv1.MultiClusterEndpointSliceSpec{
						Endpoints: []federationv1.MultiClusterEndpointSliceEd{
							{
								Addresses:     []string{"10.0.0.1"},
								NodeAddresses: []string{"192.168.1.1"},
								Ports: []federationv1.EndpointPort{
									{
										Name:     &portName,
										Protocol: protocolPtr(k8scorev1.ProtocolTCP),
										Port:     &targetPort,
										HostPort: &hostPort,
									},
								},
							},
						},
					},
				},
				{
					Spec: federationv1.MultiClusterEndpointSliceSpec{
						Endpoints: []federationv1.MultiClusterEndpointSliceEd{
							{
								Addresses:     []string{"10.0.0.2"},
								NodeAddresses: []string{"192.168.1.1"},
								Ports: []federationv1.EndpointPort{
									{
										Name:     &portName,
										Protocol: protocolPtr(k8scorev1.ProtocolTCP),
										Port:     &targetPort,
										HostPort: &hostPort,
									},
								},
							},
						},
					},
				},
			},
			svcRoute: &networkextensionv1.ServiceRoute{HostPort: true},
			want: []networkextensionv1.ListenerBackend{
				{IP: "192.168.1.1", Port: 30080, Weight: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rc.generateBackends(tt.matchedEps, svcPort, tt.svcRoute)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateBackends() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
